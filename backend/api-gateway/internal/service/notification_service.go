// Package service provides notification management services
package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NotificationService handles notification creation and delivery
type NotificationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *gorm.DB, logger *zap.Logger) *NotificationService {
	return &NotificationService{
		db:     db,
		logger: logger,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(userID uuid.UUID, notifType model.NotificationType, title, message string, priority model.NotificationPriority) (*model.Notification, error) {
	notification := &model.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Message:   message,
		Priority:  priority,
		Status:    model.NotificationStatusPending,
		Read:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.db.Create(notification).Error
	if err != nil {
		return nil, err
	}

	// Trigger async delivery
	go s.deliverNotification(notification)

	return notification, nil
}

// CreateBulkNotification creates notifications for multiple users
func (s *NotificationService) CreateBulkNotification(userIDs []uuid.UUID, notifType model.NotificationType, title, message string, priority model.NotificationPriority) ([]*model.Notification, error) {
	var notifications []*model.Notification
	now := time.Now()

	for _, userID := range userIDs {
		notification := &model.Notification{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      notifType,
			Title:     title,
			Message:   message,
			Priority:  priority,
			Status:    model.NotificationStatusPending,
			Read:      false,
			CreatedAt: now,
			UpdatedAt: now,
		}
		notifications = append(notifications, notification)
	}

	err := s.db.Create(&notifications).Error
	if err != nil {
		return nil, err
	}

	// Trigger async delivery for all
	for _, notif := range notifications {
		go s.deliverNotification(notif)
	}

	return notifications, nil
}

// GetUserNotifications retrieves notifications for a user
func (s *NotificationService) GetUserNotifications(userID uuid.UUID, limit int, unreadOnly bool) ([]model.Notification, error) {
	var notifications []model.Notification
	query := s.db.Where("user_id = ?", userID)

	if unreadOnly {
		query = query.Where("read = ?", false)
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&notifications).Error
	return notifications, err
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&model.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&model.Notification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"read":     true,
			"read_at":  &now,
			"updated_at": time.Now(),
		}).Error
}

// MarkAllAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllAsRead(userID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&model.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Updates(map[string]interface{}{
			"read":       true,
			"read_at":    &now,
			"updated_at": time.Now(),
		}).Error
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(notificationID uuid.UUID) error {
	return s.db.Delete(&model.Notification{}, "id = ?", notificationID).Error
}

