// Package model provides data models for audit logging
package model

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID      uuid.UUID `json:"userId" gorm:"type:uuid;not null;index:idx_user_action_time"`
	Username    string    `json:"username" gorm:"type:varchar(255);not null"`
	Action      string    `json:"action" gorm:"type:varchar(100);not null;index:idx_user_action_time"`
	Resource    string    `json:"resource" gorm:"type:varchar(255);not null"`  // e.g., "hosts", "clusters", "users"
	ResourceID  string    `json:"resourceId" gorm:"type:varchar(255)"`
	// Request details
	Method      string    `json:"method" gorm:"type:varchar(10)"`      // GET, POST, PUT, DELETE, etc.
	Path        string    `json:"path" gorm:"type:varchar(500)"`
	IPAddress   string    `json:"ipAddress" gorm:"type:varchar(50)"`
	UserAgent    string    `json:"userAgent" gorm:"type:text"`
	// Response details
	StatusCode  int       `json:"statusCode" gorm:"type:int"`
	ErrorMsg    string    `json:"errorMsg" gorm:"type:text"`
	// Changes tracking
	OldValue    string    `json:"oldValue" gorm:"type:text"`     // JSON of previous state
	NewValue    string    `json:"newValue" gorm:"type:text"`     // JSON of new state
	// Metadata
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime;index:idx_user_action_time"`
}

// OperationType represents the type of operation
type OperationType string

const (
	OperationCreate OperationType = "create"
	OperationRead   OperationType = "read"
	OperationUpdate OperationType = "update"
	OperationDelete OperationType = "delete"
	OperationLogin  OperationType = "login"
	OperationLogout OperationType = "logout"
)

// ResourceType represents the type of resource
type ResourceType string

const (
	ResourceHost       ResourceType = "hosts"
	ResourceCluster   ResourceType = "clusters"
	ResourceUser      ResourceType = "users"
	ResourceBatchTask  ResourceType = "batch-tasks"
	ResourceAlertRule  ResourceType = "alert-rules"
	ResourceAlert      ResourceType = "alerts"
)

// AuditLogFilter represents filters for querying audit logs
type AuditLogFilter struct {
	UserID     *uuid.UUID `json:"userId"`
	Username   *string    `json:"username"`
	Action     *string    `json:"action"`
	Resource   *string    `json:"resource"`
	ResourceID *string    `json:"resourceId"`
	StartTime  *time.Time `json:"startTime"`
	EndTime    *time.Time `json:"endTime"`
	IPAddress  *string    `json:"ipAddress"`
}

// AuditLogSummary represents a summary of audit activities
type AuditLogSummary struct {
	TotalOperations    int64     `json:"totalOperations"`
	UserActivity      int64     `json:"userActivity"`
	FailedOperations  int64     `json:"failedOperations"`
	OperationsByType  map[string]int64 `json:"operationsByType"`
	OperationsByResource map[string]int64 `json:"operationsByResource"`
	TopUsers           []UserActivityStats `json:"topUsers"`
	TopResources      []ResourceActivityStats `json:"topResources"`
}

// UserActivityStats represents user activity statistics
type UserActivityStats struct {
	Username       string `json:"username"`
	OperationCount int64  `json:"operationCount"`
}

// ResourceActivityStats represents resource activity statistics
type ResourceActivityStats struct {
	Resource      string `json:"resource"`
	OperationCount int64 `json:"operationCount"`
}
