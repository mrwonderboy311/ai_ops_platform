// Package model provides data models for Helm repository management
package model

import (
	"time"

	"github.com/google/uuid"
)

// HelmRepoType represents the type of Helm repository
type HelmRepoType string

const (
	HelmRepoTypeHTTP  HelmRepoType = "http"
	HelmRepoTypeHTTPS HelmRepoType = "https"
	HelmRepoTypeOCI   HelmRepoType = "oci"
)

// HelmRepoStatus represents the status of a Helm repository
type HelmRepoStatus string

const (
	HelmRepoStatusActive   HelmRepoStatus = "active"
	HelmRepoStatusInactive HelmRepoStatus = "inactive"
	HelmRepoStatusError    HelmRepoStatus = "error"
)

// HelmRepository represents a Helm chart repository
type HelmRepository struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID          uuid.UUID      `json:"userId" gorm:"type:uuid;not null;index"`
	Name            string         `json:"name" gorm:"type:varchar(255);not null;uniqueIndex:user_id_name"`
	Description     string         `json:"description" gorm:"type:text"`
	Type            HelmRepoType   `json:"type" gorm:"type:varchar(10);not null"`
	Status          HelmRepoStatus `json:"status" gorm:"type:varchar(20);not null;index"`
	URL             string         `json:"url" gorm:"type:varchar(500);not null"`
	Username        string         `json:"username" gorm:"type:varchar(255)"`
	Password        string         `json:"-" gorm:"type:varchar(255)"` // Encrypted
	CAFile          string         `json:"-" gorm:"type:text"`          // CA certificate (encrypted)
	CertFile        string         `json:"-" gorm:"type:text"`          // Client certificate (encrypted)
	KeyFile         string         `json:"-" gorm:"type:text"`          // Client key (encrypted)
	InsecureSkipTLS bool           `json:"insecureSkipTLS" gorm:"type:boolean;default:false"`
	LastSyncedAt    *time.Time     `json:"lastSyncedAt" gorm:"type:timestamp"`
	LastSyncStatus  string         `json:"lastSyncStatus" gorm:"type:varchar(20)"`
	LastSyncError   string         `json:"lastSyncError" gorm:"type:text"`
	ChartCount      int32          `json:"chartCount" gorm:"type:int;default:0"`
	CreatedAt       time.Time      `json:"createdAt" gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt       time.Time      `json:"updatedAt" gorm:"type:timestamp;autoUpdateTime"`
	// Relations
	User            *User          `json:"user,omitempty" gorm:"foreigner:UserID"`
}

// TableName specifies the table name for HelmRepository
func (HelmRepository) TableName() string {
	return "helm_repositories"
}

// HelmChart represents a Helm chart from a repository
type HelmChart struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	RepoID        uuid.UUID `json:"repoId" gorm:"type:uuid;not null;index"`
	Name          string    `json:"name" gorm:"type:varchar(255);not null"`
	Description   string    `json:"description" gorm:"type:text"`
	Version       string    `json:"version" gorm:"type:varchar(50);not null"`
	AppVersion    string    `json:"appVersion" gorm:"type:varchar(50)"`
	Icon          string    `json:"icon" gorm:"type:varchar(500)"`
	Home          string    `json:"home" gorm:"type:varchar(500)"`
	Keywords      string    `json:"keywords" gorm:"type:text"` // JSON array
	Sources       string    `json:"sources" gorm:"type:text"`  // JSON array
	Maintainers   string    `json:"maintainers" gorm:"type:text"` // JSON array
	Deprecated    bool      `json:"deprecated" gorm:"type:boolean;default:false"`
	CreatedAt     time.Time `json:"createdAt" gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"type:timestamp;autoUpdateTime"`
	// Relations
	Repository    *HelmRepository `json:"repository,omitempty" gorm:"foreigner:RepoID"`
}

// TableName specifies the table name for HelmChart
func (HelmChart) TableName() string {
	return "helm_charts"
}

// CreateHelmRepoRequest represents a request to create a Helm repository
type CreateHelmRepoRequest struct {
	Name            string       `json:"name" binding:"required"`
	Description     string       `json:"description"`
	Type            HelmRepoType `json:"type" binding:"required"`
	URL             string       `json:"url" binding:"required"`
	Username        string       `json:"username"`
	Password        string       `json:"password"`
	CAFile          string       `json:"caFile"`
	CertFile        string       `json:"certFile"`
	KeyFile         string       `json:"keyFile"`
	InsecureSkipTLS bool         `json:"insecureSkipTLS"`
}

// UpdateHelmRepoRequest represents a request to update a Helm repository
type UpdateHelmRepoRequest struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	URL             string  `json:"url"`
	Username        string  `json:"username"`
	Password        string  `json:"password"`
	CAFile          string  `json:"caFile"`
	CertFile        string  `json:"certFile"`
	KeyFile         string  `json:"keyFile"`
	InsecureSkipTLS bool    `json:"insecureSkipTLS"`
}

// ListHelmReposRequest represents a request to list Helm repositories
type ListHelmReposRequest struct {
	Status   *HelmRepoStatus `form:"status"`
	Type     *HelmRepoType   `form:"type"`
	Page     int             `form:"page" binding:"min=1"`
	PageSize int             `form:"pageSize" binding:"min=1,max=100"`
}

// HelmRepoTestRequest represents a request to test a Helm repository connection
type HelmRepoTestRequest struct {
	Type            HelmRepoType `json:"type" binding:"required"`
	URL             string       `json:"url" binding:"required"`
	Username        string       `json:"username"`
	Password        string       `json:"password"`
	CAFile          string       `json:"caFile"`
	CertFile        string       `json:"certFile"`
	KeyFile         string       `json:"keyFile"`
	InsecureSkipTLS bool         `json:"insecureSkipTLS"`
}

// HelmRepoTestResponse represents the response from a repository test
type HelmRepoTestResponse struct {
	Success    bool   `json:"success"`
	ChartCount int    `json:"chartCount"`
	Message    string `json:"message"`
	Error      string `json:"error,omitempty"`
}

// HelmReleaseStatus represents the status of a Helm release
type HelmReleaseStatus string

const (
	HelmReleaseStatusDeployed      HelmReleaseStatus = "deployed"
	HelmReleaseStatusPending       HelmReleaseStatus = "pending"
	HelmReleaseStatusPendingUpgrade HelmReleaseStatus = "pending-upgrade"
	HelmReleaseStatusPendingRollback HelmReleaseStatus = "pending-rollback"
	HelmReleaseStatusSuperseded    HelmReleaseStatus = "superseded"
	HelmReleaseStatusFailed        HelmReleaseStatus = "failed"
	HelmReleaseStatusUnknown       HelmReleaseStatus = "unknown"
	HelmReleaseStatusUninstalling   HelmReleaseStatus = "uninstalling"
)

// HelmRelease represents a Helm release (installed application)
type HelmRelease struct {
	ID              uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID          uuid.UUID         `json:"userId" gorm:"type:uuid;not null;index"`
	ClusterID       uuid.UUID         `json:"clusterId" gorm:"type:uuid;not null;index"`
	Namespace       string            `json:"namespace" gorm:"type:varchar(255);not null;index"`
	Name            string            `json:"name" gorm:"type:varchar(255);not null"`
	Revision        int32             `json:"revision" gorm:"type:int;not null"`
	Updated         string            `json:"updated" gorm:"type:varchar(50)"`
	Status          HelmReleaseStatus `json:"status" gorm:"type:varchar(30);not null"`
	Chart           string            `json:"chart" gorm:"type:varchar(500);not null"`
	ChartVersion    string            `json:"chartVersion" gorm:"type:varchar(50);not null"`
	AppVersion      string            `json:"appVersion" gorm:"type:varchar(50)"`
	Icon            string            `json:"icon" gorm:"type:varchar(500)"`
	Description     string            `json:"description" gorm:"type:text"`
	Values          string            `json:"-" gorm:"type:text"` // YAML values (encrypted if sensitive)
	Notes           string            `json:"notes" gorm:"type:text"`
	CreatedAt       time.Time         `json:"createdAt" gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt       time.Time         `json:"updatedAt" gorm:"type:timestamp;autoUpdateTime"`
	// Relations
	Cluster         *K8sCluster       `json:"cluster,omitempty" gorm:"foreigner:ClusterID"`
	User            *User             `json:"user,omitempty" gorm:"foreigner:UserID"`
}

