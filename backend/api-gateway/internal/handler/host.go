// Package handler provides HTTP handlers for host management
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// HostHandler handles host HTTP requests
type HostHandler struct {
	db *gorm.DB
}

// NewHostHandler creates a new HostHandler
func NewHostHandler(db *gorm.DB) *HostHandler {
	return &HostHandler{db: db}
}

// CreateHostRequest represents a create host request
type CreateHostRequest struct {
	Hostname  string            `json:"hostname"`
	IPAddress string            `json:"ipAddress"`
	Port      int               `json:"port"`
	OSType    string            `json:"osType"`
	OSVersion string            `json:"osVersion"`
	CPUCores  *int              `json:"cpuCores"`
	MemoryGB  *int              `json:"memoryGB"`
	DiskGB    *int64            `json:"diskGB"`
	Labels    map[string]string `json:"labels"`
	Tags      []string          `json:"tags"`
	ClusterID *string           `json:"clusterId"`
}

// UpdateHostRequest represents an update host request
type UpdateHostRequest struct {
	Hostname  string            `json:"hostname"`
	Port      int               `json:"port"`
	OSType    string            `json:"osType"`
	OSVersion string            `json:"osVersion"`
	CPUCores  *int              `json:"cpuCores"`
	MemoryGB  *int              `json:"memoryGB"`
	DiskGB    *int64            `json:"diskGB"`
	Labels    map[string]string `json:"labels"`
	Tags      []string          `json:"tags"`
	ClusterID *string           `json:"clusterId"`
}

// HostFilter represents filter options for listing hosts
type HostFilter struct {
	Page        int              `json:"page"`
	PageSize    int              `json:"pageSize"`
	Status      model.HostStatus `json:"status"`
	Hostname    string           `json:"hostname"`
	IPAddress   string           `json:"ipAddress"`
	RegisteredBy *uuid.UUID      `json:"registeredBy"`
	Labels      map[string]string `json:"labels"`
	Tags        []string         `json:"tags"`
	SortBy      string           `json:"sortBy"`
	SortDesc    bool             `json:"sortDesc"`
}

// ServeHTTP handles HTTP requests for hosts (router-based)
func (h *HostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/api/v1/hosts" || r.URL.Path == "/api/v1/hosts/" {
			h.listHosts(w, r)
		} else {
			h.getHost(w, r)
		}
	case http.MethodPost:
		h.createHost(w, r)
	case http.MethodPut:
		h.updateHost(w, r)
	case http.MethodDelete:
		h.deleteHost(w, r)
	case http.MethodPatch:
		h.approveHost(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// getHost retrieves a single host by ID
func (h *HostHandler) getHost(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	idStr := r.URL.Path[len("/api/v1/hosts/"):]
	if idStr == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Host ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "Invalid host ID format")
		return
	}

	var host model.Host
	err = h.db.Preload("RegisteredByUser").Preload("ApprovedByUser").
		Where("id = ?", id).First(&host).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      host,
		"requestId": generateRequestID(),
	})
}

// listHosts returns a paginated list of hosts
func (h *HostHandler) listHosts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := buildHostFilter(query)

	var hosts []model.Host
	var total int64

	dbQuery := h.db.Model(&model.Host{})

	// Apply filters
	if filter.Status != "" {
		dbQuery = dbQuery.Where("status = ?", filter.Status)
	}
	if filter.Hostname != "" {
		dbQuery = dbQuery.Where("hostname ILIKE ?", "%"+filter.Hostname+"%")
	}
	if filter.IPAddress != "" {
		dbQuery = dbQuery.Where("ip_address ILIKE ?", "%"+filter.IPAddress+"%")
	}

	// Count total
	if err := dbQuery.Count(&total).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Apply pagination
	if filter.PageSize > 0 {
		dbQuery = dbQuery.Limit(filter.PageSize)
	}
	if filter.Page > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		dbQuery = dbQuery.Offset(offset)
	}

	// Load associations and query
	if err := dbQuery.Preload("RegisteredByUser").Preload("ApprovedByUser").
		Find(&hosts).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"hosts": hosts,
			"total": total,
			"page":  filter.Page,
			"pageSize": filter.PageSize,
		},
		"requestId": generateRequestID(),
	})
}

