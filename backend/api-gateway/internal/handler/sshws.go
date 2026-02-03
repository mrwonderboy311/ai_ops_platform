// Package handler provides WebSocket handlers for SSH terminal
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/wangjialin/myops/pkg/model"
	"github.com/wangjialin/myops/pkg/ssh"
	"gorm.io/gorm"
)

// SSHWebSocketHandler handles SSH WebSocket connections
type SSHWebSocketHandler struct {
	db           *gorm.DB
	proxy         *ssh.SSHProxy
	upgrader      websocket.Upgrader
	pingInterval  time.Duration
	pingWait      time.Duration
	writeWait     time.Duration
}

// NewSSHWebSocketHandler creates a new SSH WebSocket handler
func NewSSHWebSocketHandler(db *gorm.DB, logger io.Writer) *SSHWebSocketHandler {
	return &SSHWebSocketHandler{
		db:      db,
		proxy:   ssh.NewSSHProxy(logger),
		upgrader: websocket.Upgrader{
			ReadBufferSize:   8192,
			WriteBufferSize: 8192,
			CheckOrigin: func(r *http.Request) bool {
				return true // TODO: Configure proper origin checking
			},
		},
		pingInterval: 30 * time.Second,
		pingWait:     10 * time.Second,
		writeWait:    10 * time.Second,
	}
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type string      `json:"type"` // "input", "resize", "ping", "pong"
	Data string      `json:"data"`
	Rows uint16       `json:"rows,omitempty"`
	Cols uint16       `json:"cols,omitempty"`
}

