// Package model provides data models for Prometheus integration
package model

import (
	"time"

	"github.com/google/uuid"
)

// PrometheusDataSource represents a Prometheus data source configuration
type PrometheusDataSource struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID     uuid.UUID `gorm:"type:uuid;not null;index:idx_prometheus_user_id" json:"userId"`
	ClusterID  *uuid.UUID `gorm:"type:uuid;index:idx_prometheus_cluster_id" json:"clusterId,omitempty"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	URL        string    `gorm:"size:2048;not null" json:"url"`
	Username   string    `gorm:"size:255" json:"username,omitempty"`
	Password   string    `gorm:"size:255" json:"-"` // Never expose in JSON
	Status     string    `gorm:"size:50;default:DSStatusActive" json:"status"`

	// TLS configuration
	InsecureSkipTLS bool   `gorm:"default:false" json:"insecureSkipTLS"`
	CACert          string `gorm:"type:text" json:"caCert,omitempty"`
	ClientCert      string `gorm:"type:text" json:"clientCert,omitempty"`
	ClientKey       string `gorm:"type:text" json:"-"` // Never expose in JSON

	// Headers (stored as JSON string)
	Headers string `gorm:"type:text" json:"headers,omitempty"`

	// Test results
	LastTestAt     *time.Time `json:"lastTestAt,omitempty"`
	LastTestStatus string     `gorm:"size:50" json:"lastTestStatus,omitempty"`
	LastTestError  string     `gorm:"type:text" json:"lastTestError,omitempty"`

	// Metrics
	QueryCount       int       `gorm:"default:0" json:"queryCount"`
	LastQueriedAt    *time.Time `json:"lastQueriedAt,omitempty"`
	AverageQueryTime float64   `gorm:"default:0" json:"averageQueryTime"`

	// Relationships
	Cluster *K8sCluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// PrometheusAlertRule represents a Prometheus alerting rule
type PrometheusAlertRule struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_prometheus_alert_user_id" json:"userId"`
	DataSourceID uuid.UUID `gorm:"type:uuid;not null;index:idx_prometheus_alert_datasource_id" json:"dataSourceId"`
	ClusterID    *uuid.UUID `gorm:"type:uuid;index:idx_prometheus_alert_cluster_id" json:"clusterId,omitempty"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	Expression   string    `gorm:"type:text;not null" json:"expression"`
	Duration     int       `gorm:"not null;default:300" json:"duration"` // seconds
	Severity     string    `gorm:"size:50;default:AlertSeverityWarning" json:"severity"`
	Summary      string    `gorm:"type:text" json:"summary,omitempty"`
	Description  string    `gorm:"type:text" json:"description,omitempty"`
	Labels       string    `gorm:"type:text" json:"labels,omitempty"` // JSON string
	Annotations  string    `gorm:"type:text" json:"annotations,omitempty"` // JSON string
	Enabled      bool      `gorm:"default:true" json:"enabled"`

	// Sync status
	Synced       bool      `gorm:"default:false" json:"synced"`
	SyncedAt     *time.Time `json:"syncedAt,omitempty"`
	SyncError    string    `gorm:"type:text" json:"syncError,omitempty"`

	// Statistics
	TriggerCount int        `gorm:"default:0" json:"triggerCount"`
	LastTriggeredAt *time.Time `json:"lastTriggeredAt,omitempty"`

	// Relationships
	DataSource *PrometheusDataSource `gorm:"foreignKey:DataSourceID" json:"dataSource,omitempty"`
	Cluster    *K8sCluster          `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// PrometheusQuery represents a query execution record
type PrometheusQuery struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_prometheus_query_user_id" json:"userId"`
	DataSourceID uuid.UUID `gorm:"type:uuid;not null;index:idx_prometheus_query_datasource_id" json:"dataSourceId"`
	Query        string    `gorm:"type:text;not null" json:"query"`
	QueryType    string    `gorm:"size:50;not null" json:"queryType"` // instant, range
	StartTime    string    `gorm:"size:100" json:"startTime,omitempty"`
	EndTime      string    `gorm:"size:100" json:"endTime,omitempty"`
	Step         string    `gorm:"size:50" json:"step,omitempty"`

	Duration      int64   `gorm:"default:0" json:"duration"` // milliseconds
	Success       bool    `gorm:"default:true" json:"success"`
	ErrorMessage  string  `gorm:"type:text" json:"errorMessage,omitempty"`
	ResultCount   int     `gorm:"default:0" json:"resultCount"`

	// Relationships
	DataSource *PrometheusDataSource `gorm:"foreignKey:DataSourceID" json:"dataSource,omitempty"`
}

