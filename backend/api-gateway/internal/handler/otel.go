// Package handler provides HTTP handlers for OpenTelemetry Collector management
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// OtelHandler handles OpenTelemetry Collector operations
type OtelHandler struct {
	db *gorm.DB
}

// NewOtelHandler creates a new Otel handler
func NewOtelHandler(db *gorm.DB) *OtelHandler {
	return &OtelHandler{db: db}
}

// CreateCollector creates a new OpenTelemetry collector deployment
func (h *OtelHandler) CreateCollector(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCollectorRequest
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

	// Verify cluster ownership
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", req.ClusterID, userUUID).First(&cluster).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch cluster")
		}
		return
	}

	// Check if collector name already exists for this cluster/namespace
	var existingCollector model.OtelCollector
	if err := h.db.Where("cluster_id = ? AND namespace = ? AND name = ?", req.ClusterID, req.Namespace, req.Name).First(&existingCollector).Error; err == nil {
		respondWithError(w, http.StatusConflict, "CONFLICT", "Collector name already exists in this namespace")
		return
	}

	// Create collector
	collector := model.OtelCollector{
		ID:              uuid.New(),
		UserID:          userUUID,
		ClusterID:       req.ClusterID,
		Name:            req.Name,
		Namespace:       req.Namespace,
		Type:            req.Type,
		Status:          model.CollectorStatusPending,
		Config:          req.Config,
		Replicas:        req.Replicas,
		Resources:       req.Resources,
		MetricsEndpoint: req.MetricsEndpoint,
		LogsEndpoint:    req.LogsEndpoint,
		TracesEndpoint:  req.TracesEndpoint,
	}

	if collector.Replicas == 0 {
		collector.Replicas = 1
	}

	if err := h.db.Create(&collector).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create collector")
		return
	}

	// TODO: Trigger Kubernetes deployment
	// This would use the Kubernetes client to deploy the OTEL collector

	respondWithJSON(w, http.StatusCreated, collector)
}

// ListCollectors lists all OpenTelemetry collectors
func (h *OtelHandler) ListCollectors(w http.ResponseWriter, r *http.Request) {
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
	query := h.db.Model(&model.OtelCollector{}).Where("user_id = ?", userUUID)

	// Apply filters
	if clusterID := r.URL.Query().Get("clusterId"); clusterID != "" {
		clusterUUID, err := uuid.Parse(clusterID)
		if err == nil {
			query = query.Where("cluster_id = ?", clusterUUID)
		}
	}
	if namespace := r.URL.Query().Get("namespace"); namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}
	if collectorType := r.URL.Query().Get("type"); collectorType != "" {
		query = query.Where("type = ?", collectorType)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch collectors with cluster info
	var collectors []model.OtelCollector
	offset := (page - 1) * pageSize
	if err := query.Preload("Cluster").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&collectors).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch collectors")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       collectors,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetCollector gets a specific OpenTelemetry collector
func (h *OtelHandler) GetCollector(w http.ResponseWriter, r *http.Request) {
	// Extract collector ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collector ID")
		return
	}

	collectorID := parts[4]
	collectorUUID, err := uuid.Parse(collectorID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collector ID format")
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

	// Fetch collector
	var collector model.OtelCollector
	if err := h.db.Preload("Cluster").Where("id = ? AND user_id = ?", collectorUUID, userUUID).First(&collector).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Collector not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch collector")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, collector)
}

// UpdateCollector updates an OpenTelemetry collector
func (h *OtelHandler) UpdateCollector(w http.ResponseWriter, r *http.Request) {
	// Extract collector ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collector ID")
		return
	}

	collectorID := parts[4]
	collectorUUID, err := uuid.Parse(collectorID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collector ID format")
		return
	}

	var req model.UpdateCollectorRequest
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

	// Fetch collector
	var collector model.OtelCollector
	if err := h.db.Where("id = ? AND user_id = ?", collectorUUID, userUUID).First(&collector).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Collector not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch collector")
		}
		return
	}

	// Update fields
	updates := map[string]interface{}{}
	if req.Config != "" {
		updates["config"] = req.Config
	}
	if req.Replicas > 0 {
		updates["replicas"] = req.Replicas
	}
	if req.Resources != "" {
		updates["resources"] = req.Resources
	}
	if req.MetricsEndpoint != "" {
		updates["metrics_endpoint"] = req.MetricsEndpoint
	}
	if req.LogsEndpoint != "" {
		updates["logs_endpoint"] = req.LogsEndpoint
	}
	if req.TracesEndpoint != "" {
		updates["traces_endpoint"] = req.TracesEndpoint
	}

	if err := h.db.Model(&collector).Updates(updates).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update collector")
		return
	}

	// TODO: Trigger Kubernetes update

	// Fetch updated collector
	h.db.Preload("Cluster").First(&collector, collectorUUID)
	respondWithJSON(w, http.StatusOK, collector)
}

