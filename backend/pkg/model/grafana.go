// Package model provides data models for Grafana integration
package model

import (
	"time"

	"github.com/google/uuid"
)

// GrafanaInstance represents a Grafana instance configuration
type GrafanaInstance struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID     uuid.UUID `gorm:"type:uuid;not null;index:idx_grafana_user_id" json:"userId"`
	ClusterID  *uuid.UUID `gorm:"type:uuid;index:idx_grafana_cluster_id" json:"clusterId,omitempty"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	URL        string    `gorm:"size:2048;not null" json:"url"`
	APIKey     string    `gorm:"size:500" json:"-"` // Never expose in JSON
	Username   string    `gorm:"size:255" json:"username,omitempty"`
	Password   string    `gorm:"size:255" json:"-"` // Never expose in JSON
	Status     string    `gorm:"size:50;default:GrafanaStatusActive" json:"status"`

	// Service account configuration
	ServiceAccountID  string `gorm:"size:255" json:"serviceAccountId,omitempty"`
	ServiceAccountToken string `gorm:"size:500" json:"-"` // Never expose

	// Sync configuration
	AutoSync   bool      `gorm:"default:true" json:"autoSync"`
	SyncInterval int     `gorm:"default:300" json:"syncInterval"` // seconds
	LastSyncAt *time.Time `json:"lastSyncAt,omitempty"`
	SyncStatus string    `gorm:"size:50;default:SyncStatusPending" json:"syncStatus"`
	SyncError  string    `gorm:"type:text" json:"syncError,omitempty"`

	// Statistics
	DashboardCount int `gorm:"default:0" json:"dashboardCount"`
	DataSourceCount int `gorm:"default:0" json:"dataSourceCount"`

	// Relationships
	Cluster *K8sCluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// GrafanaDashboard represents a synced Grafana dashboard
type GrafanaDashboard struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_grafana_dashboard_user_id" json:"userId"`
	InstanceID   uuid.UUID `gorm:"type:uuid;not null;index:idx_grafana_dashboard_instance_id" json:"instanceId"`
	ClusterID    *uuid.UUID `gorm:"type:uuid;index:idx_grafana_dashboard_cluster_id" json:"clusterId,omitempty"`

	// Grafana properties
	GrafanaUID   string `gorm:"size:255;not null" json:"grafanaUid"`
	GrafanaID    int    `gorm:"not null" json:"grafanaId"`
	Title        string `gorm:"size:500;not null" json:"title"`
	Slug         string `gorm:"size:255" json:"slug,omitempty"`
	Tags         string `gorm:"type:text" json:"tags,omitempty"` // JSON array
	FolderTitle  string `gorm:"size:255" json:"folderTitle,omitempty"`
	FolderUID    string `gorm:"size:255" json:"folderUid,omitempty"`
	FolderID     int    `json:"folderId,omitempty"`

	// Dashboard configuration
	Config       string `gorm:"type:text;not null" json:"config"` // JSON: full dashboard model
	Version      int    `gorm:"default:1" json:"version"`
	IsStarred    bool   `gorm:"default:false" json:"isStarred"`

	// Sync status
	Synced       bool      `gorm:"default:false" json:"synced"`
	SyncedAt     *time.Time `json:"syncedAt,omitempty"`

	// Relationships
	Instance *GrafanaInstance `gorm:"foreignKey:InstanceID" json:"instance,omitempty"`
	Cluster  *K8sCluster     `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// GrafanaDataSource represents a synced Grafana data source
type GrafanaDataSource struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_grafana_ds_user_id" json:"userId"`
	InstanceID   uuid.UUID `gorm:"type:uuid;not null;index:idx_grafana_ds_instance_id" json:"instanceId"`

	// Grafana properties
	GrafanaUID   string `gorm:"size:255;not null" json:"grafanaUid"`
	GrafanaID    int    `gorm:"not null" json:"grafanaId"`
	Name         string `gorm:"size:255;not null" json:"name"`
	Type         string `gorm:"size:100;not null" json:"type"` // prometheus, loki, elasticsearch, etc.
	IsDefault    bool   `gorm:"default:false" json:"isDefault"`

	// Connection details
	URL          string `gorm:"size:2048" json:"url,omitempty"`
	Database     string `gorm:"size:255" json:"database,omitempty"`
	JSONData     string `gorm:"type:text" json:"jsonData,omitempty"` // JSON string

	// Health status
	HealthStatus string `gorm:"size:50" json:"healthStatus,omitempty"` // OK, ERROR

	// Relationships
	Instance *GrafanaInstance `gorm:"foreignKey:InstanceID" json:"instance,omitempty"`
}

// GrafanaFolder represents a Grafana folder
type GrafanaFolder struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_grafana_folder_user_id" json:"userId"`
	InstanceID   uuid.UUID `gorm:"type:uuid;not null;index:idx_grafana_folder_instance_id" json:"instanceId"`

	// Grafana properties
	GrafanaUID   string `gorm:"size:255;not null" json:"grafanaUid"`
	GrafanaID    int    `gorm:"not null" json:"grafanaId"`
	Title        string `gorm:"size:255;not null" json:"title"`

	// Sync status
	Synced       bool      `gorm:"default:false" json:"synced"`
	SyncedAt     *time.Time `json:"syncedAt,omitempty"`

	// Relationships
	Instance *GrafanaInstance `gorm:"foreignKey:InstanceID" json:"instance,omitempty"`
}

