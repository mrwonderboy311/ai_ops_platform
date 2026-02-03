// Package handler provides HTTP handlers for agent operations
package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/config"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// AgentHandler handles agent HTTP requests
type AgentHandler struct {
	db                *gorm.DB
	autoApprovalConfig *config.AutoApprovalConfig
}

// NewAgentHandler creates a new AgentHandler
func NewAgentHandler(db *gorm.DB) *AgentHandler {
	return &AgentHandler{
		db:                db,
		autoApprovalConfig: config.DefaultAutoApprovalConfig(),
	}
}

// AgentReportRequest represents an agent report request
type AgentReportRequest struct {
	Hostname      string                 `json:"hostname"`
	IPAddress     string                 `json:"ipAddress"`
	MACAddress    string                 `json:"macAddress,omitempty"`
	Gateway       string                 `json:"gateway,omitempty"`
	OSType        string                 `json:"osType"`
	OSVersion     string                 `json:"osVersion"`
	KernelVersion string                 `json:"kernelVersion"`
	Arch          string                 `json:"arch"`
	CPUModel      string                 `json:"cpuModel,omitempty"`
	CPUCores      int32                  `json:"cpuCores"`
	MemoryTotal   uint64                 `json:"memoryTotal"` // bytes
	Disks         []json.RawMessage      `json:"disks,omitempty"`
	Networks      []json.RawMessage      `json:"networks,omitempty"`
}

// ServeHTTP handles HTTP requests for agent reporting
func (h *AgentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req AgentReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate required fields
	if req.IPAddress == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "IP address is required")
		return
	}

	// Find existing host by IP address
	var host model.Host
	err := h.db.Where("ip_address = ?", req.IPAddress).First(&host).Error

	now := time.Now()

	if err == gorm.ErrRecordNotFound {
		// Create labels first (needed for auto-approval check)
		labels := make(model.LabelMap)
		if req.Arch != "" {
			labels["arch"] = req.Arch
		}
		if req.KernelVersion != "" {
			labels["kernel_version"] = req.KernelVersion
		}
		if req.CPUModel != "" {
			labels["cpu_model"] = req.CPUModel
		}

		// Determine status based on auto-approval rules
		status := model.HostStatusPending // Default to pending
		if h.autoApprovalConfig.ShouldAutoApprove(req.IPAddress, labels) {
			status = model.HostStatusApproved
		}

		// Create new host
		host = model.Host{
			ID:        uuid.New(),
			Hostname:  req.Hostname,
			IPAddress: req.IPAddress,
			Port:      22,
			Status:    status,
			OSType:    req.OSType,
			OSVersion: req.OSVersion,
			LastSeenAt: &now,
			Labels:    labels,
		}

		// Set CPU cores if provided
		if req.CPUCores > 0 {
			cores := int(req.CPUCores)
			host.CPUCores = &cores
		}

		// Set memory in GB
		if req.MemoryTotal > 0 {
			memGB := int(req.MemoryTotal / (1024 * 1024 * 1024))
			host.MemoryGB = &memGB
		}

		if err := h.db.Create(&host).Error; err != nil {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create host")
			return
		}
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	} else {
		// Check if host is rejected
		if host.Status == model.HostStatusRejected {
			respondWithError(w, http.StatusForbidden, "HOST_REJECTED", "Host has been rejected and cannot report")
			return
		}

		// Update existing host
		updates := map[string]interface{}{
			"last_seen_at": now,
		}

		if req.Hostname != "" {
			updates["hostname"] = req.Hostname
		}
		if req.OSType != "" {
			updates["os_type"] = req.OSType
		}
		if req.OSVersion != "" {
			updates["os_version"] = req.OSVersion
		}

		// Update CPU cores if provided
		if req.CPUCores > 0 {
			cores := int(req.CPUCores)
			updates["cpu_cores"] = cores
		}

		// Update memory if provided
		if req.MemoryTotal > 0 {
			memGB := int(req.MemoryTotal / (1024 * 1024 * 1024))
			updates["memory_gb"] = memGB
		}

		// Update labels with additional info
		host.Labels = make(model.LabelMap)
		if req.Arch != "" {
			host.Labels["arch"] = req.Arch
		}
		if req.KernelVersion != "" {
			host.Labels["kernel_version"] = req.KernelVersion
		}
		if req.CPUModel != "" {
			host.Labels["cpu_model"] = req.CPUModel
		}
		updates["labels"] = host.Labels

		// Update status to online if was approved
		if host.Status == model.HostStatusApproved {
			updates["status"] = model.HostStatusOnline
		}

		if err := h.db.Model(&host).Updates(updates).Error; err != nil {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update host")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"success":   true,
			"hostId":    host.ID,
			"status":    string(host.Status),
			"message":   "Report received successfully",
		},
		"requestId": generateRequestID(),
	})
}
