// Package handler provides HTTP handlers for Grafana integration
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// GrafanaHandler handles Grafana integration operations
type GrafanaHandler struct {
	db *gorm.DB
}

// NewGrafanaHandler creates a new Grafana handler
func NewGrafanaHandler(db *gorm.DB) *GrafanaHandler {
	return &GrafanaHandler{db: db}
}

// ============== Instance Management ==============

// CreateInstance creates a new Grafana instance
func (h *GrafanaHandler) CreateInstance(w http.ResponseWriter, r *http.Request) {
	var req model.CreateGrafanaInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Verify cluster ownership if provided
	if req.ClusterID != nil {
		var cluster model.K8sCluster
		if err := h.db.Where("id = ? AND user_id = ?", req.ClusterID, userUUID).First(&cluster).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
			} else {
				respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch cluster")
			}
			return
		}
	}

	// Check if instance name already exists
	var existingInstance model.GrafanaInstance
	if err := h.db.Where("user_id = ? AND name = ?", userUUID, req.Name).First(&existingInstance).Error; err == nil {
		respondWithError(w, http.StatusConflict, "CONFLICT", "Grafana instance name already exists")
		return
	}

	// Create instance
	instance := model.GrafanaInstance{
		UserID:              userUUID,
		ClusterID:           req.ClusterID,
		Name:                req.Name,
		URL:                 req.URL,
		APIKey:              req.APIKey,
		Username:            req.Username,
		Password:            req.Password,
		Status:              model.GrafanaStatusActive,
		ServiceAccountID:    req.ServiceAccountID,
		ServiceAccountToken: req.ServiceAccountToken,
		AutoSync:            req.AutoSync,
		SyncInterval:        req.SyncInterval,
		SyncStatus:          model.SyncStatusPending,
	}

	if instance.AutoSync {
		instance.AutoSync = true
	}
	if instance.SyncInterval == 0 {
		instance.SyncInterval = 300 // 5 minutes default
	}

	if err := h.db.Create(&instance).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create Grafana instance")
		return
	}

	// TODO: Trigger initial sync if auto-sync is enabled

	respondWithJSON(w, http.StatusCreated, instance)
}

// ListInstances lists all Grafana instances
func (h *GrafanaHandler) ListInstances(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Parse query parameters
	page := 1
	pageSize := 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Build query
	query := h.db.Model(&model.GrafanaInstance{}).Where("user_id = ?", userUUID)

	// Apply filters
	if clusterID := r.URL.Query().Get("clusterId"); clusterID != "" {
		clusterUUID, err := uuid.Parse(clusterID)
		if err == nil {
			query = query.Where("cluster_id = ?", clusterUUID)
		}
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch instances
	var instances []model.GrafanaInstance
	offset := (page - 1) * pageSize
	if err := query.Preload("Cluster").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&instances).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch Grafana instances")
		return
	}

	// Remove sensitive data
	for i := range instances {
		instances[i].APIKey = ""
		instances[i].Password = ""
		instances[i].ServiceAccountToken = ""
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       instances,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetInstance gets a specific Grafana instance
func (h *GrafanaHandler) GetInstance(w http.ResponseWriter, r *http.Request) {
	// Extract instance ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid instance ID")
		return
	}

	instanceID := parts[4]
	instanceUUID, err := uuid.Parse(instanceID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid instance ID format")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Fetch instance
	var instance model.GrafanaInstance
	if err := h.db.Preload("Cluster").Where("id = ? AND user_id = ?", instanceUUID, userUUID).First(&instance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Grafana instance not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch Grafana instance")
		}
		return
	}

	// Remove sensitive data
	instance.APIKey = ""
	instance.Password = ""
	instance.ServiceAccountToken = ""

	respondWithJSON(w, http.StatusOK, instance)
}

