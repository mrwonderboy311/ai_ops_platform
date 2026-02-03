// Package handler provides HTTP handlers for audit log management
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// AuditHandler handles audit log operations
type AuditHandler struct {
	db *gorm.DB
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(db *gorm.DB) *AuditHandler {
	return &AuditHandler{db: db}
}

// ListAuditLogs handles audit log list requests
func (h *AuditHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (admin only)
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

	// Parse filters
	var filters model.AuditLogFilter
	if username := r.URL.Query().Get("username"); username != "" {
		filters.Username = &username
	}
	if action := r.URL.Query().Get("action"); action != "" {
		filters.Action = &action
	}
	if resource := r.URL.Query().Get("resource"); resource != "" {
		filters.Resource = &resource
	}
	if resourceId := r.URL.Query().Get("resourceId"); resourceId != "" {
		filters.ResourceID = &resourceId
	}
	if startTime := r.URL.Query().Get("startTime"); startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			filters.StartTime = &t
		}
	}
	if endTime := r.URL.Query().Get("endTime"); endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			filters.EndTime = &t
		}
	}

	// Build query
	query := h.db.Model(&model.AuditLog{})

	if filters.Username != nil {
		query = query.Where("username = ?", *filters.Username)
	}
	if filters.Action != nil {
		query = query.Where("action = ?", *filters.Action)
	}
	if filters.Resource != nil {
		query = query.Where("resource = ?", *filters.Resource)
	}
	if filters.ResourceID != nil {
		query = query.Where("resource_id = ?", *filters.ResourceID)
	}
	if filters.StartTime != nil {
		query = query.Where("created_at >= ?", *filters.StartTime)
	}
	if filters.EndTime != nil {
		query = query.Where("created_at <= ?", *filters.EndTime)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get logs with pagination
	var logs []model.AuditLog
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&logs).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve audit logs")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"logs":     logs,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// GetAuditLogSummary handles audit log summary requests
func (h *AuditHandler) GetAuditLogSummary(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (admin only)
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

	// Parse time range (default: last 24 hours)
	duration := r.URL.Query().Get("duration")
	startTime := time.Now().Add(-24 * time.Hour)
	if duration != "" {
		if d, err := time.ParseDuration(duration); err == nil {
			startTime = time.Now().Add(-d)
		}
	}

	// Get total operations
	var totalOps int64
	h.db.Model(&model.AuditLog{}).Where("created_at >= ?", startTime).Count(&totalOps)

	// Get failed operations
	var failedOps int64
	h.db.Model(&model.AuditLog{}).Where("created_at >= ? AND status_code >= ?", startTime, 400).Count(&failedOps)

	// Get operations by type
	var opsByType []struct {
		Action string
		Count  int64
	}
	h.db.Model(&model.AuditLog{}).
		Select("action, count(*) as count").
		Where("created_at >= ?", startTime).
		Group("action").
		Scan(&opsByType)

	// Get operations by resource
	var opsByResource []struct {
		Resource string
		Count    int64
	}
	h.db.Model(&model.AuditLog{}).
		Select("resource, count(*) as count").
		Where("created_at >= ?", startTime).
		Group("resource").
		Order("count DESC").
		Limit(10).
		Scan(&opsByResource)

	// Get top users
	var topUsers []struct {
		Username string
		Count    int64
	}
	h.db.Model(&model.AuditLog{}).
		Select("username, count(*) as count").
		Where("created_at >= ?", startTime).
		Group("username").
		Order("count DESC").
		Limit(10).
		Scan(&topUsers)

	// Build operations by type map
	operationsByType := make(map[string]int64)
	for _, op := range opsByType {
		operationsByType[op.Action] = op.Count
	}

	// Build operations by resource map
	operationsByResource := make(map[string]int64)
	for _, op := range opsByResource {
		operationsByResource[op.Resource] = op.Count
	}

	// Build top users stats
	userActivityStats := make([]model.UserActivityStats, len(topUsers))
	for i, user := range topUsers {
		userActivityStats[i] = model.UserActivityStats{
			Username:       user.Username,
			OperationCount: user.Count,
		}
	}

	// Build top resources stats
	resourceActivityStats := make([]model.ResourceActivityStats, len(opsByResource))
	for i, res := range opsByResource {
		resourceActivityStats[i] = model.ResourceActivityStats{
			Resource:       res.Resource,
			OperationCount: res.Count,
		}
	}

	summary := &model.AuditLogSummary{
		TotalOperations:     totalOps,
		UserActivity:       int64(len(userActivityStats)),
		FailedOperations:   failedOps,
		OperationsByType:    operationsByType,
		OperationsByResource: operationsByResource,
		TopUsers:            userActivityStats,
		TopResources:       resourceActivityStats,
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": summary,
	})
}

// GetUserActivity handles user activity requests
func (h *AuditHandler) GetUserActivity(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (admin only)
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

	// Parse time range
	duration := r.URL.Query().Get("duration")
	startTime := time.Now().Add(-24 * time.Hour)
	if duration != "" {
		if d, err := time.ParseDuration(duration); err == nil {
			startTime = time.Now().Add(-d)
		}
	}

	// Parse user ID from URL if specified
	targetUserID := userID
	if userIDParam := r.URL.Query().Get("userId"); userIDParam != "" {
		if uid, err := uuid.Parse(userIDParam); err == nil {
			targetUserID = uid
		}
	}

	// Get user activity
	var activities []struct {
		Action     string
		Resource   string
		ResourceID string
		Count      int64
		LastSeenAt  time.Time
	}
	h.db.Model(&model.AuditLog{}).
		Select("action, resource, resource_id, count(*) as count, max(created_at) as last_seen_at").
		Where("user_id = ? AND created_at >= ?", targetUserID, startTime).
		Group("action, resource, resource_id").
		Order("last_seen_at DESC").
		Limit(50).
		Scan(&activities)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": activities,
	})
}

// GetResourceActivity handles resource activity requests
func (h *AuditHandler) GetResourceActivity(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (admin only)
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

	// Parse time range
	duration := r.URL.Query().Get("duration")
	startTime := time.Now().Add(-24 * time.Hour)
	if duration != "" {
		if d, err := time.ParseDuration(duration); err == nil {
			startTime = time.Now().Add(-d)
		}
	}

	// Get resource activity
	var activities []struct {
		Resource   string
		ResourceID string
		Action     string
		Count      int64
		LastSeenAt  time.Time
	}
	h.db.Model(&model.AuditLog{}).
		Select("resource, resource_id, action, count(*) as count, max(created_at) as last_seen_at").
		Where("created_at >= ?", startTime).
		Group("resource, resource_id, action").
		Order("last_seen_at DESC").
		Limit(50).
		Scan(&activities)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": activities,
	})
}
