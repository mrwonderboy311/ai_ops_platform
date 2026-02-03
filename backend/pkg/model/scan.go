// Package model provides data models for the application
package model

import (
	"time"

	"github.com/google/uuid"
)

// ScanTaskStatus represents the status of a scan task
type ScanTaskStatus string

const (
	ScanTaskStatusRunning   ScanTaskStatus = "running"
	ScanTaskStatusCompleted ScanTaskStatus = "completed"
	ScanTaskStatusFailed    ScanTaskStatus = "failed"
	ScanTaskStatusCancelled ScanTaskStatus = "cancelled"
)

// ScanTask represents an IP range scan task
type ScanTask struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID       `gorm:"type:uuid;not null" json:"userId"`
	IPRange         string          `gorm:"not null" json:"ipRange"`
	Ports           []int           `gorm:"type:integer[]" json:"ports"`
	TimeoutSeconds  int             `gorm:"default:5" json:"timeoutSeconds"`
	Status          ScanTaskStatus  `gorm:"size:50;default:'running'" json:"status"`
	EstimatedHosts  int             `json:"estimatedHosts"`
	DiscoveredHosts int             `gorm:"default:0" json:"discoveredHosts"`
	StartedAt       time.Time       `json:"startedAt"`
	CompletedAt     *time.Time      `json:"completedAt,omitempty"`
	ErrorMessage    string          `json:"errorMessage,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// DiscoveredHost represents a host found during scanning
type DiscoveredHost struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ScanTaskID  uuid.UUID `gorm:"type:uuid;not null" json:"scanTaskId"`
	IPAddress  string    `gorm:"not null" json:"ipAddress"`
	Port       int       `gorm:"not null" json:"port"`
	Hostname    string    `json:"hostname"`
	OSType      string    `json:"osType"`
	OSVersion  string    `json:"osVersion"`
	Status      string    `json:"status"` // "success", "timeout", "error"
	CreatedAt   time.Time `json:"createdAt"`
}
