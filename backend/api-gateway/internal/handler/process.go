// Package handler provides HTTP handlers for process management operations
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"github.com/wangjialin/myops/pkg/ssh"
	"gorm.io/gorm"
)

// ProcessManagementHandler handles process management operations
type ProcessManagementHandler struct {
	db *gorm.DB
}

// NewProcessManagementHandler creates a new process management handler
func NewProcessManagementHandler(db *gorm.DB) *ProcessManagementHandler {
	return &ProcessManagementHandler{db: db}
}

// ListProcesses handles process list requests
func (h *ProcessManagementHandler) ListProcesses(w http.ResponseWriter, r *http.Request) {
	var req model.ListProcessesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
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

	// Verify host exists
	var host model.Host
	err := h.db.Where("id = ?", req.HostID).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Check if host is available
	if host.Status != model.HostStatusApproved && host.Status != model.HostStatusOnline {
		respondWithError(w, http.StatusForbidden, "HOST_NOT_AVAILABLE", "Host is not available")
		return
	}

	// Create SSH client
	config := &ssh.SSHConfig{
		HostID:     req.HostID.String(),
		IPAddress:  host.IPAddress,
		Port:       host.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: []byte(req.Key),
		Timeout:    30 * time.Second,
	}

	client, err := ssh.NewProcessClient(config)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CONNECTION_FAILED", fmt.Sprintf("Failed to connect: %v", err))
		return
	}
	defer client.Close()

	// List processes
	processes, err := client.ListProcesses()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "LIST_FAILED", fmt.Sprintf("Failed to list processes: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": model.ListProcessesResponse{
			Processes: processes,
			Count:     len(processes),
		},
	})
}

// GetProcess handles get process details requests
func (h *ProcessManagementHandler) GetProcess(w http.ResponseWriter, r *http.Request) {
	var req model.GetProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
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

	// Verify host exists
	var host model.Host
	err := h.db.Where("id = ?", req.HostID).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Check if host is available
	if host.Status != model.HostStatusApproved && host.Status != model.HostStatusOnline {
		respondWithError(w, http.StatusForbidden, "HOST_NOT_AVAILABLE", "Host is not available")
		return
	}

	// Create SSH client
	config := &ssh.SSHConfig{
		HostID:     req.HostID.String(),
		IPAddress:  host.IPAddress,
		Port:       host.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: []byte(req.Key),
		Timeout:    30 * time.Second,
	}

	client, err := ssh.NewProcessClient(config)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CONNECTION_FAILED", fmt.Sprintf("Failed to connect: %v", err))
		return
	}
	defer client.Close()

	// Get process
	process, err := client.GetProcess(req.PID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "PROCESS_NOT_FOUND", fmt.Sprintf("Failed to get process: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": process,
	})
}

// KillProcess handles kill process requests
func (h *ProcessManagementHandler) KillProcess(w http.ResponseWriter, r *http.Request) {
	var req model.KillProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
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

	// Verify host exists
	var host model.Host
	err := h.db.Where("id = ?", req.HostID).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Check if host is available
	if host.Status != model.HostStatusApproved && host.Status != model.HostStatusOnline {
		respondWithError(w, http.StatusForbidden, "HOST_NOT_AVAILABLE", "Host is not available")
		return
	}

	// Create SSH client
	config := &ssh.SSHConfig{
		HostID:     req.HostID.String(),
		IPAddress:  host.IPAddress,
		Port:       host.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: []byte(req.Key),
		Timeout:    30 * time.Second,
	}

	client, err := ssh.NewProcessClient(config)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CONNECTION_FAILED", fmt.Sprintf("Failed to connect: %v", err))
		return
	}
	defer client.Close()

	// Kill process
	signal := req.Signal
	if signal == 0 {
		signal = 9 // Default to SIGKILL
	}

	if err := client.KillProcess(req.PID, signal); err != nil {
		respondWithError(w, http.StatusInternalServerError, "KILL_FAILED", fmt.Sprintf("Failed to kill process: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("Process %d terminated with signal %d", req.PID, signal),
	})
}

