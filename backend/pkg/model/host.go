// Package model provides data models for the application
package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// LabelMap represents a JSONB map for host labels
type LabelMap map[string]string

// Scan implements sql.Scanner for LabelMap
func (lm *LabelMap) Scan(value interface{}) error {
	if value == nil {
		*lm = make(LabelMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, lm)
}

// Value implements driver.Valuer for LabelMap
func (lm LabelMap) Value() (driver.Value, error) {
	if lm == nil {
		return "{}", nil
	}
	return json.Marshal(lm)
}

// HostStatus represents the status of a host
type HostStatus string

const (
	HostStatusPending  HostStatus = "pending"  // awaiting approval
	HostStatusApproved HostStatus = "approved" // approved and active
	HostStatusRejected HostStatus = "rejected" // rejected during registration
	HostStatusOffline  HostStatus = "offline"  // offline/not reachable
	HostStatusOnline   HostStatus = "online"   // online and reachable
)

// Host represents a managed host/server
type Host struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Hostname    string         `gorm:"not null" json:"hostname"`
	IPAddress   string         `gorm:"type:inet;not null" json:"ipAddress"`
	Port        int            `gorm:"default:22" json:"port"`
	Status      HostStatus     `gorm:"size:50;default:'pending'" json:"status"`
	OSType      string         `gorm:"size:100" json:"osType"`
	OSVersion   string         `gorm:"size:100" json:"osVersion"`
	CPUCores    *int           `json:"cpuCores"`
	MemoryGB    *int           `json:"memoryGB"`
	DiskGB      *int64         `json:"diskGB"`
	Labels      LabelMap       `gorm:"type:jsonb;default:'{}'" json:"labels"`
	Tags        pq.StringArray `gorm:"type:text[];default:'{}'" json:"tags"`
	ClusterID   *uuid.UUID     `gorm:"type:uuid" json:"clusterId"`
	RegisteredBy *uuid.UUID    `gorm:"type:uuid" json:"registeredBy"`
	ApprovedBy  *uuid.UUID     `gorm:"type:uuid" json:"approvedBy"`
	ApprovedAt  *time.Time     `json:"approvedAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`

	// Eager loaded associations
	RegisteredByUser *User  `gorm:"foreignKey:RegisteredBy" json:"registeredByUser,omitempty"`
	ApprovedByUser   *User  `gorm:"foreignKey:ApprovedBy" json:"approvedByUser,omitempty"`
	Cluster          *Cluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// TableName specifies the table name for Host model
func (Host) TableName() string {
	return "hosts"
}

// Cluster represents a Kubernetes cluster (for future Epic 4)
type Cluster struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"not null;unique" json:"name"`
	APIEndpoint string    `json:"apiEndpoint"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TableName specifies the table name for Cluster model
func (Cluster) TableName() string {
	return "clusters"
}