// SSHConnectRequest represents a connection request
type SSHConnectRequest struct {
	HostID    string `json:"hostId"`
	Username  string `json:"username"`
	Password  string `json:"password,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
	Rows      uint16  `json:"rows,omitempty"`
	Cols      uint16  `json:"cols,omitempty"`
}

// ServeHTTP handles WebSocket upgrade requests
func (h *SSHWebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract host ID from path: /ws/ssh/{hostId}
	path := r.URL.Path
	prefix := "/ws/ssh/"
	if len(path) < len(prefix) {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Host ID is required")
		return
	}

	hostID := path[len(prefix):]
	if hostID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Host ID is required")
		return
	}

	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Verify host exists and user has permission
	var host model.Host
	err := h.db.Where("id = ?", hostID).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Check if host is approved/online
	if host.Status != model.HostStatusApproved && host.Status != model.HostStatusOnline {
		respondWithError(w, http.StatusForbidden, "HOST_NOT_AVAILABLE", "Host is not available for connection")
		return
	}

	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	sessionID := uuid.New().String()
	h.log("WebSocket connected: session=%s, host=%s, user=%s", sessionID, hostID, userID)

	// Wait for connection message with credentials
	connectMsg, err := h.readWSMessage(conn, time.Second*10)
	if err != nil {
		h.writeWSError(conn, "Failed to read connect message", err)
		return
	}

	var connectReq SSHConnectRequest
	if err := json.Unmarshal([]byte(connectMsg), &connectReq); err != nil {
		h.writeWSError(conn, "Invalid connect request", err)
		return
	}

	if connectReq.HostID != hostID {
		h.writeWSError(conn, "Host ID mismatch", fmt.Errorf("expected host ID %s, got %s", hostID, connectReq.HostID))
		return
	}

	// Set defaults
	if connectReq.Username == "" {
		connectReq.Username = "root"
	}
	if connectReq.Rows == 0 {
		connectReq.Rows = 24
	}
	if connectReq.Cols == 0 {
		connectReq.Cols = 80
	}

	// Connect to SSH
	connectConfig := &ssh.ConnectConfig{
		HostID:      hostID,
		IPAddress:   host.IPAddress,
		Port:        host.Port,
		Username:    connectReq.Username,
		Password:    connectReq.Password,
		PrivateKey:  []byte(connectReq.PrivateKey),
		Timeout:     30 * time.Second,
		InitialRows: connectReq.Rows,
		InitialCols: connectReq.Cols,
	}

	sessionCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session, err := h.proxy.Connect(sessionCtx, connectConfig)
	if err != nil {
		h.writeWSError(conn, "SSH connection failed", err)
		return
	}
	defer session.Close()

	// Send success message
	h.writeWSMessage(conn, map[string]interface{}{
		"type":    "connected",
		"sessionId": session.ID,
	})

	// Start SSH session (shell)
	if err := session.SSHSession.Shell(); err != nil {
		h.writeWSError(conn, "Failed to start shell", err)
		return
	}

	// Start goroutines for bidirectional communication
	var wg sync.WaitGroup
	wg.Add(3)

	// Read from SSH and write to WebSocket
	done := make(chan struct{})

	// Stdout reader
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				data, err := session.Read(50 * time.Millisecond)
				if err != nil {
					if err != io.EOF {
						h.log("Error reading from SSH: session=%s, error=%v", session.ID, err)
					}
					return
				}
				if len(data) > 0 {
					if err := h.writeWSMessage(conn, map[string]interface{}{
						"type": "output",
						"data": string(data),
					}); err != nil {
						h.log("Error writing to WebSocket: session=%s, error=%v", session.ID, err)
						return
					}
				}
			}
		}
	}()

	// WebSocket reader
	go func() {
		defer wg.Done()
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err) && !websocket.IsUnexpectedCloseError(err, messageType) {
					h.log("WebSocket read error: session=%s, error=%v", session.ID, err)
				}
				return
			}

			switch messageType {
			case websocket.TextMessage:
				var msg WSMessage
				if err := json.Unmarshal(data, &msg); err != nil {
					h.log("Invalid WebSocket message: session=%s, error=%v", session.ID, err)
					continue
				}

				switch msg.Type {
				case "input":
					if _, err := session.Write([]byte(msg.Data)); err != nil {
						h.log("Error writing to SSH: session=%s, error=%v", session.ID, err)
						return
					}
				case "resize":
					if err := session.ResizeTerminal(msg.Rows, msg.Cols); err != nil {
						h.log("Error resizing terminal: session=%s, error=%v", session.ID, err)
					}
				case "ping":
					h.writeWSMessage(conn, map[string]interface{}{
						"type": "pong",
					})
				case "pong":
					// Handle pong
				}
			case websocket.BinaryMessage:
				// Not supported yet
			}
		}
	}()

	// Ping handler
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(h.pingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					h.log("WebSocket ping failed: session=%s", session.ID)
					return
				}
			}
		}
	}()

	// Wait for disconnect
	<-session.Ctx.Done()
	close(done)

	// Wait for goroutines to finish
	wg.Wait()

	h.log("WebSocket disconnected: session=%s", sessionID)

	// Clean up session
	h.proxy.CloseSession(session.ID)
}

// readWSMessage reads a message from WebSocket with timeout
func (h *SSHWebSocketHandler) readWSMessage(conn *websocket.Conn, timeout time.Duration) (string, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, data, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", errors.New("no data received")
	}
	return string(data), nil
}

// writeWSMessage writes a message to WebSocket
func (h *SSHWebSocketHandler) writeWSMessage(conn *websocket.Conn, msg interface{}) error {
	conn.SetWriteDeadline(time.Now().Add(h.writeWait))
	return conn.WriteJSON(msg)
}

// writeWSError writes an error message to WebSocket and closes the connection
func (h *SSHWebSocketHandler) writeWSError(conn *websocket.Conn, message string, err error) {
	h.writeWSMessage(conn, map[string]interface{}{
		"type":    "error",
		"message": message,
		"error":   err.Error(),
	})
	conn.Close()
}

// log writes a log message
func (h *SSHWebSocketHandler) log(format string, args ...interface{}) {
	msg := fmt.Sprintf("[SSH WS] "+format, args...)
	fmt.Printf("%s\n", msg) // TODO: Use proper logger
}
