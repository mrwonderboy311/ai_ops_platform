// Package ssh provides SSH proxy functionality for remote terminal access
package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Session represents an active SSH session
type Session struct {
	ID           string
	HostID       string
	SSHClient    *ssh.Client
	SSHSession   *ssh.Session
	StdinPipe    io.WriteCloser
	StdoutPipe   io.Reader
	StderrPipe   io.Reader
	Ctx          context.Context
	CancelFunc   context.CancelFunc
	CreatedAt    time.Time
	LastActivity time.Time
	Rows         uint16
	Cols         uint16
}

// SSHProxy manages SSH connections
type SSHProxy struct {
	sessions      map[string]*Session
	sessionsMutex sync.RWMutex
	logger        io.Writer
}

// NewSSHProxy creates a new SSH proxy
func NewSSHProxy(logger io.Writer) *SSHProxy {
	return &SSHProxy{
		sessions: make(map[string]*Session),
		logger:   logger,
	}
}

// ConnectConfig holds SSH connection configuration
type ConnectConfig struct {
	HostID       string
	IPAddress    string
	Port         int
	Username     string
	Password     string
	PrivateKey   []byte
	Timeout      time.Duration
	InitialRows  uint16
	InitialCols  uint16
}

// Connect establishes an SSH connection and returns a session
func (p *SSHProxy) Connect(ctx context.Context, config *ConnectConfig) (*Session, error) {
	// Create SSH client config
	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            config.Username,
		Timeout:         config.Timeout,
	}

	// Add authentication method
	if config.Password != "" {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(config.Password),
		}
	}
	if len(config.PrivateKey) > 0 {
		signer, err := ssh.ParsePrivateKey(config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	}

	// Establish TCP connection
	address := fmt.Sprintf("%s:%d", config.IPAddress, config.Port)
	conn, err := net.DialTimeout("tcp", address, config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	// Create SSH connection
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, sshConfig)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("SSH handshake failed: %w", err)
	}

	client := ssh.NewClient(sshConn, chans, reqs)

	// Create session
	sshSession, err := client.NewSession()
	if err != nil {
		client.Close()
		sshConn.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}

	// Set up pseudo-terminal (convert uint16 to int)
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,      // enable echo
		ssh.TTY_OP_ISPEED: 115200, // input speed = 115200 baud
		ssh.TTY_OP_OSPEED: 115200, // output speed = 115200 baud
	}

	if err := sshSession.RequestPty("xterm-256color", int(config.InitialRows), int(config.InitialCols), modes); err != nil {
		sshSession.Close()
		client.Close()
		sshConn.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to request pseudo-terminal: %w", err)
	}

	// Get pipes for stdin/stdout/stderr
	stdinPipe, err := sshSession.StdinPipe()
	if err != nil {
		sshSession.Close()
		client.Close()
		sshConn.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdoutPipe, err := sshSession.StdoutPipe()
	if err != nil {
		sshSession.Close()
		client.Close()
		sshConn.Close()
		conn.Close()
		stdinPipe.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderrPipe, err := sshSession.StderrPipe()
	if err != nil {
		sshSession.Close()
		client.Close()
		sshConn.Close()
		conn.Close()
		stdinPipe.Close()
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Create session context with cancel
	sessionCtx, cancel := context.WithCancel(ctx)

	sessionID := fmt.Sprintf("%s-%d", config.HostID, time.Now().UnixNano())

	session := &Session{
		ID:           sessionID,
		HostID:       config.HostID,
		SSHClient:    client,
		SSHSession:   sshSession,
		StdinPipe:    stdinPipe,
		StdoutPipe:   stdoutPipe,
		StderrPipe:   stderrPipe,
		Ctx:          sessionCtx,
		CancelFunc:   cancel,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Rows:         config.InitialRows,
		Cols:         config.InitialCols,
	}

	// Register session
	p.sessionsMutex.Lock()
	p.sessions[sessionID] = session
	p.sessionsMutex.Unlock()

	p.log("SSH session created: %s for host %s", sessionID, config.HostID)

	return session, nil
}

// Write writes data to the SSH session stdin
func (s *Session) Write(data []byte) (int, error) {
	s.LastActivity = time.Now()
	return s.StdinPipe.Write(data)
}

// Read reads data from SSH session stdout with timeout
func (s *Session) Read(timeout time.Duration) ([]byte, error) {
	buf := make([]byte, 4096)

	// Use channel to implement timeout
	result := make(chan []byte)
	errChan := make(chan error, 1)

	go func() {
		n, err := s.StdoutPipe.Read(buf)
		if err != nil {
			errChan <- err
			return
		}
		result <- buf[:n]
	}()

	select {
	case data := <-result:
		s.LastActivity = time.Now()
		return data, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(timeout):
		return nil, nil // Timeout is not an error for non-blocking read
	case <-s.Ctx.Done():
		return nil, io.EOF
	}
}

// ResizeTerminal resizes the pseudo-terminal
func (s *Session) ResizeTerminal(rows, cols uint16) error {
	s.Rows = rows
	s.Cols = cols
	s.LastActivity = time.Now()

	// Send window change signal (convert to int as expected by WindowChange)
	err := s.SSHSession.WindowChange(int(rows), int(cols))
	if err != nil {
		return fmt.Errorf("failed to send window-change request: %w", err)
	}

	return nil
}

// Close closes the SSH session
func (s *Session) Close() error {
	s.CancelFunc()

	// Close stdin pipe
	s.StdinPipe.Close()

	// Close SSH session and client
	s.SSHSession.Close()
	s.SSHClient.Close()

	return nil
}

// GetSession retrieves a session by ID
func (p *SSHProxy) GetSession(sessionID string) *Session {
	p.sessionsMutex.RLock()
	defer p.sessionsMutex.RUnlock()
	return p.sessions[sessionID]
}

// CloseSession closes and removes a session
func (p *SSHProxy) CloseSession(sessionID string) error {
	p.sessionsMutex.Lock()
	defer p.sessionsMutex.Unlock()

	session, exists := p.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	delete(p.sessions, sessionID)

	p.log("Closing SSH session: %s", sessionID)
	return session.Close()
}

// GetSessionByHostID gets the active session for a host
func (p *SSHProxy) GetSessionByHostID(hostID string) *Session {
	p.sessionsMutex.RLock()
	defer p.sessionsMutex.RUnlock()

	for _, session := range p.sessions {
		if session.HostID == hostID {
			return session
		}
	}
	return nil
}

// CloseHostSessions closes all sessions for a host
func (p *SSHProxy) CloseHostSessions(hostID string) {
	p.sessionsMutex.Lock()
	defer p.sessionsMutex.Unlock()

	for sessionID, session := range p.sessions {
		if session.HostID == hostID {
			session.Close()
			delete(p.sessions, sessionID)
			p.log("Closed SSH session: %s for host: %s", sessionID, hostID)
		}
	}
}

// CleanupStaleSessions removes sessions that have been inactive for too long
func (p *SSHProxy) CleanupStaleSessions(timeout time.Duration) {
	p.sessionsMutex.Lock()
	defer p.sessionsMutex.Unlock()

	now := time.Now()
	for sessionID, session := range p.sessions {
		if now.Sub(session.LastActivity) > timeout {
			session.Close()
			delete(p.sessions, sessionID)
			p.log("Cleaned up stale SSH session: %s", sessionID)
		}
	}
}

// log writes a log message
func (p *SSHProxy) log(format string, args ...interface{}) {
	if p.logger != nil {
		msg := fmt.Sprintf("[SSH Proxy] "+format, args...)
		p.logger.Write([]byte(msg + "\n"))
	}
}