// createHost creates a new host
func (h *HostHandler) createHost(w http.ResponseWriter, r *http.Request) {
	var req CreateHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context (set by auth middleware)
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	// Check if host already exists
	var existingHost model.Host
	err := h.db.Where("ip_address = ? AND port = ?", req.IPAddress, req.Port).First(&existingHost).Error
	if err == nil {
		respondWithError(w, http.StatusConflict, "HOST_EXISTS", "Host with this IP:port already exists")
		return
	} else if err != gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	host := &model.Host{
		ID:          uuid.New(),
		Hostname:    req.Hostname,
		IPAddress:   req.IPAddress,
		Port:        req.Port,
		Status:      model.HostStatusPending,
		OSType:      req.OSType,
		OSVersion:   req.OSVersion,
		CPUCores:    req.CPUCores,
		MemoryGB:    req.MemoryGB,
		DiskGB:      req.DiskGB,
		Labels:      req.Labels,
		Tags:        req.Tags,
		RegisteredBy: &userID,
	}

	if req.ClusterID != nil && *req.ClusterID != "" {
		clusterID, err := uuid.Parse(*req.ClusterID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID format")
			return
		}
		host.ClusterID = &clusterID
	}

	if err := h.db.Create(host).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      host,
		"requestId": generateRequestID(),
	})
}

// updateHost updates an existing host
func (h *HostHandler) updateHost(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	idStr := r.URL.Path[len("/api/v1/hosts/"):]
	if idStr == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Host ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "Invalid host ID format")
		return
	}

	var host model.Host
	err = h.db.Where("id = ?", id).First(&host).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	var req UpdateHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Hostname != "" {
		updates["hostname"] = req.Hostname
	}
	if req.Port != 0 {
		updates["port"] = req.Port
	}
	if req.OSType != "" {
		updates["os_type"] = req.OSType
	}
	if req.OSVersion != "" {
		updates["os_version"] = req.OSVersion
	}
	if req.CPUCores != nil {
		updates["cpu_cores"] = req.CPUCores
	}
	if req.MemoryGB != nil {
		updates["memory_gb"] = req.MemoryGB
	}
	if req.DiskGB != nil {
		updates["disk_gb"] = req.DiskGB
	}
	if req.Labels != nil {
		updates["labels"] = req.Labels
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.ClusterID != nil {
		if *req.ClusterID == "" {
			updates["cluster_id"] = nil
		} else {
			clusterID, err := uuid.Parse(*req.ClusterID)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID format")
				return
			}
			updates["cluster_id"] = clusterID
		}
	}

	if err := h.db.Model(&host).Updates(updates).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Reload to get updated data
	h.db.Preload("RegisteredByUser").Preload("ApprovedByUser").First(&host, "id = ?", id)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      host,
		"requestId": generateRequestID(),
	})
}

// deleteHost deletes a host
func (h *HostHandler) deleteHost(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	idStr := r.URL.Path[len("/api/v1/hosts/"):]
	if idStr == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Host ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "Invalid host ID format")
		return
	}

	if err := h.db.Delete(&model.Host{}, "id = ?", id).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// approveHost approves a host registration
func (h *HostHandler) approveHost(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	// Path format: /api/v1/hosts/{id}/approve
	path := r.URL.Path
	if len(path) < len("/api/v1/hosts//approve") {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request path")
		return
	}

	// Extract ID between /api/v1/hosts/ and /approve
	idStr := path[len("/api/v1/hosts/"):]
	approveIndex := len(idStr) - len("/approve")
	if approveIndex <= 0 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Host ID is required")
		return
	}
	idStr = idStr[:approveIndex]

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ID", "Invalid host ID format")
		return
	}

	// Get user ID from context (set by auth middleware)
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

	// Update host status
	updates := map[string]interface{}{
		"status":      model.HostStatusApproved,
		"approved_by": userID,
		"approved_at": gorm.Expr("NOW()"),
	}

	if err := h.db.Model(&model.Host{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	// Fetch updated host
	var host model.Host
	h.db.Preload("RegisteredByUser").Preload("ApprovedByUser").First(&host, "id = ?", id)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      host,
		"requestId": generateRequestID(),
	})
}

// buildHostFilter builds a HostFilter from URL query parameters
func buildHostFilter(query map[string][]string) *HostFilter {
	filter := &HostFilter{
		Page:     1,
		PageSize: 20,
	}

	if page, ok := query["page"]; ok && len(page) > 0 {
		if p, err := strconv.Atoi(page[0]); err == nil {
			filter.Page = p
		}
	}
	if pageSize, ok := query["page_size"]; ok && len(pageSize) > 0 {
		if ps, err := strconv.Atoi(pageSize[0]); err == nil {
			filter.PageSize = ps
		}
	}
	if status, ok := query["status"]; ok && len(status) > 0 {
		filter.Status = model.HostStatus(status[0])
	}
	if hostname, ok := query["hostname"]; ok && len(hostname) > 0 {
		filter.Hostname = hostname[0]
	}
	if ipAddress, ok := query["ip_address"]; ok && len(ipAddress) > 0 {
		filter.IPAddress = ipAddress[0]
	}

	return filter
}
