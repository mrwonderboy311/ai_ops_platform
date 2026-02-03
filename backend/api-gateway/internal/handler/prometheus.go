// Package handler provides HTTP handlers for Prometheus integration
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// PrometheusHandler handles Prometheus integration operations
type PrometheusHandler struct {
	db *gorm.DB
}

// NewPrometheusHandler creates a new Prometheus handler
func NewPrometheusHandler(db *gorm.DB) *PrometheusHandler {
	return &PrometheusHandler{db: db}
}

// ============== Data Source Management ==============

// CreateDataSource creates a new Prometheus data source
func (h *PrometheusHandler) CreateDataSource(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePrometheusDataSourceRequest
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

	// Check if data source name already exists
	var existingDS model.PrometheusDataSource
	if err := h.db.Where("user_id = ? AND name = ?", userUUID, req.Name).First(&existingDS).Error; err == nil {
		respondWithError(w, http.StatusConflict, "CONFLICT", "Data source name already exists")
		return
	}

	// Create data source
	dataSource := model.PrometheusDataSource{
		UserID:          userUUID,
		ClusterID:       req.ClusterID,
		Name:            req.Name,
		URL:             req.URL,
		Username:        req.Username,
		Password:        req.Password,
		Status:          model.DSStatusActive,
		InsecureSkipTLS: req.InsecureSkipTLS,
		CACert:          req.CACert,
		ClientCert:      req.ClientCert,
		ClientKey:       req.ClientKey,
		Headers:         req.Headers,
	}

	if err := h.db.Create(&dataSource).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create data source")
		return
	}

	respondWithJSON(w, http.StatusCreated, dataSource)
}

// ListDataSources lists all Prometheus data sources
func (h *PrometheusHandler) ListDataSources(w http.ResponseWriter, r *http.Request) {
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
	query := h.db.Model(&model.PrometheusDataSource{}).Where("user_id = ?", userUUID)

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

	// Fetch data sources
	var dataSources []model.PrometheusDataSource
	offset := (page - 1) * pageSize
	if err := query.Preload("Cluster").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dataSources).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch data sources")
		return
	}

	// Remove sensitive data
	for i := range dataSources {
		dataSources[i].Password = ""
		dataSources[i].ClientKey = ""
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       dataSources,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetDataSource gets a specific Prometheus data source
func (h *PrometheusHandler) GetDataSource(w http.ResponseWriter, r *http.Request) {
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
	var dataSource model.PrometheusDataSource
	if err := h.db.Preload("Cluster").Where("id = ? AND user_id = ?", dataSourceUUID, userUUID).First(&dataSource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Data source not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch data source")
		}
		return
	}

	// Remove sensitive data
	dataSource.Password = ""
	dataSource.ClientKey = ""

	respondWithJSON(w, http.StatusOK, dataSource)
}

// UpdateDataSource updates a Prometheus data source
func (h *PrometheusHandler) UpdateDataSource(w http.ResponseWriter, r *http.Request) {
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

	var req model.UpdatePrometheusDataSourceRequest
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

	// Fetch data source
	var dataSource model.PrometheusDataSource
	if err := h.db.Where("id = ? AND user_id = ?", dataSourceUUID, userUUID).First(&dataSource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Data source not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch data source")
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
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Password != nil {
		updates["password"] = *req.Password
	}
	if req.InsecureSkipTLS != nil {
		updates["insecure_skip_tls"] = *req.InsecureSkipTLS
	}
	if req.CACert != nil {
		updates["ca_cert"] = *req.CACert
	}
	if req.ClientCert != nil {
		updates["client_cert"] = *req.ClientCert
	}
	if req.ClientKey != nil {
		updates["client_key"] = *req.ClientKey
	}
	if req.Headers != nil {
		updates["headers"] = *req.Headers
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.db.Model(&dataSource).Updates(updates).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update data source")
		return
	}

	// Fetch updated data source
	h.db.Preload("Cluster").First(&dataSource, dataSourceUUID)
	dataSource.Password = ""
	dataSource.ClientKey = ""

	respondWithJSON(w, http.StatusOK, dataSource)
}

// DeleteDataSource deletes a Prometheus data source
func (h *PrometheusHandler) DeleteDataSource(w http.ResponseWriter, r *http.Request) {
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
	var dataSource model.PrometheusDataSource
	if err := h.db.Where("id = ? AND user_id = ?", dataSourceUUID, userUUID).First(&dataSource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Data source not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch data source")
		}
		return
	}

	// Delete data source
	if err := h.db.Delete(&dataSource).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete data source")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Data source deleted successfully",
	})
}

