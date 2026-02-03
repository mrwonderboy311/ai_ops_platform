// Package ssh provides common SSH client functionality
package ssh

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClient wraps a standard SSH client
type SSHClient struct {
	client *ssh.Client
}

// SSHConfig holds SSH connection configuration
type SSHConfig struct {
	HostID     string
	IPAddress  string
	Port       int
	Username   string
	Password   string
	PrivateKey []byte
	Timeout    time.Duration
}

// CommandOutput represents the output of a command execution
type CommandOutput struct {
	Stdout    string
	Stderr    string
	ExitCode  *int32
	ExitError error
}

// NewClient creates a new SSH client
func NewClient(config *SSHConfig) (*SSHClient, error) {
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

	return &SSHClient{
		client: client,
	}, nil
}

// GetClient returns the underlying SSH client
func (c *SSHClient) GetClient() *ssh.Client {
	return c.client
}

// Close closes the SSH client
func (c *SSHClient) Close() error {
	return c.client.Close()
}

// ExecuteCommand executes a command on the remote host
func (c *SSHClient) ExecuteCommand(cmd string, timeout time.Duration) (*CommandOutput, error) {
	// Create a new session
	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Run command with timeout if specified
	var exitErr error
	if timeout > 0 {
		// Use channel to implement timeout
		type result struct {
			err error
		}
		resultChan := make(chan result, 1)

		go func() {
			resultChan <- result{err: session.Run(cmd)}
		}()

		select {
		case res := <-resultChan:
			exitErr = res.err
		case <-time.After(timeout):
			session.Close()
			exitErr = fmt.Errorf("command timed out after %v", timeout)
		}
	} else {
		exitErr = session.Run(cmd)
	}

	// Get exit code
	exitCode := int32(0)
	if exitErr != nil {
		// Try to get exit code from exit error
		if exitSig, ok := exitErr.(*ssh.ExitError); ok {
			exitCode = int32(exitSig.ExitStatus())
		} else {
			exitCode = -1
		}
	}

	return &CommandOutput{
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		ExitCode:  &exitCode,
		ExitError: exitErr,
	}, nil
}
