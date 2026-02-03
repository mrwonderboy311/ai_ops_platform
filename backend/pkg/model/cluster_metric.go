// Package model provides data models for cluster monitoring
package model

import (
	"time"

	"github.com/google/uuid"
)

// ClusterMetric represents cluster-level metrics
type ClusterMetric struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID       uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_cluster_metrics_time,sort:ordered"`
	Timestamp       int64     `json:"timestamp" gorm:"not null;index:idx_cluster_metrics_time,sort:ordered"`
	CPUUsagePercent float64   `json:"cpuUsagePercent" gorm:"type:decimal(5,2)"`
	MemoryUsageBytes int64    `json:"memoryUsageBytes" gorm:"type:bigint"`
	MemoryTotalBytes int64    `json:"memoryTotalBytes" gorm:"type:bigint"`
	PodCount        int32     `json:"podCount" gorm:"type:int"`
	RunningPodCount int32     `json:"runningPodCount" gorm:"type:int"`
	PendingPodCount int32     `json:"pendingPodCount" gorm:"type:int"`
	FailedPodCount  int32     `json:"failedPodCount" gorm:"type:int"`
	NodeCount       int32     `json:"nodeCount" gorm:"type:int"`
	ReadyNodeCount  int32     `json:"readyNodeCount" gorm:"type:int"`
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// NodeMetric represents node-level metrics
type NodeMetric struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID       uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_node_metrics_time,sort:ordered"`
	NodeName        string    `json:"nodeName" gorm:"type:varchar(255);not null;index:idx_node_metrics_time,sort:ordered"`
	Timestamp       int64     `json:"timestamp" gorm:"not null;index:idx_node_metrics_time,sort:ordered"`
	CPUUsagePercent float64   `json:"cpuUsagePercent" gorm:"type:decimal(5,2)"`
	MemoryUsageBytes int64    `json:"memoryUsageBytes" gorm:"type:bigint"`
	MemoryTotalBytes int64    `json:"memoryTotalBytes" gorm:"type:bigint"`
	DiskUsageBytes  int64    `json:"diskUsageBytes" gorm:"type:bigint"`
	DiskTotalBytes  int64    `json:"diskTotalBytes" gorm:"type:bigint"`
	PodCount        int32     `json:"podCount" gorm:"type:int"`
	NetworkRxBytes  int64     `json:"networkRxBytes" gorm:"type:bigint"`
	NetworkTxBytes  int64     `json:"networkTxBytes" gorm:"type:bigint"`
	Status          string    `json:"status" gorm:"type:varchar(20)"`
	Ready           bool      `json:"ready" gorm:"type:boolean"`
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// PodMetric represents pod-level metrics
type PodMetric struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClusterID       uuid.UUID `json:"clusterId" gorm:"type:uuid;not null;index:idx_pod_metrics_time,sort:ordered"`
	Namespace       string    `json:"namespace" gorm:"type:varchar(255);not null;index:idx_pod_metrics_time,sort:ordered"`
	PodName         string    `json:"podName" gorm:"type:varchar(255);not null;index:idx_pod_metrics_time,sort:ordered"`
	Timestamp       int64     `json:"timestamp" gorm:"not null;index:idx_pod_metrics_time,sort:ordered"`
	CPUUsageCores   float64   `json:"cpuUsageCores" gorm:"type:decimal(10,4)"`
	MemoryUsageBytes int64    `json:"memoryUsageBytes" gorm:"type:bigint"`
	RestartCount    int32     `json:"restartCount" gorm:"type:int"`
	Status          string    `json:"status" gorm:"type:varchar(20)"`
	Ready           bool      `json:"ready" gorm:"type:boolean"`
	NodeName        string    `json:"nodeName" gorm:"type:varchar(255)"`
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// ClusterMetricSummary represents aggregated cluster metrics
type ClusterMetricSummary struct {
	Timestamp       int64   `json:"timestamp"`
	CPUUsagePercent float64 `json:"cpuUsagePercent"`
	MemoryUsagePercent float64 `json:"memoryUsagePercent"`
	PodCount        int32   `json:"podCount"`
	RunningPodCount int32   `json:"runningPodCount"`
	PendingPodCount int32   `json:"pendingPodCount"`
	FailedPodCount  int32   `json:"failedPodCount"`
	NodeCount       int32   `json:"nodeCount"`
	ReadyNodeCount  int32   `json:"readyNodeCount"`
}

// NodeMetricSummary represents aggregated node metrics
type NodeMetricSummary struct {
	NodeName        string  `json:"nodeName"`
	Timestamp       int64   `json:"timestamp"`
	CPUUsagePercent float64 `json:"cpuUsagePercent"`
	MemoryUsagePercent float64 `json:"memoryUsagePercent"`
	PodCount        int32   `json:"podCount"`
	Status          string  `json:"status"`
	Ready           bool    `json:"ready"`
}

// PodMetricSummary represents aggregated pod metrics
type PodMetricSummary struct {
	Namespace       string  `json:"namespace"`
	PodName         string  `json:"podName"`
	Timestamp       int64   `json:"timestamp"`
	CPUUsageCores   float64 `json:"cpuUsageCores"`
	MemoryUsageMB   int64   `json:"memoryUsageMb"`
	RestartCount    int32   `json:"restartCount"`
	Status          string  `json:"status"`
	Ready           bool    `json:"ready"`
	NodeName        string  `json:"nodeName"`
}
