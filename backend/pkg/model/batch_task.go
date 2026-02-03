// Package model provides data models for batch task management
package model

import (
	"time"

	"github.com/google/uuid"
)

// BatchTaskStatus represents the status of a batch task
type BatchTaskStatus string

const (
	BatchTaskStatusPending   BatchTaskStatus = "pending"
	BatchTaskStatusRunning   BatchTaskStatus = "running"
	BatchTaskStatusCompleted BatchTaskStatus = "completed"
	BatchTaskStatusFailed    BatchTaskStatus = "failed"
	BatchTaskStatusCancelled BatchTaskStatus = "cancelled"
)

// BatchTaskType represents the type of batch task
type BatchTaskType string

const (
	BatchTaskTypeCommand  BatchTaskType = "command"
	BatchTaskTypeScript   BatchTaskType = "script"
	BatchTaskTypeFileOp   BatchTaskType = "file_op"
)

// TaskExecutionStrategy defines how tasks are executed across hosts
type TaskExecutionStrategy string

const (
	StrategyParallel TaskExecutionStrategy = "parallel" // Execute on all hosts simultaneously
	StrategySerial   TaskExecutionStrategy = "serial"   // Execute one host at a time
	StrategyRolling  TaskExecutionStrategy = "rolling"  // Execute with a percentage at a time
)

// BatchTask represents a batch task to be executed on multiple hosts
type BatchTask struct {
	ID            uuid.UUID             `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID        uuid.UUID             `json:"userId" gorm:"type:uuid;not null;index"`
	Name          string                `json:"name" gorm:"type:varchar(255);not null"`
	Description   string                `json:"description" gorm:"type:text"`
	Type          BatchTaskType         `json:"type" gorm:"type:varchar(20);not null"`
	Status        BatchTaskStatus       `json:"status" gorm:"type:varchar(20);not null;index"`
	Strategy      TaskExecutionStrategy `json:"strategy" gorm:"type:varchar(20);not null"`
	Command       string                `json:"command" gorm:"type:text"`           // Command or script content
	Script        string                `json:"script" gorm:"type:text"`            // Embedded script
	Timeout       int32                 `json:"timeout" gorm:"type:int;default:60"` // Timeout per host in seconds
	MaxRetries    int32                 `json:"maxRetries" gorm:"type:int;default:0"`
	Parallelism   int32                 `json:"parallelism" gorm:"type:int;default:0"` // 0 = all at once
	TotalHosts    int32                 `json:"totalHosts" gorm:"type:int;default:0"`
	CompletedHosts int32                `json:"completedHosts" gorm:"type:int;default:0"`
	FailedHosts   int32                 `json:"failedHosts" gorm:"type:int;default:0"`
	StartedAt     *time.Time            `json:"startedAt" gorm:"type:timestamp"`
	CompletedAt   *time.Time            `json:"completedAt" gorm:"type:timestamp"`
	CreatedAt     time.Time             `json:"createdAt" gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt     time.Time             `json:"updatedAt" gorm:"type:timestamp;autoUpdateTime"`
}

// BatchTaskHost represents a batch task execution on a specific host
type BatchTaskHost struct {
	ID              uuid.UUID          `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BatchTaskID     uuid.UUID          `json:"batchTaskId" gorm:"type:uuid;not null;index"`
	HostID          uuid.UUID          `json:"hostId" gorm:"type:uuid;not null;index"`
	Status          BatchTaskStatus    `json:"status" gorm:"type:varchar(20);not null"`
	ExitCode        *int32             `json:"exitCode" gorm:"type:int"`
	Stdout          string             `json:"stdout" gorm:"type:text"`
	Stderr          string             `json:"stderr" gorm:"type:text"`
	Duration        int64              `json:"duration" gorm:"type:bigint"` // Duration in milliseconds
	ErrorMessage    string             `json:"errorMessage" gorm:"type:text"`
	RetryCount      int32              `json:"retryCount" gorm:"type:int;default:0"`
	StartedAt       *time.Time         `json:"startedAt" gorm:"type:timestamp"`
	CompletedAt     *time.Time         `json:"completedAt" gorm:"type:timestamp"`
	CreatedAt       time.Time          `json:"createdAt" gorm:"type:timestamp;autoCreateTime"`
	// Relations
	BatchTask       *BatchTask         `json:"batchTask,omitempty" gorm:"foreignKey:BatchTaskID"`
	Host            *Host              `json:"host,omitempty" gorm:"foreignKey:HostID"`
}

// TableName specifies the table name for BatchTask
func (BatchTask) TableName() string {
	return "batch_tasks"
}

// TableName specifies the table name for BatchTaskHost
func (BatchTaskHost) TableName() string {
	return "batch_task_hosts"
}

// CreateBatchTaskRequest represents a request to create a batch task
type CreateBatchTaskRequest struct {
	Name          string                `json:"name" binding:"required"`
	Description   string                `json:"description"`
	Type          BatchTaskType         `json:"type" binding:"required"`
	Strategy      TaskExecutionStrategy `json:"strategy" binding:"required"`
	Command       string                `json:"command"`
	Script        string                `json:"script"`
	Timeout       int32                 `json:"timeout"`
	MaxRetries    int32                 `json:"maxRetries"`
	Parallelism   int32                 `json:"parallelism"`
	HostIDs       []uuid.UUID           `json:"hostIds" binding:"required"`
}

// ExecuteBatchTaskRequest represents a request to execute a batch task
type ExecuteBatchTaskRequest struct {
	TaskID        uuid.UUID   `json:"taskId" binding:"required"`
	HostIDs       []uuid.UUID `json:"hostIds"`
}

// CancelBatchTaskRequest represents a request to cancel a batch task
type CancelBatchTaskRequest struct {
	TaskID        uuid.UUID   `json:"taskId" binding:"required"`
}

// BatchTaskResponse represents a response with batch task details
type BatchTaskResponse struct {
	*BatchTask
	Hosts          []BatchTaskHost `json:"hosts,omitempty"`
	Progress       float64         `json:"progress"` // 0-100
}

// ListBatchTasksRequest represents a request to list batch tasks
type ListBatchTasksRequest struct {
	Status   *BatchTaskStatus `form:"status"`
	Type     *BatchTaskType   `form:"type"`
	Page     int              `form:"page" binding:"min=1"`
	PageSize int              `form:"pageSize" binding:"min=1,max=100"`
}

// BatchTaskSummary represents a summary of batch task execution
type BatchTaskSummary struct {
	TotalTasks     int64            `json:"totalTasks"`
	RunningTasks   int64            `json:"runningTasks"`
	CompletedTasks int64            `json:"completedTasks"`
	FailedTasks    int64            `json:"failedTasks"`
	PendingTasks   int64            `json:"pendingTasks"`
}