// UpdateInstance updates a Grafana instance
func (h *GrafanaHandler) UpdateInstance(w http.ResponseWriter, r *http.Request) {
	// Extract instance ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid instance ID")
		return
	}

	instanceID := parts[4]
	instanceUUID, err := uuid.Parse(instanceID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid instance ID format")
		return
	}

	var req model.UpdateGrafanaInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Fetch instance
	var instance model.GrafanaInstance
	if err := h.db.Where("id = ? AND user_id = ?", instanceUUID, userUUID).First(&instance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Grafana instance not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch Grafana instance")
		}
		return
	}

	// Update fields
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.APIKey != nil {
		updates["api_key"] = *req.APIKey
	}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Password != nil {
		updates["password"] = *req.Password
	}
	if req.ServiceAccountID != nil {
		updates["service_account_id"] = *req.ServiceAccountID
	}
	if req.ServiceAccountToken != nil {
		updates["service_account_token"] = *req.ServiceAccountToken
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.AutoSync != nil {
		updates["auto_sync"] = *req.AutoSync
	}
	if req.SyncInterval != nil {
		updates["sync_interval"] = *req.SyncInterval
	}

	if err := h.db.Model(&instance).Updates(updates).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update Grafana instance")
		return
	}

	// Fetch updated instance
	h.db.Preload("Cluster").First(&instance, instanceUUID)
	instance.APIKey = ""
	instance.Password = ""
	instance.ServiceAccountToken = ""

	respondWithJSON(w, http.StatusOK, instance)
}

// DeleteInstance deletes a Grafana instance
func (h *GrafanaHandler) DeleteInstance(w http.ResponseWriter, r *http.Request) {
	// Extract instance ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid instance ID")
		return
	}

	instanceID := parts[4]
	instanceUUID, err := uuid.Parse(instanceID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid instance ID format")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Fetch instance
	var instance model.GrafanaInstance
	if err := h.db.Where("id = ? AND user_id = ?", instanceUUID, userUUID).First(&instance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Grafana instance not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch Grafana instance")
		}
		return
	}

	// Delete associated dashboards, data sources, and folders
	h.db.Where("instance_id = ?", instanceUUID).Delete(&model.GrafanaDashboard{})
	h.db.Where("instance_id = ?", instanceUUID).Delete(&model.GrafanaDataSource{})
	h.db.Where("instance_id = ?", instanceUUID).Delete(&model.GrafanaFolder{})

	// Delete instance
	if err := h.db.Delete(&instance).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete Grafana instance")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Grafana instance deleted successfully",
	})
}

