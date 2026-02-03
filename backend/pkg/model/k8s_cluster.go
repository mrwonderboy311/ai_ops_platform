// Package model provides data models for Kubernetes cluster management
package model

import (
	"time"

	"github.com/google/uuid"
)

// ClusterStatus represents the status of a Kubernetes cluster
type ClusterStatus string

const (
	ClusterStatusPending  ClusterStatus = "pending"
	ClusterStatusConnected ClusterStatus = "connected"
	ClusterStatusError    ClusterStatus = "error"
	ClusterStatusDisabled ClusterStatus = "disabled"
)

// ClusterType represents the type of Kubernetes cluster
type ClusterType string

const (
	ClusterTypeManaged   ClusterType = "managed"    // EKS, GKE, AKS, etc.
	ClusterTypeSelfHosted ClusterType = "self-hosted" // kubeadm, k3s, etc.
)

// K8sCluster represents a Kubernetes cluster
type K8sCluster struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID          uuid.UUID      `json:"userId" gorm:"type:uuid;not null;index"`
	Name            string         `json:"name" gorm:"type:varchar(255);not null;uniqueIndex:user_id_name"`
	Description     string         `json:"description" gorm:"type:text"`
	Type            ClusterType    `json:"type" gorm:"type:varchar(20);not null"`
	Status          ClusterStatus  `json:"status" gorm:"type:varchar(20);not null;index"`
	Endpoint        string         `json:"endpoint" gorm:"type:varchar(500)"` // API Server endpoint
	Kubeconfig      string         `json:"-" gorm:"type:text"`               // Encrypted kubeconfig
	Version         string         `json:"version" gorm:"type:varchar(50)"`  // Kubernetes version
	NodeCount       int32          `json:"nodeCount" gorm:"type:int"`
	Region          string         `json:"region" gorm:"type:varchar(100)"`
	Provider        string         `json:"provider" gorm:"type:varchar(50)"` // aws, gcp, azure, etc.
	LastConnectedAt *time.Time     `json:"lastConnectedAt" gorm:"type:timestamp"`
	ErrorMessage    string         `json:"errorMessage" gorm:"type:text"`
	CreatedAt       time.Time      `json:"createdAt" gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt       time.Time      `json:"updatedAt" gorm:"type:timestamp;autoUpdateTime"`
	// Relations
	User            *User          `json:"user,omitempty" gorm:"foreigner:UserID"`
}

// TableName specifies the table name for K8sCluster
func (K8sCluster) TableName() string {
	return "k8s_clusters"
}

// ClusterNode represents a node in a Kubernetes cluster
type ClusterNode struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID         uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index"`
	Name              string    `json:"name" gorm:"type:varchar(255);not null"`
	InternalIP        string    `json:"internalIp" gorm:"type:varchar(50)"`
	ExternalIP        string    `json:"externalIp" gorm:"type:varchar(50)"`
	Status            string    `json:"status" gorm:"type:varchar(50)"`
	Roles             string    `json:"roles" gorm:"type:varchar(100)"` // master, worker, etc.
	Version           string    `json:"version" gorm:"type:varchar(50)"`
	OSImage           string    `json:"osImage" gorm:"type:varchar(255)"`
	KernelVersion     string    `json:"kernelVersion" gorm:"type:varchar(100)"`
	ContainerRuntime  string    `json:"containerRuntime" gorm:"type:varchar(100)"`
	CPUCapacity       string    `json:"cpuCapacity" gorm:"type:varchar(20)"`
	MemoryCapacity    string    `json:"memoryCapacity" gorm:"type:varchar(20)"`
	StorageCapacity   string    `json:"storageCapacity" gorm:"type:varchar(20)"`
	CPUAllocatable    string    `json:"cpuAllocatable" gorm:"type:varchar(20)"`
	MemoryAllocatable string    `json:"memoryAllocatable" gorm:"type:varchar(20)"`
	StorageAllocatable string   `json:"storageAllocatable" gorm:"type:varchar(20)"`
	PodCount          int32     `json:"podCount" gorm:"type:int"`
	Conditions        string    `json:"conditions" gorm:"type:text"` // JSON string
	CreatedAt         time.Time `json:"createdAt" gorm:"type:timestamp;autoCreateTime"`
	UpdatedAt         time.Time `json:"updatedAt" gorm:"type:timestamp;autoUpdateTime"`
	// Relations
	Cluster           *K8sCluster `json:"cluster,omitempty" gorm:"foreigner:ClusterID"`
}

// TableName specifies the table name for ClusterNode
func (ClusterNode) TableName() string {
	return "k8s_cluster_nodes"
}

// ClusterNamespace represents a namespace in a Kubernetes cluster
type ClusterNamespace struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null"`
	Status    string    `json:"status" gorm:"type:varchar(50)"`
	CreatedAt time.Time `json:"createdAt" gorm:"type:timestamp;autoCreateTime"`
	// Relations
	Cluster   *K8sCluster `json:"cluster,omitempty" gorm:"foreigner:ClusterID"`
}

// TableName specifies the table name for ClusterNamespace
func (ClusterNamespace) TableName() string {
	return "k8s_namespaces"
}

// CreateClusterRequest represents a request to create a cluster
type CreateClusterRequest struct {
	Name        string      `json:"name" binding:"required"`
	Description string      `json:"description"`
	Type        ClusterType `json:"type" binding:"required"`
	Endpoint    string      `json:"endpoint"`
	Kubeconfig  string      `json:"kubeconfig" binding:"required"`
	Region      string      `json:"region"`
	Provider    string      `json:"provider"`
}

// UpdateClusterRequest represents a request to update a cluster
type UpdateClusterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	Kubeconfig  string `json:"kubeconfig"`
}

// ListClustersRequest represents a request to list clusters
type ListClustersRequest struct {
	Status   *ClusterStatus `form:"status"`
	Type     *ClusterType   `form:"type"`
	Provider string         `form:"provider"`
	Page     int            `form:"page" binding:"min=1"`
	PageSize int            `form:"pageSize" binding:"min=1,max=100"`
}

// ClusterConnectionTestRequest represents a request to test cluster connection
type ClusterConnectionTestRequest struct {
	Kubeconfig string `json:"kubeconfig" binding:"required"`
	Endpoint   string `json:"endpoint"`
}

// ClusterConnectionTestResponse represents the response from a connection test
type ClusterConnectionTestResponse struct {
	Success   bool   `json:"success"`
	Version   string `json:"version"`
	NodeCount int32  `json:"nodeCount"`
	Error     string `json:"error,omitempty"`
}

// ClusterSummary represents a summary of cluster resources
type ClusterSummary struct {
	NodeCount        int32 `json:"nodeCount"`
	NamespaceCount   int32 `json:"namespaceCount"`
	PodCount         int32 `json:"podCount"`
	DeploymentCount  int32 `json:"deploymentCount"`
	ServiceCount     int32 `json:"serviceCount"`
	IngressCount     int32 `json:"ingressCount"`
	ConfigMapCount   int32 `json:"configMapCount"`
	SecretCount      int32 `json:"secretCount"`
}