// TableName specifies the table name for HelmRelease
func (HelmRelease) TableName() string {
	return "helm_releases"
}

// CreateHelmReleaseRequest represents a request to install a Helm release
type CreateHelmReleaseRequest struct {
	ClusterID    uuid.UUID `json:"clusterId" binding:"required"`
	Namespace    string    `json:"namespace" binding:"required"`
	Name         string    `json:"name" binding:"required"`
	Chart        string    `json:"chart" binding:"required"`
	ChartVersion string    `json:"chartVersion"`
	Values       string    `json:"values"`
	Description  string    `json:"description"`
}

// UpdateHelmReleaseRequest represents a request to upgrade a Helm release
type UpdateHelmReleaseRequest struct {
	ChartVersion string `json:"chartVersion"`
	Values       string `json:"values"`
}

// RollbackHelmReleaseRequest represents a request to rollback a Helm release
type RollbackHelmReleaseRequest struct {
	Revision int32 `json:"revision"`
}

// ListHelmReleasesRequest represents a request to list Helm releases
type ListHelmReleasesRequest struct {
	ClusterID  *uuid.UUID         `form:"clusterId"`
	Namespace  *string            `form:"namespace"`
	Status     *HelmReleaseStatus `form:"status"`
	Page       int                `form:"page" binding:"min=1"`
	PageSize   int                `form:"pageSize" binding:"min=1,max=100"`
}

// HelmReleaseHistory represents a revision of a Helm release
type HelmReleaseHistory struct {
	Revision     int32             `json:"revision"`
	Updated      string            `json:"updated"`
	Status       HelmReleaseStatus `json:"status"`
	Chart        string            `json:"chart"`
	ChartVersion string            `json:"chartVersion"`
	AppVersion   string            `json:"appVersion"`
	Description  string            `json:"description"`
}

