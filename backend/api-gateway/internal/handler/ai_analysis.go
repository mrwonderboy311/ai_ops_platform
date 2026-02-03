// Package handler provides HTTP handlers for AI analysis operations
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

// AIAnalysisHandler handles AI analysis operations
type AIAnalysisHandler struct {
	db *gorm.DB
}

// NewAIAnalysisHandler creates a new AI analysis handler
func NewAIAnalysisHandler(db *gorm.DB) *AIAnalysisHandler {
	return &AIAnalysisHandler{db: db}
}

// ============== Anomaly Detection Rules ==============

// CreateAnomalyRule creates a new anomaly detection rule
func (h *AIAnalysisHandler) CreateAnomalyRule(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAnomalyDetectionRuleRequest
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

	// Verify data source ownership if provided
	if req.DataSourceID != nil {
		var dataSource model.PrometheusDataSource
		if err := h.db.Where("id = ? AND user_id = ?", req.DataSourceID, userUUID).First(&dataSource).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Data source not found")
			} else {
				respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch data source")
			}
			return
		}
	}

	// Check if rule name already exists
	var existingRule model.AnomalyDetectionRule
	if err := h.db.Where("user_id = ? AND name = ?", userUUID, req.Name).First(&existingRule).Error; err == nil {
		respondWithError(w, http.StatusConflict, "CONFLICT", "Anomaly detection rule name already exists")
		return
	}

	// Create rule
	rule := model.AnomalyDetectionRule{
		UserID:            userUUID,
		ClusterID:          req.ClusterID,
		DataSourceID:       req.DataSourceID,
		Name:               req.Name,
		Description:        req.Description,
		MetricQuery:        req.MetricQuery,
		Algorithm:          req.Algorithm,
		Sensitivity:        req.Sensitivity,
		WindowSize:         req.WindowSize,
		MinValue:           req.MinValue,
		MaxValue:           req.MaxValue,
		Enabled:            req.Enabled,
		EvalInterval:       req.EvalInterval,
		AlertThreshold:     req.AlertThreshold,
		AlertOnRecovery:    req.AlertOnRecovery,
		NotificationChannels: req.NotificationChannels,
	}

	if rule.Sensitivity == 0 {
		rule.Sensitivity = 0.95
	}
	if rule.WindowSize == 0 {
		rule.WindowSize = 100
	}
	if rule.EvalInterval == 0 {
		rule.EvalInterval = 300
	}
	if rule.AlertThreshold == 0 {
		rule.AlertThreshold = 0.8
	}

	if err := h.db.Create(&rule).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create anomaly detection rule")
		return
	}

	respondWithJSON(w, http.StatusCreated, rule)
}

// ListAnomalyRules lists all anomaly detection rules
func (h *AIAnalysisHandler) ListAnomalyRules(w http.ResponseWriter, r *http.Request) {
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
	query := h.db.Model(&model.AnomalyDetectionRule{}).Where("user_id = ?", userUUID)

	// Apply filters
	if clusterID := r.URL.Query().Get("clusterId"); clusterID != "" {
		clusterUUID, err := uuid.Parse(clusterID)
		if err == nil {
			query = query.Where("cluster_id = ?", clusterUUID)
		}
	}
	if dataSourceID := r.URL.Query().Get("dataSourceId"); dataSourceID != "" {
		dataSourceUUID, err := uuid.Parse(dataSourceID)
		if err == nil {
			query = query.Where("data_source_id = ?", dataSourceUUID)
		}
	}
	if algorithm := r.URL.Query().Get("algorithm"); algorithm != "" {
		query = query.Where("algorithm = ?", algorithm)
	}
	if enabled := r.URL.Query().Get("enabled"); enabled != "" {
		query = query.Where("enabled = ?", enabled == "true")
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch rules
	var rules []model.AnomalyDetectionRule
	offset := (page - 1) * pageSize
	if err := query.Preload("Cluster").Preload("DataSource").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&rules).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch anomaly detection rules")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       rules,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetAnomalyRule gets a specific anomaly detection rule
func (h *AIAnalysisHandler) GetAnomalyRule(w http.ResponseWriter, r *http.Request) {
	// Extract rule ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid rule ID")
		return
	}

	ruleID := parts[4]
	ruleUUID, err := uuid.Parse(ruleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid rule ID format")
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

	// Fetch rule
	var rule model.AnomalyDetectionRule
	if err := h.db.Preload("Cluster").Preload("DataSource").Where("id = ? AND user_id = ?", ruleUUID, userUUID).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Anomaly detection rule not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch anomaly detection rule")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, rule)
}

// UpdateAnomalyRule updates an anomaly detection rule
func (h *AIAnalysisHandler) UpdateAnomalyRule(w http.ResponseWriter, r *http.Request) {
	// Extract rule ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid rule ID")
		return
	}

	ruleID := parts[4]
	ruleUUID, err := uuid.Parse(ruleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid rule ID format")
		return
	}

	var req model.UpdateAnomalyDetectionRuleRequest
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

	// Fetch rule
	var rule model.AnomalyDetectionRule
	if err := h.db.Where("id = ? AND user_id = ?", ruleUUID, userUUID).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Anomaly detection rule not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch anomaly detection rule")
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
	if req.MetricQuery != nil {
		updates["metric_query"] = *req.MetricQuery
	}
	if req.Algorithm != nil {
		updates["algorithm"] = *req.Algorithm
	}
	if req.Sensitivity != nil {
		updates["sensitivity"] = *req.Sensitivity
	}
	if req.WindowSize != nil {
		updates["window_size"] = *req.WindowSize
	}
	if req.MinValue != nil {
		updates["min_value"] = *req.MinValue
	}
	if req.MaxValue != nil {
		updates["max_value"] = *req.MaxValue
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.EvalInterval != nil {
		updates["eval_interval"] = *req.EvalInterval
	}
	if req.AlertThreshold != nil {
		updates["alert_threshold"] = *req.AlertThreshold
	}
	if req.AlertOnRecovery != nil {
		updates["alert_on_recovery"] = *req.AlertOnRecovery
	}
	if req.NotificationChannels != nil {
		updates["notification_channels"] = *req.NotificationChannels
	}

	if err := h.db.Model(&rule).Updates(updates).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update anomaly detection rule")
		return
	}

	// Fetch updated rule
	h.db.Preload("Cluster").Preload("DataSource").First(&rule, ruleUUID)

	respondWithJSON(w, http.StatusOK, rule)
}

