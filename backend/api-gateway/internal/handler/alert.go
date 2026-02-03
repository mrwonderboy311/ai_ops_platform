// Package handler provides HTTP handlers for alert management
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// AlertHandler handles alert operations
type AlertHandler struct {
	db *gorm.DB
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(db *gorm.DB) *AlertHandler {
	return &AlertHandler{db: db}
}

// ListAlertRules handles alert rule list requests
func (h *AlertHandler) ListAlertRules(w http.ResponseWriter, r *http.Request) {
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
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build query
	query := h.db.Model(&model.AlertRule{}).Where("user_id = ?", userID)

	if enabled := r.URL.Query().Get("enabled"); enabled != "" {
		if enabled == "true" {
			query = query.Where("enabled = ?", true)
		} else if enabled == "false" {
			query = query.Where("enabled = ?", false)
		}
	}

	if targetType := r.URL.Query().Get("targetType"); targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get rules with pagination
	var rules []model.AlertRule
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&rules).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve alert rules")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"rules":    rules,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateAlertRule handles alert rule creation requests
func (h *AlertHandler) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	var req model.AlertRule
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

	// Set user ID and generate ID
	req.ID = uuid.New()
	req.UserID = userID
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if err := h.db.Create(&req).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create alert rule")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"data": req,
	})
}

// GetAlertRule handles alert rule retrieval requests
func (h *AlertHandler) GetAlertRule(w http.ResponseWriter, r *http.Request) {
	// Get rule ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 4 || pathParts[3] != "alert-rules" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	ruleID, err := uuid.Parse(pathParts[4])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_RULE_ID", "Invalid rule ID")
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

	// Get rule
	var rule model.AlertRule
	if err := h.db.Where("id = ? AND user_id = ?", ruleID, userID).First(&rule).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert rule not found")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": rule,
	})
}

// UpdateAlertRule handles alert rule update requests
func (h *AlertHandler) UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Get rule ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 4 || pathParts[3] != "alert-rules" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	ruleID, err := uuid.Parse(pathParts[4])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_RULE_ID", "Invalid rule ID")
		return
	}

	var req model.AlertRule
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

	// Get rule
	var rule model.AlertRule
	if err := h.db.Where("id = ? AND user_id = ?", ruleID, userID).First(&rule).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert rule not found")
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.TargetType != "" {
		updates["target_type"] = req.TargetType
	}
	if req.TargetID != "" {
		updates["target_id"] = req.TargetID
	}
	if req.MetricType != "" {
		updates["metric_type"] = req.MetricType
	}
	if req.Operator != "" {
		updates["operator"] = req.Operator
	}
	if req.Threshold != 0 {
		updates["threshold"] = req.Threshold
	}
	if req.Duration != 0 {
		updates["duration"] = req.Duration
	}
	if req.Severity != "" {
		updates["severity"] = req.Severity
	}
	updates["notify_email"] = req.NotifyEmail
	updates["notify_webhook"] = req.NotifyWebhook
	if req.WebhookURL != "" {
		updates["webhook_url"] = req.WebhookURL
	}
	updates["updated_at"] = time.Now()

	if err := h.db.Model(&rule).Updates(updates).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update alert rule")
		return
	}

	// Get updated rule
	h.db.Where("id = ?", ruleID).First(&rule)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": rule,
	})
}

// DeleteAlertRule handles alert rule deletion requests
func (h *AlertHandler) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Get rule ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 4 || pathParts[3] != "alert-rules" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	ruleID, err := uuid.Parse(pathParts[4])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_RULE_ID", "Invalid rule ID")
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

	// Verify rule ownership
	var rule model.AlertRule
	if err := h.db.Where("id = ? AND user_id = ?", ruleID, userID).First(&rule).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert rule not found")
		return
	}

	// Delete related alerts
	h.db.Where("rule_id = ?", ruleID).Delete(&model.Alert{})

	// Delete rule
	if err := h.db.Delete(&rule).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete alert rule")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Alert rule deleted successfully",
		"ruleId":  ruleID,
	})
}

// ListAlerts handles alert list requests
func (h *AlertHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
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
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build query
	query := h.db.Model(&model.Alert{}).Where("user_id = ?", userID)

	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if severity := r.URL.Query().Get("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get alerts with pagination
	var alerts []model.Alert
	offset := (page - 1) * pageSize
	if err := query.Order("started_at DESC").Limit(pageSize).Offset(offset).Find(&alerts).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve alerts")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"alerts":   alerts,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// GetAlertStatistics handles alert statistics requests
func (h *AlertHandler) GetAlertStatistics(w http.ResponseWriter, r *http.Request) {
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

	var stats model.AlertStatistics

	// Count total alerts
	h.db.Model(&model.Alert{}).Where("user_id = ?", userID).Count(&stats.TotalAlerts)

	// Count firing alerts
	h.db.Model(&model.Alert{}).Where("user_id = ? AND status = ?", userID, model.AlertStatusFiring).Count(&stats.FiringAlerts)

	// Count resolved alerts
	h.db.Model(&model.Alert{}).Where("user_id = ? AND status = ?", userID, model.AlertStatusResolved).Count(&stats.ResolvedAlerts)

	// Count critical alerts
	h.db.Model(&model.Alert{}).Where("user_id = ? AND severity = ?", userID, model.AlertSeverityCritical).Count(&stats.CriticalAlerts)

	// Count warning alerts
	h.db.Model(&model.Alert{}).Where("user_id = ? AND severity = ?", userID, model.AlertSeverityWarning).Count(&stats.WarningAlerts)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": stats,
	})
}

// SilenceAlert silences an alert
func (h *AlertHandler) SilenceAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Get alert ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 5 || pathParts[4] != "silence" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	alertID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_ALERT_ID", "Invalid alert ID")
		return
	}

	var req struct {
		Duration string `json:"duration"` // e.g., "1h", "24h", "7d"
	}
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

	// Verify alert ownership
	var alert model.Alert
	if err := h.db.Where("id = ? AND user_id = ?", alertID, userID).First(&alert).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert not found")
		return
	}

	// Parse duration
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_DURATION", "Invalid duration format")
		return
	}

	silencedUntil := time.Now().Add(duration)

	// Update alert
	alert.Status = model.AlertStatusSilenced
	alert.SilencedUntil = &silencedUntil
	alert.UpdatedAt = time.Now()

	if err := h.db.Save(&alert).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to silence alert")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": alert,
	})
}

// ListEvents handles event list requests
func (h *AlertHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
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
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build query
	query := h.db.Model(&model.Event{})

	if eventType := r.URL.Query().Get("type"); eventType != "" {
		query = query.Where("type = ?", eventType)
	}
	if severity := r.URL.Query().Get("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get events with pagination
	var events []model.Event
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&events).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve events")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"events":   events,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}