// GetNotificationPreference retrieves user notification preferences
func (s *NotificationService) GetNotificationPreference(userID uuid.UUID) (*model.NotificationPreference, error) {
	var pref model.NotificationPreference
	err := s.db.Where("user_id = ?", userID).First(&pref).Error
	if err == gorm.ErrRecordNotFound {
		// Create default preference
		pref = model.NotificationPreference{
			ID:           uuid.New(),
			UserID:       userID,
			EmailEnabled: true,
			WebEnabled:   true,
			PushEnabled:  false,
			AlertTypes:   []model.NotificationType{model.NotificationTypeAlert, model.NotificationTypeSystem, model.NotificationTypeSecurity},
			MinPriority:  model.NotificationPriorityMedium,
			Timezone:     "UTC",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err = s.db.Create(&pref).Error
	}
	return &pref, err
}

// UpdateNotificationPreference updates user notification preferences
func (s *NotificationService) UpdateNotificationPreference(pref *model.NotificationPreference) error {
	pref.UpdatedAt = time.Now()
	return s.db.Save(pref).Error
}

// deliverNotification handles the actual delivery of a notification
func (s *NotificationService) deliverNotification(notification *model.Notification) {
	// Get user preferences
	pref, err := s.GetNotificationPreference(notification.UserID)
	if err != nil {
		s.logger.Error("Failed to get notification preference",
			zap.String("notification_id", notification.ID.String()),
			zap.Error(err))
		return
	}

	// Check if notification should be delivered based on preferences
	if !s.shouldDeliver(notification, pref) {
		s.logger.Info("Notification skipped due to preferences",
			zap.String("notification_id", notification.ID.String()))
		return
	}

	// Mark as sent
	notification.Status = model.NotificationStatusSent
	notification.UpdatedAt = time.Now()
	s.db.Save(notification)

	// Log delivery
	s.logDelivery(notification.ID, "web", model.NotificationStatusDelivered)

	// TODO: Implement actual delivery channels (email, push, etc.)
	s.logger.Info("Notification delivered",
		zap.String("notification_id", notification.ID.String()),
		zap.String("user_id", notification.UserID.String()))
}

// shouldDeliver checks if a notification should be delivered based on preferences
func (s *NotificationService) shouldDeliver(notification *model.Notification, pref *model.NotificationPreference) bool {
	// Check priority threshold
	if !s.priorityMatches(notification.Priority, pref.MinPriority) {
		return false
	}

	// Check alert type preferences
	if len(pref.AlertTypes) > 0 {
		typeFound := false
		for _, t := range pref.AlertTypes {
			if t == notification.Type {
				typeFound = true
				break
			}
		}
		if !typeFound {
			return false
		}
	}

	// Check quiet hours
	if s.isInQuietHours(pref) {
		return false
	}

	return true
}

// priorityMatches checks if notification priority matches minimum threshold
func (s *NotificationService) priorityMatches(notificationMin, userMin model.NotificationPriority) bool {
	priorityOrder := map[model.NotificationPriority]int{
		model.NotificationPriorityLow:      0,
		model.NotificationPriorityMedium:   1,
		model.NotificationPriorityHigh:     2,
		model.NotificationPriorityCritical: 3,
	}
	return priorityOrder[notificationMin] >= priorityOrder[userMin]
}

// isInQuietHours checks if current time is within quiet hours
func (s *NotificationService) isInQuietHours(pref *model.NotificationPreference) bool {
	if pref.QuietHoursStart == "" || pref.QuietHoursEnd == "" {
		return false
	}

	// TODO: Implement timezone-aware quiet hours check
	return false
}

// logDelivery logs notification delivery
func (s *NotificationService) logDelivery(notificationID uuid.UUID, channel string, status model.NotificationStatus) {
	log := &model.NotificationDeliveryLog{
		ID:             uuid.New(),
		NotificationID: notificationID,
		Channel:        channel,
		Status:         status,
		CreatedAt:      time.Now(),
	}
	s.db.Create(log)
}

// GetNotificationsByType retrieves notifications by type for a user
func (s *NotificationService) GetNotificationsByType(userID uuid.UUID, notifType model.NotificationType, limit int) ([]model.Notification, error) {
	var notifications []model.Notification
	err := s.db.Where("user_id = ? AND type = ?", userID, notifType).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

// GetNotificationsByPriority retrieves notifications by priority for a user
func (s *NotificationService) GetNotificationsByPriority(userID uuid.UUID, priority model.NotificationPriority, limit int) ([]model.Notification, error) {
	var notifications []model.Notification
	err := s.db.Where("user_id = ? AND priority = ?", userID, priority).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

// CleanupOldNotifications removes notifications older than specified days
func (s *NotificationService) CleanupOldNotifications(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return s.db.Where("created_at < ? AND read = ?", cutoff, true).Delete(&model.Notification{}).Error
}

// GetNotificationStats returns statistics about notifications for a user
func (s *NotificationService) GetNotificationStats(userID uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total notifications
	var total int64
	s.db.Model(&model.Notification{}).Where("user_id = ?", userID).Count(&total)
	stats["total"] = total

	// Unread notifications
	var unread int64
	s.db.Model(&model.Notification{}).Where("user_id = ? AND read = ?", userID, false).Count(&unread)
	stats["unread"] = unread

	// By type
	for _, notifType := range []model.NotificationType{
		model.NotificationTypeAlert,
		model.NotificationTypeSystem,
		model.NotificationTypeTask,
		model.NotificationTypeSecurity,
		model.NotificationTypeInfo,
	} {
		var count int64
		s.db.Model(&model.Notification{}).Where("user_id = ? AND type = ?", userID, notifType).Count(&count)
		stats[string(notifType)] = count
	}

	return stats, nil
}

// CreateAlertNotification creates a notification from an alert
func (s *NotificationService) CreateAlertNotification(userID uuid.UUID, alertTitle, alertMessage string, severity string, alertID string) error {
	priority := model.NotificationPriorityLow
	switch severity {
	case "critical":
		priority = model.NotificationPriorityCritical
	case "warning":
		priority = model.NotificationPriorityHigh
	case "info":
		priority = model.NotificationPriorityMedium
	}

	_, err := s.CreateNotification(
		userID,
		model.NotificationTypeAlert,
		alertTitle,
		alertMessage,
		priority,
	)

	if err == nil {
		// Add action URL to metadata
		// notification.ActionURL = fmt.Sprintf("/alerts/%s", alertID)
	}

	return err
}

// CreateTaskNotification creates a notification for task status updates
func (s *NotificationService) CreateTaskNotification(userID uuid.UUID, taskName, status string, taskID string) error {
	title := fmt.Sprintf("Task %s", status)
	message := fmt.Sprintf("Task '%s' is now %s", taskName, status)

	priority := model.NotificationPriorityLow
	if status == "failed" {
		priority = model.NotificationPriorityHigh
	}

	_, err := s.CreateNotification(
		userID,
		model.NotificationTypeTask,
		title,
		message,
		priority,
	)

	return err
}

// CreateSystemNotification creates a system-wide notification for all users
func (s *NotificationService) CreateSystemNotification(title, message string, priority model.NotificationPriority) error {
	// Get all users
	var userIDs []uuid.UUID
	err := s.db.Model(&model.Notification{}).Distinct("user_id").Pluck("user_id", &userIDs).Error
	if err != nil {
		return err
	}

	if len(userIDs) == 0 {
		return nil // No users yet
	}

	_, err = s.CreateBulkNotification(userIDs, model.NotificationTypeSystem, title, message, priority)
	return err
}