// ExecuteCommand handles command execution requests
func (h *ProcessManagementHandler) ExecuteCommand(w http.ResponseWriter, r *http.Request) {
	var req model.ExecuteCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
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

	// Verify host exists
	var host model.Host
	err := h.db.Where("id = ?", req.HostID).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Check if host is available
	if host.Status != model.HostStatusApproved && host.Status != model.HostStatusOnline {
		respondWithError(w, http.StatusForbidden, "HOST_NOT_AVAILABLE", "Host is not available")
		return
	}

	// Validate command
	if req.Command == "" {
		respondWithError(w, http.StatusBadRequest, "EMPTY_COMMAND", "Command cannot be empty")
		return
	}

	// Set default timeout
	timeout := 60 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	// Create execution record
	executionID := uuid.New()
	now := time.Now()
	execution := &model.ProcessExecution{
		ID:        executionID,
		HostID:    req.HostID,
		UserID:    userID,
		Command:   req.Command,
		Status:    "running",
		StartedAt: &now,
	}

	if err := h.db.Create(execution).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create execution record")
		return
	}

	// Create SSH client
	config := &ssh.SSHConfig{
		HostID:     req.HostID.String(),
		IPAddress:  host.IPAddress,
		Port:       host.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: []byte(req.Key),
		Timeout:    30 * time.Second,
	}

	client, err := ssh.NewProcessClient(config)
	if err != nil {
		h.updateExecutionStatus(executionID, "failed", nil, "", err.Error(), 0)
		respondWithError(w, http.StatusInternalServerError, "CONNECTION_FAILED", fmt.Sprintf("Failed to connect: %v", err))
		return
	}
	defer client.Close()

	// Execute command
	response, err := client.ExecuteCommand(req.Command, timeout, req.WorkingDir)

	// Parse duration
	duration := int64(0)
	if response != nil {
		if d, err := time.ParseDuration(response.Duration); err == nil {
			duration = d.Milliseconds()
		}
	}

	if err != nil {
		status := "failed"
		exitCode := int32(-1)
		stdout := ""
		stderr := ""
		if response != nil {
			stdout = response.Stdout
			stderr = response.Stderr
			if response.ExitCode != 0 {
				exitCode = response.ExitCode
			}
		}
		h.updateExecutionStatus(executionID, status, &exitCode, stdout, stderr, duration)
		respondWithError(w, http.StatusInternalServerError, "EXECUTION_FAILED", fmt.Sprintf("Command execution failed: %v", err))
		return
	}

	// Update execution record as completed
	h.updateExecutionStatus(executionID, "completed", &response.ExitCode, response.Stdout, response.Stderr, duration)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"executionId": executionID,
			"response":    response,
		},
	})
}

// GetExecutions retrieves command execution history for a user
func (h *ProcessManagementHandler) GetExecutions(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	hostIDStr := r.URL.Query().Get("hostId")
	status := r.URL.Query().Get("status")

	query := h.db.Model(&model.ProcessExecution{}).Where("user_id = ?", userID)

	if hostIDStr != "" {
		hostID, err := uuid.Parse(hostIDStr)
		if err == nil {
			query = query.Where("host_id = ?", hostID)
		}
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get host ID from path (for /api/v1/hosts/{hostId}/executions)
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) >= 4 && pathParts[3] == "hosts" && len(pathParts) >= 6 && pathParts[5] == "executions" {
		if hostID, err := uuid.Parse(pathParts[4]); err == nil {
			query = query.Where("host_id = ?", hostID)
		}
	}

	var executions []model.ProcessExecution
	if err := query.Order("created_at DESC").Limit(100).Find(&executions).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve executions")
		return
	}

	// Eager load host and user
	for i := range executions {
		h.db.Preload("Host").Preload("User").First(&executions[i], executions[i].ID)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": executions,
	})
}

// Helper functions

func (h *ProcessManagementHandler) updateExecutionStatus(executionID uuid.UUID, status string, exitCode *int32, stdout, stderr string, duration int64) {
	now := time.Now()
	updates := map[string]interface{}{
		"status":   status,
		"stdout":   stdout,
		"stderr":   stderr,
		"duration": duration,
	}

	if exitCode != nil {
		updates["exit_code"] = *exitCode
	}

	if status == "completed" || status == "failed" {
		updates["completed_at"] = &now
	}

	h.db.Model(&model.ProcessExecution{}).Where("id = ?", executionID).Updates(updates)
}
