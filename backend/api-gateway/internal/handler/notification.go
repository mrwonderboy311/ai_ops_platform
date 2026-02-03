// Package handler provides HTTP handlers for notification management
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/api-gateway/internal/service"
	"github.com/wangjialin/myops/pkg/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NotificationHandler handles notification requests
type NotificationHandler struct {
	db                 *gorm.DB
	notificationService *service.NotificationService
	logger             *zap.Logger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(db *gorm.DB, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		db:                 db,
		notificationService: service.NewNotificationService(db, logger),
		logger:             logger,
	}
}

// GetNotifications handles notification list retrieval
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
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
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	unreadOnly := r.URL.Query().Get("unreadOnly") == "true"

	notifications, err := h.notificationService.GetUserNotifications(userID, limit, unreadOnly)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve notifications")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"notifications": notifications,
			"count":         len(notifications),
		},
	})
}

// GetUnreadCount handles unread count retrieval
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
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

	count, err := h.notificationService.GetUnreadCount(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve unread count")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"unreadCount": count,
		},
	})
}

// MarkAsRead handles marking a notification as read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
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

	// Get notification ID from path
	path := r.URL.Path
	notificationID := extractIDFromPath(path, "/api/v1/notifications/")
	if notificationID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Notification ID is required")
		return
	}

	id, err := uuid.Parse(notificationID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid notification ID")
		return
	}

	// Verify ownership
	var notification model.Notification
	err = h.db.Where("id = ? AND user_id = ?", id, userID).First(&notification).Error
	if err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Notification not found")
		return
	}

	err = h.notificationService.MarkAsRead(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark notification as read")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"success": true,
		},
	})
}

// MarkAllAsRead handles marking all notifications as read
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
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

	err := h.notificationService.MarkAllAsRead(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark all notifications as read")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"success": true,
		},
	})
}

// DeleteNotification handles notification deletion
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
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

	// Get notification ID from path
	path := r.URL.Path
	notificationID := extractIDFromPath(path, "/api/v1/notifications/")
	if notificationID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Notification ID is required")
		return
	}

	id, err := uuid.Parse(notificationID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid notification ID")
		return
	}

	// Verify ownership
	var notification model.Notification
	err = h.db.Where("id = ? AND user_id = ?", id, userID).First(&notification).Error
	if err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Notification not found")
		return
	}

	err = h.notificationService.DeleteNotification(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete notification")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"success": true,
		},
	})
}

// GetNotificationPreference handles notification preference retrieval
func (h *NotificationHandler) GetNotificationPreference(w http.ResponseWriter, r *http.Request) {
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

	pref, err := h.notificationService.GetNotificationPreference(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve notification preferences")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": pref,
	})
}

// UpdateNotificationPreference handles notification preference update
func (h *NotificationHandler) UpdateNotificationPreference(w http.ResponseWriter, r *http.Request) {
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

	var pref model.NotificationPreference
	if err := json.NewDecoder(r.Body).Decode(&pref); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Ensure user ID matches
	pref.UserID = userID

	err := h.notificationService.UpdateNotificationPreference(&pref)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update notification preferences")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": pref,
	})
}

// GetNotificationStats handles notification statistics retrieval
func (h *NotificationHandler) GetNotificationStats(w http.ResponseWriter, r *http.Request) {
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

	stats, err := h.notificationService.GetNotificationStats(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve notification statistics")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": stats,
	})
}

// CreateNotification handles notification creation (admin/internal endpoint)
func (h *NotificationHandler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string                   `json:"userId"`
		Type     model.NotificationType   `json:"type"`
		Title    string                   `json:"title"`
		Message  string                   `json:"message"`
		Priority model.NotificationPriority `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID")
		return
	}

	notification, err := h.notificationService.CreateNotification(
		userID,
		req.Type,
		req.Title,
		req.Message,
		req.Priority,
	)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create notification")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"data": notification,
	})
}

// helper function to extract ID from path
func extractIDFromPath(path, prefix string) string {
	// Simple extraction - assumes ID is the last segment
	parts := strings.Split(path[len(prefix):], "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
