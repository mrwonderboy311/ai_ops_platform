// Package model provides data models for notifications
package model

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeAlert    NotificationType = "alert"
	NotificationTypeSystem   NotificationType = "system"
	NotificationTypeTask     NotificationType = "task"
	NotificationTypeSecurity NotificationType = "security"
	NotificationTypeInfo     NotificationType = "info"
)

// NotificationPriority represents the priority level
type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityMedium   NotificationPriority = "medium"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityCritical NotificationPriority = "critical"
)

// NotificationStatus represents the delivery status
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusDelivered NotificationStatus = "delivered"
)

// Notification represents a user notification
type Notification struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID      uuid.UUID `gorm:"type:uuid;index"`
	Type        NotificationType
	Title       string
	Message     string
	Priority    NotificationPriority
	Status      NotificationStatus
	Read        bool      `gorm:"default:false"`
	ReadAt      *time.Time
	ActionURL   string
	ActionLabel string
	Metadata    map[string]string `gorm:"serializer:json"`
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NotificationPreference represents user notification preferences
type NotificationPreference struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID          uuid.UUID `gorm:"type:uuid;index"`
	EmailEnabled    bool `gorm:"default:true"`
	WebEnabled      bool `gorm:"default:true"`
	PushEnabled     bool `gorm:"default:false"`
	AlertTypes      []NotificationType `gorm:"type:text[]"`
	MinPriority     NotificationPriority `gorm:"default:low"`
	QuietHoursStart string // Format: "HH:MM"
	QuietHoursEnd   string // Format: "HH:MM"`
	Timezone        string `gorm:"default:'UTC'"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NotificationDeliveryLog represents notification delivery logs
type NotificationDeliveryLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	NotificationID uuid.UUID `gorm:"type:uuid;index"`
	Channel        string // email, web, push, webhook
	Status         NotificationStatus
	ErrorCode      string
	ErrorMessage   string
	RetriedCount   int `gorm:"default:0"`
	SentAt         *time.Time
	DeliveredAt    *time.Time
	CreatedAt      time.Time
}

// NotificationTemplate represents notification templates
type NotificationTemplate struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string `gorm:"uniqueIndex"`
	Type        NotificationType
	Title       string
	Body        string
	Variables   []string `gorm:"type:text[]"`
	Enabled     bool `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName specifies the table name for Notification
func (Notification) TableName() string {
	return "notifications"
}

// TableName specifies the table name for NotificationPreference
func (NotificationPreference) TableName() string {
	return "notification_preferences"
}

// TableName specifies the table name for NotificationDeliveryLog
func (NotificationDeliveryLog) TableName() string {
	return "notification_delivery_logs"
}

// TableName specifies the table name for NotificationTemplate
func (NotificationTemplate) TableName() string {
	return "notification_templates"
}
