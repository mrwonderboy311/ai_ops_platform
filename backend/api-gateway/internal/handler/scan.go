// Package handler provides HTTP handlers for scan operations
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/wangjialin/myops/pkg/model"
	"github.com/wangjialin/myops/pkg/ssh"
	"gorm.io/gorm"
)

// ScanHandler handles scan HTTP requests
type ScanHandler struct {
	db      *gorm.DB
	scanner *ssh.Scanner
	taskRepo interface{} // Will be ScanTaskRepository
}

// NewScanHandler creates a new ScanHandler
func NewScanHandler(db *gorm.DB) *ScanHandler {
	return &ScanHandler{
		db:      db,
		scanner: ssh.NewScanner(5 * time.Second),
	}
}

// ScanRequest represents a scan request
type ScanRequest struct {
	IPRange        string `json:"ipRange"`
	Ports          []int  `json:"ports"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
}

// ScanResponse represents the initial scan response
type ScanResponse struct {
	TaskID         string `json:"taskId"`
	Status         string `json:"status"`
	IPRange        string `json:"ipRange"`
	EstimatedHosts  int    `json:"estimatedHosts"`
}

// ServeHTTP handles HTTP requests for scanning
func (h *ScanHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate input
	if req.IPRange == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "IP range is required")
		return
	}
	if len(req.Ports) == 0 {
		req.Ports = []int{22} // Default SSH port
	}
	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = 5 // Default timeout
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

	// Calculate estimated hosts
	estimatedHosts, err := sshscanner.GetEstimatedHostCount(req.IPRange)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_IP_RANGE", "Invalid IP range format")
		return
	}

	// Create scan task
	taskID := uuid.New()

	// Convert []int to pq.Int64Array for database
	portsArray := make(pq.Int64Array, len(req.Ports))
	for i, p := range req.Ports {
		portsArray[i] = int64(p)
	}

	task := &model.ScanTask{
		ID:              taskID,
		UserID:          userID,
		IPRange:         req.IPRange,
		Ports:           portsArray,
		TimeoutSeconds:  req.TimeoutSeconds,
		Status:          model.ScanTaskStatusRunning,
		EstimatedHosts:  estimatedHosts * len(req.Ports),
		StartedAt:       time.Now(),
	}

	if err := h.db.Create(task).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create scan task")
		return
	}

	// Start background scan
	go h.runScan(taskID, req)

	// Return immediate response
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": ScanResponse{
			TaskID:        taskID.String(),
			Status:        string(model.ScanTaskStatusRunning),
			IPRange:       req.IPRange,
			EstimatedHosts: estimatedHosts * len(req.Ports),
		},
		"requestId": generateRequestID(),
	})
}

// runScan runs the scan in the background
func (h *ScanHandler) runScan(taskID uuid.UUID, req ScanRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.TimeoutSeconds)*time.Second+time.Minute)
	defer cancel()

	resultChan := make(chan *ssh.DiscoveredHost)

	config := &ssh.ScanConfig{
		IPRange:      req.IPRange,
		Ports:        req.Ports,
		Timeout:      time.Duration(req.TimeoutSeconds) * time.Second,
		MaxConcurrent: 50,
	}

	// Run scan
	if err := h.scanner.ScanRange(ctx, config, resultChan); err != nil {
		h.db.Model(&model.ScanTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
			"status":        model.ScanTaskStatusFailed,
			"error_message": err.Error(),
			"completed_at":   time.Now(),
		})
		return
	}

	// Process results
	for host := range resultChan {
		if host.Status == "success" || host.Status == "open" {
			// Save discovered host
			discovered := &model.DiscoveredHost{
				ScanTaskID: taskID,
				IPAddress: host.IPAddress,
				Port:      host.Port,
				Hostname:  host.Hostname,
				OSType:    host.OSType,
				OSVersion: host.OSVersion,
				Status:    host.Status,
				CreatedAt: time.Now(),
			}

			h.db.Create(discovered)
			h.db.Model(&model.ScanTask{}).Where("id = ?", taskID).
				UpdateColumn("discovered_hosts", gorm.Expr("discovered_hosts + 1"))
		}
	}

	// Mark as completed
	h.db.Model(&model.ScanTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":       model.ScanTaskStatusCompleted,
		"completed_at": time.Now(),
	})
}

// GetScanStatus handles scan task status queries
func (h *ScanHandler) GetScanStatus(w http.ResponseWriter, r *http.Request) {
	// Extract task ID from path
	// Path format: /api/v1/hosts/scan-tasks/{taskId}
	prefix := "/api/v1/hosts/scan-tasks/"
	if len(r.URL.Path) < len(prefix) {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Task ID is required")
		return
	}

	taskID := r.URL.Path[len(prefix):]
	if taskID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Task ID is required")
		return
	}

	// Query task
	var task model.ScanTask
	err := h.db.Where("id = ?", taskID).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Scan task not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	// Get discovered hosts if completed
	var hosts []model.DiscoveredHost
	if task.Status == model.ScanTaskStatusCompleted {
		err = h.db.Where("scan_task_id = ?", taskID).Find(&hosts).Error
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"taskId":         task.ID,
			"status":         string(task.Status),
			"discoveredHosts": task.DiscoveredHosts,
			"hosts":          hosts,
			"ipRange":        task.IPRange,
		},
		"requestId": generateRequestID(),
	})
}
