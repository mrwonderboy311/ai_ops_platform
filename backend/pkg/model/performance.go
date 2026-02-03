// Package model provides data models for performance monitoring
package model

import (
	"time"

	"github.com/google/uuid"
)

// PerformanceMetric represents a system performance metric
type PerformanceMetric struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MetricType  string    // cpu, memory, disk, network, response_time, error_rate, throughput
	EntityType  string    // host, cluster, pod, container, node
	EntityID    string
	Value       float64
	Unit        string    // percent, bytes, ms, requests/sec, errors/sec
	Timestamp   int64
	Labels      map[string]string `gorm:"serializer:json"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PerformanceSnapshot represents an aggregated performance snapshot
type PerformanceSnapshot struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TimeWindow      string    // 5m, 15m, 1h, 1d
	StartTime       int64
	EndTime         int64
	EntityCount     int32
	AvgCPUUsage     float64
	MaxCPUUsage     float64
	AvgMemoryUsage  float64
	MaxMemoryUsage  float64
	AvgDiskUsage    float64
	MaxDiskUsage    float64
	AvgResponseTime float64 // ms
	ErrorRate       float64 // percentage
	Throughput      float64 // requests/sec
	CreatedAt       time.Time
}

// SystemHealth represents overall system health status
type SystemHealth struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OverallStatus   string    // healthy, warning, critical
	HostCount       int32
	ClusterCount    int32
	HealthyHosts    int32
	HealthyClusters int32
	WarningCount    int32
	CriticalCount   int32
	Timestamp       int64
	CreatedAt       time.Time
}

// PerformanceAlertThreshold represents alert thresholds for performance metrics
type PerformanceAlertThreshold struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MetricType  string
	EntityType  string
	WarningMin  float64
	WarningMax  float64
	CriticalMin float64
	CriticalMax float64
	Duration    int64 // seconds
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName specifies the table name for PerformanceMetric
func (PerformanceMetric) TableName() string {
	return "performance_metrics"
}

// TableName specifies the table name for PerformanceSnapshot
func (PerformanceSnapshot) TableName() string {
	return "performance_snapshots"
}

// TableName specifies the table name for SystemHealth
func (SystemHealth) TableName() string {
	return "system_health"
}

// TableName specifies the table name for PerformanceAlertThreshold
func (PerformanceAlertThreshold) TableName() string {
	return "performance_alert_thresholds"
}
