// Package model provides data models for process management
package model

import (
	"time"

	"github.com/google/uuid"
)

// ProcessStatus represents the status of a process
type ProcessStatus string

const (
	ProcessStatusRunning   ProcessStatus = "running"
	ProcessStatusSleeping  ProcessStatus = "sleeping"
	ProcessStatusStopped   ProcessStatus = "stopped"
	ProcessStatusZombie    ProcessStatus = "zombie"
	ProcessStatusUnknown   ProcessStatus = "unknown"
)

// ProcessInfo represents information about a running process
type ProcessInfo struct {
	PID         int32             `json:"pid"`
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	User        string            `json:"user"`
	Status      ProcessStatus     `json:"status"`
	CPUPercent  float64           `json:"cpuPercent"`
	MemoryBytes int64             `json:"memoryBytes"`
	MemoryMB    float64           `json:"memoryMB"`
	StartTime   time.Time         `json:"startTime"`
	RunTime     string            `json:"runTime"`
	Terminal    string            `json:"terminal"`
}

// ListProcessesRequest represents a request to list processes
type ListProcessesRequest struct {
	HostID   uuid.UUID `json:"hostId"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Key      string    `json:"key"`
}

// ListProcessesResponse represents a response listing processes
type ListProcessesResponse struct {
	Processes []ProcessInfo `json:"processes"`
	Count     int           `json:"count"`
}

// GetProcessRequest represents a request to get process details
type GetProcessRequest struct {
	HostID   uuid.UUID `json:"hostId"`
	PID      int32     `json:"pid"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Key      string    `json:"key"`
}

// KillProcessRequest represents a request to kill a process
type KillProcessRequest struct {
	HostID   uuid.UUID `json:"hostId"`
	PID      int32     `json:"pid"`
	Signal   int32     `json:"signal"`   // Signal number (default: 9 for SIGKILL)
	Username string    `json:"username"`
	Password string    `json:"password"`
	Key      string    `json:"key"`
}

// ExecuteCommandRequest represents a request to execute a command
type ExecuteCommandRequest struct {
	HostID     uuid.UUID `json:"hostId"`
	Command    string    `json:"command"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Key        string    `json:"key"`
	Timeout    int32     `json:"timeout"`    // Timeout in seconds
	WorkingDir string    `json:"workingDir"` // Optional working directory
}

// ExecuteCommandResponse represents a response from command execution
type ExecuteCommandResponse struct {
	ExitCode int32  `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Duration string `json:"duration"`
}

// ProcessExecution represents a command execution record
type ProcessExecution struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	HostID      uuid.UUID `json:"hostId" gorm:"type:uuid;not null;index"`
	UserID      uuid.UUID `json:"userId" gorm:"type:uuid;not null;index"`
	Command     string    `json:"command" gorm:"type:varchar(1000);not null"`
	ExitCode    *int32    `json:"exitCode" gorm:"type:int"`
	Stdout      string    `json:"stdout" gorm:"type:text"`
	Stderr      string    `json:"stderr" gorm:"type:text"`
	Duration    int64     `json:"duration" gorm:"type:bigint"` // Duration in milliseconds
	Status      string    `json:"status" gorm:"type:varchar(20);not null"` // running, completed, failed
	StartedAt   *time.Time `json:"startedAt" gorm:"type:timestamp"`
	CompletedAt *time.Time `json:"completedAt" gorm:"type:timestamp"`
	CreatedAt   time.Time `json:"createdAt" gorm:"type:timestamp;autoCreateTime"`
}

// TableName specifies the table name for ProcessExecution
func (ProcessExecution) TableName() string {
	return "process_executions"
}