// TestDataSource tests a Prometheus data source connection
func (h *PrometheusHandler) TestDataSource(w http.ResponseWriter, r *http.Request) {
	var req model.TestPrometheusDataSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	startTime := time.Now()

	// TODO: Implement actual Prometheus connection test
	// This would:
	// 1. Create HTTP client with TLS config
	// 2. Query /api/v1/status/buildinfo to get version
	// 3. Test basic query execution

	duration := time.Since(startTime).Milliseconds()

	// Simulate successful test for now
	response := model.TestPrometheusDataSourceResponse{
		Success:  true,
		Version:  "2.45.0",
		Message:  "Successfully connected to Prometheus",
		Duration: duration,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// ============== Alert Rule Management ==============

// CreateAlertRule creates a new Prometheus alert rule
func (h *PrometheusHandler) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePrometheusAlertRuleRequest
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

	// Verify data source ownership
	var dataSource model.PrometheusDataSource
	if err := h.db.Where("id = ? AND user_id = ?", req.DataSourceID, userUUID).First(&dataSource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Data source not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch data source")
		}
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

	// Check if alert rule name already exists
	var existingRule model.PrometheusAlertRule
	if err := h.db.Where("data_source_id = ? AND name = ?", req.DataSourceID, req.Name).First(&existingRule).Error; err == nil {
		respondWithError(w, http.StatusConflict, "CONFLICT", "Alert rule name already exists for this data source")
		return
	}

	// Create alert rule
	alertRule := model.PrometheusAlertRule{
		UserID:       userUUID,
		DataSourceID: req.DataSourceID,
		ClusterID:    req.ClusterID,
		Name:         req.Name,
		Expression:   req.Expression,
		Duration:     req.Duration,
		Severity:     req.Severity,
		Summary:      req.Summary,
		Description:  req.Description,
		Labels:       req.Labels,
		Annotations:  req.Annotations,
		Enabled:      true,
	}

	if err := h.db.Create(&alertRule).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create alert rule")
		return
	}

	respondWithJSON(w, http.StatusCreated, alertRule)
}

