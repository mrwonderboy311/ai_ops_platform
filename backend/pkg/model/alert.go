// Package model provides data models for alerts and events
package model

import (
	"time"

	"github.com/google/uuid"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusPending  AlertStatus = "pending"
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
	AlertStatusSilenced AlertStatus = "silenced"
)

// AlertRule represents an alert rule configuration
type AlertRule struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID      uuid.UUID    `json:"userId" gorm:"type:uuid;not null;index"`
	Name        string       `json:"name" gorm:"type:varchar(255);not null"`
	Description string       `json:"description" gorm:"type:text"`
	Enabled     bool         `json:"enabled" gorm:"type:boolean;default:true"`
	// Target configuration
	TargetType  string `json:"targetType" gorm:"type:varchar(50);not null"` // host, cluster, node, pod
	TargetID    string `json:"targetId" gorm:"type:varchar(255)"`
	// Rule conditions
	MetricType  string  `json:"metricType" gorm:"type:varchar(100);not null"` // cpu_usage, memory_usage, disk_usage, pod_status, node_status
	Operator    string  `json:"operator" gorm:"type:varchar(20);not null"`    // >, <, >=, <=, ==, !=
	Threshold   float64 `json:"threshold" gorm:"type:decimal(10,2);not null"`
	Duration    int32   `json:"duration" gorm:"type:int;default:300"`           // seconds
	// Notification settings
	Severity    AlertSeverity `json:"severity" gorm:"type:varchar(20);not null"`
	NotifyEmail bool          `json:"notifyEmail" gorm:"type:boolean;default:false"`
	NotifyWebhook bool        `json:"notifyWebhook" gorm:"type:boolean;default:false"`
	WebhookURL  string        `json:"webhookUrl" gorm:"type:varchar(500)"`
	// Scheduling
	SilencedUntil  *time.Time `json:"silencedUntil"`
	LastEvaluatedAt *time.Time `json:"lastEvaluatedAt"`
	CreatedAt       time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
}

// Alert represents a triggered alert
type Alert struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	RuleID      uuid.UUID      `json:"ruleId" gorm:"type:uuid;not null;index"`
	UserID      uuid.UUID      `json:"userId" gorm:"type:uuid;not null;index"`
	ClusterID   *uuid.UUID     `json:"clusterId,omitempty" gorm:"type:uuid;index"`
	HostID      *uuid.UUID     `json:"hostId,omitempty" gorm:"type:uuid;index"`
	Status      AlertStatus    `json:"status" gorm:"type:varchar(20);not null;index"`
	Severity    AlertSeverity `json:"severity" gorm:"type:varchar(20);not null"`
	Title       string        `json:"title" gorm:"type:varchar(500);not null"`
	Description string        `json:"description" gorm:"type:text"`
	// Current value when alert fired
	Value       float64  `json:"value" gorm:"type:decimal(10,2)"`
	Threshold   float64  `json:"threshold" gorm:"type:decimal(10,2)"`
	// Timestamps
	StartedAt   time.Time `json:"startedAt" gorm:"not null;index"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"not null"`
	ResolvedAt  *time.Time `json:"resolvedAt,omitempty"`
	SilencedUntil *time.Time `json:"silencedUntil,omitempty"`
	// Labels for filtering
	Labels      string   `json:"labels" gorm:"type:text"` // JSON
	Annotations string   `json:"annotations" gorm:"type:text"` // JSON
}

// Event represents a system event
type Event struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID *uuid.UUID `json:"clusterId,omitempty" gorm:"type:uuid;index"`
	HostID    *uuid.UUID `json:"hostId,omitempty" gorm:"type:uuid;index"`
	Type      string    `json:"type" gorm:"type:varchar(100);not null;index"` // host_up, host_down, cluster_joined, cluster_left, etc.
	Severity  string    `json:"severity" gorm:"type:varchar(20);not null"`    // info, warning, error
	Title     string    `json:"title" gorm:"type:varchar(500);not null"`
	Message   string    `json:"message" gorm:"type:text"`
	Metadata  string    `json:"metadata" gorm:"type:text"` // JSON
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime;index"`
}

// AlertNotification represents a notification sent for an alert
type AlertNotification struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	AlertID   uuid.UUID `json:"alertId" gorm:"type:uuid;not null;index"`
	Type      string    `json:"type" gorm:"type:varchar(50);not null"` // email, webhook
	Status    string    `json:"status" gorm:"type:varchar(20);not null"` // pending, sent, failed
	Recipient string    `json:"recipient" gorm:"type:varchar(255)"`
	Content   string    `json:"content" gorm:"type:text"`
	Error     string    `json:"error" gorm:"type:text"`
	SentAt    *time.Time `json:"sentAt"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// AlertStatistics represents alert statistics
type AlertStatistics struct {
	TotalAlerts    int64 `json:"totalAlerts"`
	FiringAlerts   int64 `json:"firingAlerts"`
	ResolvedAlerts int64 `json:"resolvedAlerts"`
	CriticalAlerts int64 `json:"criticalAlerts"`
	WarningAlerts  int64 `json:"warningAlerts"`
}