// DeleteCollector deletes an OpenTelemetry collector
func (h *OtelHandler) DeleteCollector(w http.ResponseWriter, r *http.Request) {
	// Extract collector ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collector ID")
		return
	}

	collectorID := parts[4]
	collectorUUID, err := uuid.Parse(collectorID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collector ID format")
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

	// Fetch collector
	var collector model.OtelCollector
	if err := h.db.Where("id = ? AND user_id = ?", collectorUUID, userUUID).First(&collector).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Collector not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch collector")
		}
		return
	}

	// TODO: Trigger Kubernetes undeployment

	// Delete collector
	if err := h.db.Delete(&collector).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete collector")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Collector deleted successfully",
	})
}

// GetCollectorStatus gets the status of a collector deployment
func (h *OtelHandler) GetCollectorStatus(w http.ResponseWriter, r *http.Request) {
	// Extract collector ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request")
		return
	}

	collectorID := parts[4]
	collectorUUID, err := uuid.Parse(collectorID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collector ID format")
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

	// Fetch collector
	var collector model.OtelCollector
	if err := h.db.Where("id = ? AND user_id = ?", collectorUUID, userUUID).First(&collector).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Collector not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch collector")
		}
		return
	}

	// TODO: Get actual status from Kubernetes
	status := model.CollectorDeploymentStatus{
		Status:        collector.Status,
		PodNames:      []string{},
		Replicas:      collector.Replicas,
		ReadyReplicas: 0,
		ErrorMessage:  "",
	}

	if collector.Status == model.CollectorStatusRunning {
		status.ReadyReplicas = collector.Replicas
		status.MetricsURL = "/metrics"
	}

	respondWithJSON(w, http.StatusOK, status)
}

// StartCollector starts a collector
func (h *OtelHandler) StartCollector(w http.ResponseWriter, r *http.Request) {
	// Extract collector ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request")
		return
	}

	collectorID := parts[4]
	collectorUUID, err := uuid.Parse(collectorID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collector ID format")
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

	// Fetch collector
	var collector model.OtelCollector
	if err := h.db.Where("id = ? AND user_id = ?", collectorUUID, userUUID).First(&collector).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Collector not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch collector")
		}
		return
	}

	// TODO: Trigger Kubernetes deployment

	// Update status
	collector.Status = model.CollectorStatusDeploying
	h.db.Save(&collector)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Collector deployment initiated",
	})
}

// StopCollector stops a collector
func (h *OtelHandler) StopCollector(w http.ResponseWriter, r *http.Request) {
	// Extract collector ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request")
		return
	}

	collectorID := parts[4]
	collectorUUID, err := uuid.Parse(collectorID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collector ID format")
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

	// Fetch collector
	var collector model.OtelCollector
	if err := h.db.Where("id = ? AND user_id = ?", collectorUUID, userUUID).First(&collector).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Collector not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch collector")
		}
		return
	}

	// TODO: Trigger Kubernetes undeployment

	// Update status
	collector.Status = model.CollectorStatusStopped
	h.db.Save(&collector)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Collector stopped successfully",
	})
}

// RestartCollector restarts a collector
func (h *OtelHandler) RestartCollector(w http.ResponseWriter, r *http.Request) {
	// Extract collector ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request")
		return
	}

	collectorID := parts[4]
	collectorUUID, err := uuid.Parse(collectorID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid collector ID format")
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

	// Fetch collector
	var collector model.OtelCollector
	if err := h.db.Where("id = ? AND user_id = ?", collectorUUID, userUUID).First(&collector).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Collector not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch collector")
		}
		return
	}

	// TODO: Trigger Kubernetes restart

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Collector restart initiated",
	})
}