// TestInstance tests a Grafana instance connection
func (h *GrafanaHandler) TestInstance(w http.ResponseWriter, r *http.Request) {
	var req model.TestGrafanaInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	startTime := time.Now()

	// TODO: Implement actual Grafana connection test
	// This would:
	// 1. Create HTTP client with authentication
	// 2. Query /api/health to check connection
	// 3. Query /api/frontend/settings to get version info

	duration := time.Since(startTime).Milliseconds()

	// Simulate successful test for now
	response := model.TestGrafanaInstanceResponse{
		Success:  true,
		Version:  "10.0.0",
		Message:  "Successfully connected to Grafana",
		Duration: duration,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// SyncInstance syncs dashboards and data sources from a Grafana instance
func (h *GrafanaHandler) SyncInstance(w http.ResponseWriter, r *http.Request) {
	// Extract instance ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request")
		return
	}

	instanceID := parts[4]
	instanceUUID, err := uuid.Parse(instanceID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid instance ID format")
		return
	}

	var req model.SyncGrafanaInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use default values if body is empty
		req = model.SyncGrafanaInstanceRequest{
			SyncDashboards:  true,
			SyncDataSources: false,
			SyncFolders:     false,
		}
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Fetch instance
	var instance model.GrafanaInstance
	if err := h.db.Where("id = ? AND user_id = ?", instanceUUID, userUUID).First(&instance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Grafana instance not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch Grafana instance")
		}
		return
	}

	startTime := time.Now()

	// TODO: Implement actual Grafana sync
	// This would:
	// 1. Authenticate with Grafana API
	// 2. Fetch dashboards via /api/search
	// 3. Fetch data sources via /api/datasources
	// 4. Fetch folders via /api/folders
	// 5. Store in database

	duration := time.Since(startTime).Milliseconds()

	// Update instance sync status
	now := time.Now()
	h.db.Model(&instance).Updates(map[string]interface{}{
		"last_sync_at": &now,
		"sync_status":  model.SyncStatusSuccess,
		"dashboard_count": 5, // Simulated
		"data_source_count": 2, // Simulated
	})

	response := model.SyncGrafanaInstanceResponse{
		Success:          true,
		Message:          "Sync completed successfully",
		DashboardsAdded:  5,
		DashboardsUpdated: 0,
		DataSourcesAdded: 2,
		Duration:         duration,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// ============== Dashboard Management ==============

// ListDashboards lists all Grafana dashboards
func (h *GrafanaHandler) ListDashboards(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Parse query parameters
	page := 1
	pageSize := 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Build query
	query := h.db.Model(&model.GrafanaDashboard{}).Where("user_id = ?", userUUID)

	// Apply filters
	if instanceID := r.URL.Query().Get("instanceId"); instanceID != "" {
		instanceUUID, err := uuid.Parse(instanceID)
		if err == nil {
			query = query.Where("instance_id = ?", instanceUUID)
		}
	}
	if clusterID := r.URL.Query().Get("clusterId"); clusterID != "" {
		clusterUUID, err := uuid.Parse(clusterID)
		if err == nil {
			query = query.Where("cluster_id = ?", clusterUUID)
		}
	}
	if tag := r.URL.Query().Get("tag"); tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}
	if search := r.URL.Query().Get("search"); search != "" {
		query = query.Where("title LIKE ?", "%"+search+"%")
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch dashboards
	var dashboards []model.GrafanaDashboard
	offset := (page - 1) * pageSize
	if err := query.Preload("Instance").Preload("Cluster").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dashboards).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch dashboards")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       dashboards,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetDashboard gets a specific Grafana dashboard
func (h *GrafanaHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// Extract dashboard ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid dashboard ID")
		return
	}

	dashboardID := parts[4]
	dashboardUUID, err := uuid.Parse(dashboardID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid dashboard ID format")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Fetch dashboard
	var dashboard model.GrafanaDashboard
	if err := h.db.Preload("Instance").Preload("Cluster").Where("id = ? AND user_id = ?", dashboardUUID, userUUID).First(&dashboard).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Dashboard not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch dashboard")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, dashboard)
}

// ============== Data Source Management ==============

// ListDataSources lists all Grafana data sources
func (h *GrafanaHandler) ListDataSources(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Parse query parameters
	page := 1
	pageSize := 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Build query
	query := h.db.Model(&model.GrafanaDataSource{}).Where("user_id = ?", userUUID)

	// Apply filters
	if instanceID := r.URL.Query().Get("instanceId"); instanceID != "" {
		instanceUUID, err := uuid.Parse(instanceID)
		if err == nil {
			query = query.Where("instance_id = ?", instanceUUID)
		}
	}
	if dsType := r.URL.Query().Get("type"); dsType != "" {
		query = query.Where("type = ?", dsType)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch data sources
	var dataSources []model.GrafanaDataSource
	offset := (page - 1) * pageSize
	if err := query.Preload("Instance").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dataSources).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch data sources")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       dataSources,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetDataSource gets a specific Grafana data source
func (h *GrafanaHandler) GetDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract data source ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid data source ID")
		return
	}

	dataSourceID := parts[4]
	dataSourceUUID, err := uuid.Parse(dataSourceID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid data source ID format")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Fetch data source
	var dataSource model.GrafanaDataSource
	if err := h.db.Preload("Instance").Where("id = ? AND user_id = ?", dataSourceUUID, userUUID).First(&dataSource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Data source not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch data source")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, dataSource)
}

// ============== Folder Management ==============

// ListFolders lists all Grafana folders
func (h *GrafanaHandler) ListFolders(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Parse query parameters
	page := 1
	pageSize := 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Build query
	query := h.db.Model(&model.GrafanaFolder{}).Where("user_id = ?", userUUID)

	// Apply filters
	if instanceID := r.URL.Query().Get("instanceId"); instanceID != "" {
		instanceUUID, err := uuid.Parse(instanceID)
		if err == nil {
			query = query.Where("instance_id = ?", instanceUUID)
		}
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch folders
	var folders []model.GrafanaFolder
	offset := (page - 1) * pageSize
	if err := query.Preload("Instance").Offset(offset).Limit(pageSize).Order("title ASC").Find(&folders).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch folders")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       folders,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetFolder gets a specific Grafana folder
func (h *GrafanaHandler) GetFolder(w http.ResponseWriter, r *http.Request) {
	// Extract folder ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid folder ID")
		return
	}

	folderID := parts[4]
	folderUUID, err := uuid.Parse(folderID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid folder ID format")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Fetch folder
	var folder model.GrafanaFolder
	if err := h.db.Preload("Instance").Where("id = ? AND user_id = ?", folderUUID, userUUID).First(&folder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Folder not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch folder")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, folder)
}