// DeleteAnomalyRule deletes an anomaly detection rule
func (h *AIAnalysisHandler) DeleteAnomalyRule(w http.ResponseWriter, r *http.Request) {
	// Extract rule ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid rule ID")
		return
	}

	ruleID := parts[4]
	ruleUUID, err := uuid.Parse(ruleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid rule ID format")
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

	// Fetch rule
	var rule model.AnomalyDetectionRule
	if err := h.db.Where("id = ? AND user_id = ?", ruleUUID, userUUID).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Anomaly detection rule not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch anomaly detection rule")
		}
		return
	}

	// Delete rule
	if err := h.db.Delete(&rule).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete anomaly detection rule")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Anomaly detection rule deleted successfully",
	})
}

// ExecuteAnomalyDetection executes anomaly detection for a rule
func (h *AIAnalysisHandler) ExecuteAnomalyDetection(w http.ResponseWriter, r *http.Request) {
	var req model.ExecuteAnomalyDetectionRequest
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

	// Fetch rule
	var rule model.AnomalyDetectionRule
	if err := h.db.Where("id = ? AND user_id = ?", req.RuleID, userUUID).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Anomaly detection rule not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch anomaly detection rule")
		}
		return
	}

	startTime := time.Now()

	// TODO: Implement actual anomaly detection
	// This would:
	// 1. Query metrics from Prometheus based on rule.MetricQuery
	// 2. Apply the specified algorithm (STL, isolation forest, etc.)
	// 3. Detect anomalies based on sensitivity threshold
	// 4. Create anomaly events for detected anomalies

	duration := time.Since(startTime).Milliseconds()

	// Update rule statistics
	now := time.Now()
	h.db.Model(&rule).Updates(map[string]interface{}{
		"last_eval_at": &now,
		"total_evaluations": rule.TotalEvaluations + 1,
	})

	// Simulate response for now
	response := model.ExecuteAnomalyDetectionResponse{
		Anomalies:    []model.AnomalyEvent{},
		AnomalyCount: 0,
		EvaluatedAt:  now,
		Duration:     duration,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// ============== Anomaly Events ==============

// ListAnomalyEvents lists all anomaly events
func (h *AIAnalysisHandler) ListAnomalyEvents(w http.ResponseWriter, r *http.Request) {
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
	query := h.db.Model(&model.AnomalyEvent{}).Where("user_id = ?", userUUID)

	// Apply filters
	if clusterID := r.URL.Query().Get("clusterId"); clusterID != "" {
		clusterUUID, err := uuid.Parse(clusterID)
		if err == nil {
			query = query.Where("cluster_id = ?", clusterUUID)
		}
	}
	if ruleID := r.URL.Query().Get("ruleId"); ruleID != "" {
		ruleUUID, err := uuid.Parse(ruleID)
		if err == nil {
			query = query.Where("rule_id = ?", ruleUUID)
		}
	}
	if severity := r.URL.Query().Get("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch events
	var events []model.AnomalyEvent
	offset := (page - 1) * pageSize
	if err := query.Preload("Rule").Preload("Cluster").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&events).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch anomaly events")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       events,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// ============== LLM Conversations ==============

// CreateLLMConversation creates a new LLM conversation
func (h *AIAnalysisHandler) CreateLLMConversation(w http.ResponseWriter, r *http.Request) {
	var req model.CreateLLMConversationRequest
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

	// Create conversation
	conversation := model.LLMConversation{
		UserID:       userUUID,
		ClusterID:    req.ClusterID,
		Title:        req.Title,
		Model:        req.Model,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		SystemPrompt: req.SystemPrompt,
	}

	if conversation.Temperature == 0 {
		conversation.Temperature = 0.7
	}
	if conversation.MaxTokens == 0 {
		conversation.MaxTokens = 2000
	}

	if err := h.db.Create(&conversation).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create LLM conversation")
		return
	}

	respondWithJSON(w, http.StatusCreated, conversation)
}

// ListLLMConversations lists all LLM conversations
func (h *AIAnalysisHandler) ListLLMConversations(w http.ResponseWriter, r *http.Request) {
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
	query := h.db.Model(&model.LLMConversation{}).Where("user_id = ?", userUUID)

	// Count total
	var total int64
	query.Count(&total)

	// Fetch conversations
	var conversations []model.LLMConversation
	offset := (page - 1) * pageSize
	if err := query.Preload("Cluster").Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&conversations).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch LLM conversations")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       conversations,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetLLMConversation gets a specific LLM conversation
func (h *AIAnalysisHandler) GetLLMConversation(w http.ResponseWriter, r *http.Request) {
	// Extract conversation ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid conversation ID")
		return
	}

	conversationID := parts[4]
	conversationUUID, err := uuid.Parse(conversationID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid conversation ID format")
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

	// Fetch conversation with messages
	var conversation model.LLMConversation
	if err := h.db.Preload("Cluster").Preload("Messages").Where("id = ? AND user_id = ?", conversationUUID, userUUID).First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Conversation not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch conversation")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, conversation)
}

// DeleteLLMConversation deletes an LLM conversation
func (h *AIAnalysisHandler) DeleteLLMConversation(w http.ResponseWriter, r *http.Request) {
	// Extract conversation ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid conversation ID")
		return
	}

	conversationID := parts[4]
	conversationUUID, err := uuid.Parse(conversationID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid conversation ID format")
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

	// Fetch conversation
	var conversation model.LLMConversation
	if err := h.db.Where("id = ? AND user_id = ?", conversationUUID, userUUID).First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Conversation not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch conversation")
		}
		return
	}

	// Delete messages first
	h.db.Where("conversation_id = ?", conversationUUID).Delete(&model.LLMMessage{})

	// Delete conversation
	if err := h.db.Delete(&conversation).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete conversation")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Conversation deleted successfully",
	})
}