// Grafana status constants
const (
	GrafanaStatusActive   = "active"
	GrafanaStatusInactive = "inactive"
	GrafanaStatusError    = "error"
)

// Sync status constants
const (
	SyncStatusPending  = "pending"
	SyncStatusRunning  = "running"
	SyncStatusSuccess  = "success"
	SyncStatusFailed   = "failed"
)

// CreateGrafanaInstanceRequest represents a request to create a Grafana instance
type CreateGrafanaInstanceRequest struct {
	ClusterID          *uuid.UUID `json:"clusterId,omitempty"`
	Name               string     `json:"name" binding:"required"`
	URL                string     `json:"url" binding:"required"`
	APIKey             string     `json:"apiKey,omitempty"`
	Username           string     `json:"username,omitempty"`
	Password           string     `json:"password,omitempty"`
	ServiceAccountID   string     `json:"serviceAccountId,omitempty"`
	ServiceAccountToken string    `json:"serviceAccountToken,omitempty"`
	AutoSync           bool       `json:"autoSync,omitempty"`
	SyncInterval       int        `json:"syncInterval,omitempty"`
}

// UpdateGrafanaInstanceRequest represents a request to update a Grafana instance
type UpdateGrafanaInstanceRequest struct {
	Name               *string `json:"name,omitempty"`
	URL                *string `json:"url,omitempty"`
	APIKey             *string `json:"apiKey,omitempty"`
	Username           *string `json:"username,omitempty"`
	Password           *string `json:"password,omitempty"`
	ServiceAccountID   *string `json:"serviceAccountId,omitempty"`
	ServiceAccountToken *string `json:"serviceAccountToken,omitempty"`
	Status             *string `json:"status,omitempty"`
	AutoSync           *bool   `json:"autoSync,omitempty"`
	SyncInterval       *int    `json:"syncInterval,omitempty"`
}

// TestGrafanaInstanceRequest represents a request to test a Grafana instance
type TestGrafanaInstanceRequest struct {
	URL                string `json:"url" binding:"required"`
	APIKey             string `json:"apiKey,omitempty"`
	Username           string `json:"username,omitempty"`
	Password           string `json:"password,omitempty"`
	ServiceAccountID   string `json:"serviceAccountId,omitempty"`
	ServiceAccountToken string `json:"serviceAccountToken,omitempty"`
}

// TestGrafanaInstanceResponse represents the response from testing a Grafana instance
type TestGrafanaInstanceResponse struct {
	Success    bool   `json:"success"`
	Version    string `json:"version,omitempty"`
	Message    string `json:"message"`
	Error      string `json:"error,omitempty"`
	Duration   int64  `json:"duration"` // milliseconds
}

// SyncGrafanaInstanceRequest represents a request to sync a Grafana instance
type SyncGrafanaInstanceRequest struct {
	SyncDashboards  bool `json:"syncDashboards,omitempty"`
	SyncDataSources bool `json:"syncDataSources,omitempty"`
	SyncFolders     bool `json:"syncFolders,omitempty"`
}

// SyncGrafanaInstanceResponse represents the response from syncing a Grafana instance
type SyncGrafanaInstanceResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	DashboardsAdded int    `json:"dashboardsAdded,omitempty"`
	DashboardsUpdated int  `json:"dashboardsUpdated,omitempty"`
	DashboardsRemoved int  `json:"dashboardsRemoved,omitempty"`
	DataSourcesAdded int   `json:"dataSourcesAdded,omitempty"`
	FoldersAdded     int   `json:"foldersAdded,omitempty"`
	Duration        int64  `json:"duration"` // milliseconds
}

// GrafanaInstanceListResponse represents the response from listing Grafana instances
type GrafanaInstanceListResponse struct {
	Instances   []GrafanaInstance `json:"instances"`
	Total       int64             `json:"total"`
	Page        int               `json:"page"`
	PageSize    int               `json:"pageSize"`
	TotalPages  int64             `json:"totalPages"`
}

// GrafanaDashboardListResponse represents the response from listing Grafana dashboards
type GrafanaDashboardListResponse struct {
	Dashboards  []GrafanaDashboard `json:"dashboards"`
	Total       int64              `json:"total"`
	Page        int                `json:"page"`
	PageSize    int                `json:"pageSize"`
	TotalPages  int64              `json:"totalPages"`
}

// ImportGrafanaDashboardRequest represents a request to import a dashboard
type ImportGrafanaDashboardRequest struct {
	InstanceID    uuid.UUID `json:"instanceId" binding:"required"`
	DashboardID   int       `json:"dashboardId" binding:"required"`
	FolderUID     string    `json:"folderUid,omitempty"`
	Overwrite     bool      `json:"overwrite,omitempty"`
}

// ExportGrafanaDashboardRequest represents a request to export a dashboard
type ExportGrafanaDashboardRequest struct {
	InstanceID    uuid.UUID `json:"instanceId" binding:"required"`
	DashboardUID  string    `json:"dashboardUid" binding:"required"`
}