// PrometheusDashboard represents a custom dashboard configuration
type PrometheusDashboard struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_prometheus_dashboard_user_id" json:"userId"`
	ClusterID    *uuid.UUID `gorm:"type:uuid;index:idx_prometheus_dashboard_cluster_id" json:"clusterId,omitempty"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	Description  string    `gorm:"type:text" json:"description,omitempty"`
	Tags         string    `gorm:"type:text" json:"tags,omitempty"` // JSON array
	Config       string    `gorm:"type:text;not null" json:"config"` // JSON: panels, layout, etc.
	IsPublic     bool      `gorm:"default:false" json:"isPublic"`
	Starred      bool      `gorm:"default:false" json:"starred"`
	RefreshRate  int       `gorm:"default:30" json:"refreshRate"` // seconds

	// Relationships
	Cluster *K8sCluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// DataSource status constants
const (
	DSStatusActive   = "active"
	DSStatusInactive = "inactive"
	DSStatusError    = "error"
)

// Alert severity constants
const (
	AlertSeverityCritical = "critical"
	AlertSeverityWarning  = "warning"
	AlertSeverityInfo     = "info"
)

// CreatePrometheusDataSourceRequest represents a request to create a data source
type CreatePrometheusDataSourceRequest struct {
	ClusterID       *uuid.UUID `json:"clusterId,omitempty"`
	Name            string     `json:"name" binding:"required"`
	URL             string     `json:"url" binding:"required"`
	Username        string     `json:"username,omitempty"`
	Password        string     `json:"password,omitempty"`
	InsecureSkipTLS bool       `json:"insecureSkipTLS,omitempty"`
	CACert          string     `json:"caCert,omitempty"`
	ClientCert      string     `json:"clientCert,omitempty"`
	ClientKey       string     `json:"clientKey,omitempty"`
	Headers         string     `json:"headers,omitempty"`
}

// UpdatePrometheusDataSourceRequest represents a request to update a data source
type UpdatePrometheusDataSourceRequest struct {
	Name            *string  `json:"name,omitempty"`
	URL             *string  `json:"url,omitempty"`
	Username        *string  `json:"username,omitempty"`
	Password        *string  `json:"password,omitempty"`
	InsecureSkipTLS *bool    `json:"insecureSkipTLS,omitempty"`
	CACert          *string  `json:"caCert,omitempty"`
	ClientCert      *string  `json:"clientCert,omitempty"`
	ClientKey       *string  `json:"clientKey,omitempty"`
	Headers         *string  `json:"headers,omitempty"`
	Status          *string  `json:"status,omitempty"`
}

// TestPrometheusDataSourceRequest represents a request to test a data source
type TestPrometheusDataSourceRequest struct {
	URL             string `json:"url" binding:"required"`
	Username        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
	InsecureSkipTLS bool   `json:"insecureSkipTLS,omitempty"`
	CACert          string `json:"caCert,omitempty"`
	ClientCert      string `json:"clientCert,omitempty"`
	ClientKey       string `json:"clientKey,omitempty"`
}

// TestPrometheusDataSourceResponse represents the response from testing a data source
type TestPrometheusDataSourceResponse struct {
	Success    bool   `json:"success"`
	Version    string `json:"version,omitempty"`
	Message    string `json:"message"`
	Error      string `json:"error,omitempty"`
	Duration   int64  `json:"duration"` // milliseconds
}

// CreatePrometheusAlertRuleRequest represents a request to create an alert rule
type CreatePrometheusAlertRuleRequest struct {
	DataSourceID uuid.UUID  `json:"dataSourceId" binding:"required"`
	ClusterID    *uuid.UUID `json:"clusterId,omitempty"`
	Name         string     `json:"name" binding:"required"`
	Expression   string     `json:"expression" binding:"required"`
	Duration     int        `json:"duration" binding:"required,min=1"`
	Severity     string     `json:"severity" binding:"required,oneof=critical warning info"`
	Summary      string     `json:"summary,omitempty"`
	Description  string     `json:"description,omitempty"`
	Labels       string     `json:"labels,omitempty"`
	Annotations  string     `json:"annotations,omitempty"`
}

// UpdatePrometheusAlertRuleRequest represents a request to update an alert rule
type UpdatePrometheusAlertRuleRequest struct {
	Name        *string `json:"name,omitempty"`
	Expression  *string `json:"expression,omitempty"`
	Duration    *int    `json:"duration,omitempty"`
	Severity    *string `json:"severity,omitempty"`
	Summary     *string `json:"summary,omitempty"`
	Description *string `json:"description,omitempty"`
	Labels      *string `json:"labels,omitempty"`
	Annotations *string `json:"annotations,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

// PrometheusQueryRequest represents a request to query Prometheus
type PrometheusQueryRequest struct {
	Query     string `json:"query" binding:"required"`
	QueryType string `json:"queryType" binding:"required,oneof=instant range"`
	StartTime string `json:"startTime,omitempty"` // RFC3339 or relative (e.g., "1h ago")
	EndTime   string `json:"endTime,omitempty"`   // RFC3339 or "now"
	Step      string `json:"step,omitempty"`      // e.g., "15s", "1m"
}

// PrometheusQueryResponse represents the response from a Prometheus query
type PrometheusQueryResponse struct {
	Status   string        `json:"status"`
	Data     []PrometheusSeries `json:"data,omitempty"`
	Error    string        `json:"error,omitempty"`
	Duration int64         `json:"duration"` // milliseconds
}

// PrometheusSeries represents a single time series
type PrometheusSeries struct {
	Metric map[string]string `json:"metric"`
	Values []PrometheusValue  `json:"values,omitempty"`
	Value  *PrometheusValue   `json:"value,omitempty"`
}

// PrometheusValue represents a single value at a timestamp
type PrometheusValue struct {
	Timestamp float64 `json:"timestamp"`
	Value     string  `json:"value"`
}

// CreatePrometheusDashboardRequest represents a request to create a dashboard
type CreatePrometheusDashboardRequest struct {
	ClusterID   *uuid.UUID `json:"clusterId,omitempty"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description,omitempty"`
	Tags        string     `json:"tags,omitempty"`
	Config      string     `json:"config" binding:"required"`
	IsPublic    bool       `json:"isPublic,omitempty"`
	RefreshRate int        `json:"refreshRate,omitempty"`
}

// UpdatePrometheusDashboardRequest represents a request to update a dashboard
type UpdatePrometheusDashboardRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Tags        *string `json:"tags,omitempty"`
	Config      *string `json:"config,omitempty"`
	IsPublic    *bool   `json:"isPublic,omitempty"`
	RefreshRate *int    `json:"refreshRate,omitempty"`
	Starred     *bool   `json:"starred,omitempty"`
}