// SendLLMMessage sends a message in an LLM conversation
func (h *AIAnalysisHandler) SendLLMMessage(w http.ResponseWriter, r *http.Request) {
	// Extract conversation ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid conversation ID")
		return
	}

	conversationID := parts[4]
	conversationUUID, err := uuid.Parse(conversationID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid conversation ID format")
		return
	}

	var req model.SendLLMMessageRequest
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

	// Fetch conversation
	var conversation model.LLMConversation
	if err := h.db.Where("id = ? AND user_id = ?", conversationUUID, userUUID).First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Conversation not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch conversation")
		}
		return
	}

	// Create user message
	userMessage := model.LLMMessage{
		ConversationID: conversationUUID,
		Role:           "user",
		Content:        req.Content,
	}

	if err := h.db.Create(&userMessage).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create message")
		return
	}

	// TODO: Implement actual LLM API call
	// This would:
	// 1. Get conversation history
	// 2. Call LLM API (OpenAI, Anthropic, etc.)
	// 3. Parse response and create assistant message

	startTime := time.Now()
	duration := time.Since(startTime).Milliseconds()

	// Create assistant message (simulated for now)
	assistantMessage := model.LLMMessage{
		ConversationID: conversationUUID,
		Role:           "assistant",
		Content:        "This is a simulated response. The LLM integration is not yet implemented. Please configure your LLM API credentials in the settings.",
		TokensUsed:     50,
	}

	if err := h.db.Create(&assistantMessage).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create assistant message")
		return
	}

	response := model.SendLLMMessageResponse{
		MessageID: assistantMessage.ID.String(),
		Content:   assistantMessage.Content,
		TokensUsed: assistantMessage.TokensUsed,
	}

	respondWithJSON(w, http.StatusOK, response)
}
