// Package model provides data models for Kubernetes workloads
package model

import (
	"time"

	"github.com/google/uuid"
)

// Namespace represents a Kubernetes namespace
type Namespace struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_cluster_namespace"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null;index:idx_cluster_namespace,unique"`
	Status    string    `json:"status" gorm:"type:varchar(20)"`
	Labels    string    `json:"labels" gorm:"type:text"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// Deployment represents a Kubernetes deployment
type Deployment struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID         uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_cluster_namespace"`
	Namespace         string    `json:"namespace" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Name              string    `json:"name" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Replicas          int32     `json:"replicas" gorm:"type:int"`
	AvailableReplicas int32     `json:"availableReplicas" gorm:"type:int"`
	UpdatedReplicas   int32     `json:"updatedReplicas" gorm:"type:int"`
	ReadyReplicas     int32     `json:"readyReplicas" gorm:"type:int"`
	Labels            string    `json:"labels" gorm:"type:text"`
	Selector          string    `json:"selector" gorm:"type:text"`
	Image             string    `json:"image" gorm:"type:varchar(500)"`
	CreatedAt         time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// StatefulSet represents a Kubernetes statefulset
type StatefulSet struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID       uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_cluster_namespace"`
	Namespace       string    `json:"namespace" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Name            string    `json:"name" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Replicas        int32     `json:"replicas" gorm:"type:int"`
	ReadyReplicas   int32     `json:"readyReplicas" gorm:"type:int"`
	CurrentReplicas int32     `json:"currentReplicas" gorm:"type:int"`
	UpdatedReplicas int32     `json:"updatedReplicas" gorm:"type:int"`
	Labels          string    `json:"labels" gorm:"type:text"`
	Selector        string    `json:"selector" gorm:"type:text"`
	Image           string    `json:"image" gorm:"type:varchar(500)"`
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// DaemonSet represents a Kubernetes daemonset
type DaemonSet struct {
	ID                     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID              uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_cluster_namespace"`
	Namespace              string    `json:"namespace" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Name                   string    `json:"name" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	CurrentNumberScheduled int32     `json:"currentNumberScheduled" gorm:"type:int"`
	DesiredNumberScheduled int32     `json:"desiredNumberScheduled" gorm:"type:int"`
	NumberReady            int32     `json:"numberReady" gorm:"type:int"`
	UpdatedNumberScheduled int32     `json:"updatedNumberScheduled" gorm:"type:int"`
	Labels                 string    `json:"labels" gorm:"type:text"`
	Selector               string    `json:"selector" gorm:"type:text"`
	Image                  string    `json:"image" gorm:"type:varchar(500)"`
	CreatedAt              time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt              time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// Pod represents a Kubernetes pod
type Pod struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID    uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_cluster_namespace"`
	Namespace    string    `json:"namespace" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Name         string    `json:"name" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Status       string    `json:"status" gorm:"type:varchar(20)"`
	Phase        string    `json:"phase" gorm:"type:varchar(20)"`
	NodeName     string    `json:"nodeName" gorm:"type:varchar(255)"`
	HostIP       string    `json:"hostIp" gorm:"type:varchar(50)"`
	PodIP        string    `json:"podIp" gorm:"type:varchar(50)"`
	Ready        bool      `json:"ready" gorm:"type:boolean"`
	RestartCount int32     `json:"restartCount" gorm:"type:int"`
	Labels       string    `json:"labels" gorm:"type:text"`
	OwnerType    string    `json:"ownerType" gorm:"type:varchar(50)"`
	OwnerName    string    `json:"ownerName" gorm:"type:varchar(255)"`
	Containers   string    `json:"containers" gorm:"type:text"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// Service represents a Kubernetes service
type Service struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID   uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_cluster_namespace"`
	Namespace   string    `json:"namespace" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Type        string    `json:"type" gorm:"type:varchar(50)"`
	ClusterIP   string    `json:"clusterIp" gorm:"type:varchar(50)"`
	ExternalIPs string    `json:"externalIps" gorm:"type:text"`
	Ports       string    `json:"ports" gorm:"type:text"`
	Selector    string    `json:"selector" gorm:"type:text"`
	Labels      string    `json:"labels" gorm:"type:text"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// Ingress represents a Kubernetes ingress
type Ingress struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_cluster_namespace"`
	Namespace string    `json:"namespace" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Class     string    `json:"class" gorm:"type:varchar(100)"`
	Hosts     string    `json:"hosts" gorm:"type:text"`
	Paths     string    `json:"paths" gorm:"type:text"`
	Backend   string    `json:"backend" gorm:"type:text"`
	Labels    string    `json:"labels" gorm:"type:text"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// ConfigMap represents a Kubernetes configmap
type ConfigMap struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_cluster_namespace"`
	Namespace string    `json:"namespace" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Data      string    `json:"data" gorm:"type:text"`
	Labels    string    `json:"labels" gorm:"type:text"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// Secret represents a Kubernetes secret
type Secret struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_cluster_namespace"`
	Namespace string    `json:"namespace" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null;index:idx_cluster_namespace"`
	Type      string    `json:"type" gorm:"type:varchar(100)"`
	DataKeys  string    `json:"dataKeys" gorm:"type:text"`
	Labels    string    `json:"labels" gorm:"type:text"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