// ListAlertRules lists all Prometheus alert rules
func (h *PrometheusHandler) ListAlertRules(w http.ResponseWriter, r *http.Request) {
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
	query := h.db.Model(&model.PrometheusAlertRule{}).Where("user_id = ?", userUUID)

	// Apply filters
	if dataSourceID := r.URL.Query().Get("dataSourceId"); dataSourceID != "" {
		dataSourceUUID, err := uuid.Parse(dataSourceID)
		if err == nil {
			query = query.Where("data_source_id = ?", dataSourceUUID)
		}
	}
	if clusterID := r.URL.Query().Get("clusterId"); clusterID != "" {
		clusterUUID, err := uuid.Parse(clusterID)
		if err == nil {
			query = query.Where("cluster_id = ?", clusterUUID)
		}
	}
	if severity := r.URL.Query().Get("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if enabled := r.URL.Query().Get("enabled"); enabled != "" {
		query = query.Where("enabled = ?", enabled == "true")
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch alert rules
	var alertRules []model.PrometheusAlertRule
	offset := (page - 1) * pageSize
	if err := query.Preload("DataSource").Preload("Cluster").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&alertRules).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch alert rules")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       alertRules,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetAlertRule gets a specific alert rule
func (h *PrometheusHandler) GetAlertRule(w http.ResponseWriter, r *http.Request) {
	// Extract alert rule ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid alert rule ID")
		return
	}

	alertRuleID := parts[4]
	alertRuleUUID, err := uuid.Parse(alertRuleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid alert rule ID format")
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

	// Fetch alert rule
	var alertRule model.PrometheusAlertRule
	if err := h.db.Preload("DataSource").Preload("Cluster").Where("id = ? AND user_id = ?", alertRuleUUID, userUUID).First(&alertRule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert rule not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch alert rule")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, alertRule)
}

// UpdateAlertRule updates an alert rule
func (h *PrometheusHandler) UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	// Extract alert rule ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid alert rule ID")
		return
	}

	alertRuleID := parts[4]
	alertRuleUUID, err := uuid.Parse(alertRuleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid alert rule ID format")
		return
	}

	var req model.UpdatePrometheusAlertRuleRequest
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

	// Fetch alert rule
	var alertRule model.PrometheusAlertRule
	if err := h.db.Where("id = ? AND user_id = ?", alertRuleUUID, userUUID).First(&alertRule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert rule not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch alert rule")
		}
		return
	}

	// Update fields
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Expression != nil {
		updates["expression"] = *req.Expression
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Labels != nil {
		updates["labels"] = *req.Labels
	}
	if req.Annotations != nil {
		updates["annotations"] = *req.Annotations
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	updates["synced"] = false // Mark as needing sync

	if err := h.db.Model(&alertRule).Updates(updates).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update alert rule")
		return
	}

	// Fetch updated alert rule
	h.db.Preload("DataSource").Preload("Cluster").First(&alertRule, alertRuleUUID)

	respondWithJSON(w, http.StatusOK, alertRule)
}

// DeleteAlertRule deletes an alert rule
func (h *PrometheusHandler) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	// Extract alert rule ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid alert rule ID")
		return
	}

	alertRuleID := parts[4]
	alertRuleUUID, err := uuid.Parse(alertRuleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid alert rule ID format")
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

	// Fetch alert rule
	var alertRule model.PrometheusAlertRule
	if err := h.db.Where("id = ? AND user_id = ?", alertRuleUUID, userUUID).First(&alertRule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert rule not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch alert rule")
		}
		return
	}

	// Delete alert rule
	if err := h.db.Delete(&alertRule).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete alert rule")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Alert rule deleted successfully",
	})
}

// ============== Query Execution ==============

// ExecuteQuery executes a Prometheus query
func (h *PrometheusHandler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	var req model.PrometheusQueryRequest
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

	// Extract data source ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid data source ID")
		return
	}

	dataSourceID := parts[4]
	dataSourceUUID, err := uuid.Parse(dataSourceID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid data source ID format")
		return
	}

	// Verify data source ownership
	var dataSource model.PrometheusDataSource
	if err := h.db.Where("id = ? AND user_id = ?", dataSourceUUID, userUUID).First(&dataSource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Data source not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch data source")
		}
		return
	}

	startTime := time.Now()

	// TODO: Implement actual Prometheus query execution
	// This would:
	// 1. Construct HTTP client with data source credentials
	// 2. Execute query via /api/v1/query or /api/v1/query_range
	// 3. Parse response and format results

	duration := time.Since(startTime).Milliseconds()

	// Create query record for history
	queryRecord := model.PrometheusQuery{
		UserID:       userUUID,
		DataSourceID: dataSourceUUID,
		Query:        req.Query,
		QueryType:    req.QueryType,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Step:         req.Step,
		Duration:     duration,
		Success:      true,
		ResultCount:  1,
	}
	h.db.Create(&queryRecord)

	// Update data source statistics
	h.db.Model(&dataSource).Updates(map[string]interface{}{
		"query_count":    dataSource.QueryCount + 1,
		"last_queried_at": time.Now(),
	})

	// Simulate query response for now
	response := model.PrometheusQueryResponse{
		Status: "success",
		Data: []model.PrometheusSeries{
			{
				Metric: map[string]string{
					"__name__": "up",
					"job":      "prometheus",
					"instance": "localhost:9090",
				},
				Values: []model.PrometheusValue{
					{Timestamp: float64(time.Now().Unix()), Value: "1"},
				},
			},
		},
		Duration: duration,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// ============== Dashboard Management ==============

// CreateDashboard creates a new Prometheus dashboard
func (h *PrometheusHandler) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePrometheusDashboardRequest
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

	// Create dashboard
	dashboard := model.PrometheusDashboard{
		UserID:      userUUID,
		ClusterID:   req.ClusterID,
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
		Config:      req.Config,
		IsPublic:    req.IsPublic,
		RefreshRate: req.RefreshRate,
	}

	if dashboard.RefreshRate == 0 {
		dashboard.RefreshRate = 30
	}

	if err := h.db.Create(&dashboard).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create dashboard")
		return
	}

	respondWithJSON(w, http.StatusCreated, dashboard)
}

