// Package model provides data models for OpenTelemetry Collector management
package model

import (
	"time"

	"github.com/google/uuid"
)

// CollectorStatus represents the status of an OpenTelemetry collector
type CollectorStatus string

const (
	CollectorStatusDeploying   CollectorStatus = "deploying"
	CollectorStatusRunning     CollectorStatus = "running"
	CollectorStatusStopped     CollectorStatus = "stopped"
	CollectorStatusError       CollectorStatus = "error"
	CollectorStatusPending     CollectorStatus = "pending"
)

// CollectorType represents the type of collector
type CollectorType string

const (
	CollectorTypeMetrics CollectorType = "metrics" // Collects metrics only
	CollectorTypeLogs    CollectorType = "logs"    // Collects logs only
	CollectorTypeTraces  CollectorType = "traces"  // Collects traces only
	CollectorTypeAll     CollectorType = "all"     // Collects everything
)

// OtelCollector represents an OpenTelemetry collector deployment
type OtelCollector struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID      uuid.UUID       `json:"userId" gorm:"type:uuid;not null;index"`
	ClusterID   uuid.UUID       `json:"clusterId" gorm:"type:uuid;not null;index"`
	Name        string          `json:"name" gorm:"type:varchar(255);not null"`
	Namespace   string          `json:"namespace" gorm:"type:varchar(255);not null;default:'observability'"`
	Type        CollectorType  `json:"type" gorm:"type:varchar(20);not null;default:'all'"`
	Status      CollectorStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	Version     string          `json:"version" gorm:"type:varchar(50)"`

	// Configuration
	Config      string          `json:"-" gorm:"type:text"` // YAML configuration (encrypted if sensitive)
	Replicas    int32           `json:"replicas" gorm:"type:int;default:1"`
	Resources   string          `json:"resources" gorm:"type:text"` // JSON for resource limits/requests

	// Endpoints
	MetricsEndpoint    string `json:"metricsEndpoint" gorm:"type:varchar(500)"`
	LogsEndpoint       string `json:"logsEndpoint" gorm:"type:varchar(500)"`
	TracesEndpoint     string `json:"tracesEndpoint" gorm:"type:varchar(500)"`

	// Status info
	PodNames           string `json:"podNames" gorm:"type:text"` // JSON array of running pods
	LastHealthCheck    *time.Time `json:"lastHealthCheck" gorm:"type:timestamp"`
	ErrorMessage       string `json:"errorMessage" gorm:"type:text"`

	CreatedAt   time.Time `json:"createdAt" gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"type:timestamp;autoUpdateTime"`
	// Relations
	Cluster     *K8sCluster `json:"cluster,omitempty" gorm:"foreigner:ClusterID"`
	User        *User       `json:"user,omitempty" gorm:"foreigner:UserID"`
}

// TableName specifies the table name for OtelCollector
func (OtelCollector) TableName() string {
	return "otel_collectors"
}

// CollectorConfig represents OpenTelemetry collector configuration
type CollectorConfig struct {
	// Receivers configuration
	Receivers map[string]interface{} `json:"receivers"`

	// Processors configuration
	Processors map[string]interface{} `json:"processors"`

	// Exporters configuration
	Exporters map[string]interface{} `json:"exporters"`

	// Extensions configuration
	Extensions map[string]interface{} `json:"extensions"`

	// Service configuration
	Service struct {
		Extensions []string `json:"extensions"`
		Pipelines  map[string]interface{} `json:"pipelines"`
	} `json:"service"`
}

// CreateCollectorRequest represents a request to create a collector
type CreateCollectorRequest struct {
	ClusterID        uuid.UUID      `json:"clusterId" binding:"required"`
	Name             string         `json:"name" binding:"required"`
	Namespace        string         `json:"namespace" binding:"required"`
	Type             CollectorType  `json:"type" binding:"required"`
	Config           string         `json:"config"`
	Replicas         int32          `json:"replicas"`
	Resources        string         `json:"resources"`
	MetricsEndpoint string         `json:"metricsEndpoint"`
	LogsEndpoint     string         `json:"logsEndpoint"`
	TracesEndpoint   string         `json:"tracesEndpoint"`
}

// UpdateCollectorRequest represents a request to update a collector
type UpdateCollectorRequest struct {
	Config           string `json:"config"`
	Replicas         int32  `json:"replicas"`
	Resources        string `json:"resources"`
	MetricsEndpoint string `json:"metricsEndpoint"`
	LogsEndpoint     string `json:"logsEndpoint"`
	TracesEndpoint   string `json:"tracesEndpoint"`
}

// ListCollectorsRequest represents a request to list collectors
type ListCollectorsRequest struct {
	ClusterID *uuid.UUID       `form:"clusterId"`
	Namespace *string          `form:"namespace"`
	Type      *CollectorType   `form:"type"`
	Status    *CollectorStatus `form:"status"`
	Page      int              `form:"page" binding:"min=1"`
	PageSize  int              `form:"pageSize" binding:"min=1,max=100"`
}

// CollectorDeploymentStatus represents the deployment status
type CollectorDeploymentStatus struct {
	Status        CollectorStatus `json:"status"`
	PodNames      []string        `json:"podNames"`
	Replicas      int32            `json:"replicas"`
	ReadyReplicas int32            `json:"readyReplicas"`
	MetricsURL    string           `json:"metricsUrl"`
	ErrorMessage  string           `json:"errorMessage"`
}