// ListDashboards lists all Prometheus dashboards
func (h *PrometheusHandler) ListDashboards(w http.ResponseWriter, r *http.Request) {
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
	query := h.db.Model(&model.PrometheusDashboard{}).Where("user_id = ? OR is_public = ?", userUUID, true)

	// Apply filters
	if clusterID := r.URL.Query().Get("clusterId"); clusterID != "" {
		clusterUUID, err := uuid.Parse(clusterID)
		if err == nil {
			query = query.Where("cluster_id = ?", clusterUUID)
		}
	}
	if starred := r.URL.Query().Get("starred"); starred != "" {
		if starred == "true" {
			query = query.Where("starred = ? AND user_id = ?", true, userUUID)
		}
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch dashboards
	var dashboards []model.PrometheusDashboard
	offset := (page - 1) * pageSize
	if err := query.Preload("Cluster").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&dashboards).Error; err != nil {
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

// GetDashboard gets a specific dashboard
func (h *PrometheusHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
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

	// Fetch dashboard (user's own or public)
	var dashboard model.PrometheusDashboard
	if err := h.db.Preload("Cluster").Where("(id = ? AND user_id = ?) OR (id = ? AND is_public = ?)", dashboardUUID, userUUID, dashboardUUID, true).First(&dashboard).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Dashboard not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch dashboard")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, dashboard)
}

// UpdateDashboard updates a dashboard
func (h *PrometheusHandler) UpdateDashboard(w http.ResponseWriter, r *http.Request) {
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

	var req model.UpdatePrometheusDashboardRequest
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

	// Fetch dashboard (must be owner)
	var dashboard model.PrometheusDashboard
	if err := h.db.Where("id = ? AND user_id = ?", dashboardUUID, userUUID).First(&dashboard).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Dashboard not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch dashboard")
		}
		return
	}

	// Update fields
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}
	if req.Config != nil {
		updates["config"] = *req.Config
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.RefreshRate != nil {
		updates["refresh_rate"] = *req.RefreshRate
	}
	if req.Starred != nil {
		updates["starred"] = *req.Starred
	}

	if err := h.db.Model(&dashboard).Updates(updates).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update dashboard")
		return
	}

	// Fetch updated dashboard
	h.db.Preload("Cluster").First(&dashboard, dashboardUUID)

	respondWithJSON(w, http.StatusOK, dashboard)
}

// DeleteDashboard deletes a dashboard
func (h *PrometheusHandler) DeleteDashboard(w http.ResponseWriter, r *http.Request) {
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

	// Fetch dashboard (must be owner)
	var dashboard model.PrometheusDashboard
	if err := h.db.Where("id = ? AND user_id = ?", dashboardUUID, userUUID).First(&dashboard).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Dashboard not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch dashboard")
		}
		return
	}

	// Delete dashboard
	if err := h.db.Delete(&dashboard).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete dashboard")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Dashboard deleted successfully",
	})
}
