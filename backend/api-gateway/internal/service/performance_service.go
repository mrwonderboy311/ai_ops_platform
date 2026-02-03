// Package service provides performance monitoring services
package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PerformanceService handles performance metrics aggregation and analysis
type PerformanceService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPerformanceService creates a new performance service
func NewPerformanceService(db *gorm.DB, logger *zap.Logger) *PerformanceService {
	return &PerformanceService{
		db:     db,
		logger: logger,
	}
}

// CollectMetric collects a single performance metric
func (s *PerformanceService) CollectMetric(metric *model.PerformanceMetric) error {
	metric.CreatedAt = time.Now()
	metric.UpdatedAt = time.Now()
	return s.db.Create(metric).Error
}

// GetMetrics retrieves metrics with filters
func (s *PerformanceService) GetMetrics(metricType, entityType, entityID string, startTime, endTime int64) ([]model.PerformanceMetric, error) {
	var metrics []model.PerformanceMetric
	query := s.db.Model(&model.PerformanceMetric{})

	if metricType != "" {
		query = query.Where("metric_type = ?", metricType)
	}
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if entityID != "" {
		query = query.Where("entity_id = ?", entityID)
	}
	if startTime > 0 {
		query = query.Where("timestamp >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("timestamp <= ?", endTime)
	}

	err := query.Order("timestamp DESC").Limit(1000).Find(&metrics).Error
	return metrics, err
}

// AggregateSnapshot creates a performance snapshot for a time window
func (s *PerformanceService) AggregateSnapshot(timeWindow string, startTime, endTime int64) (*model.PerformanceSnapshot, error) {
	var metrics []model.PerformanceMetric
	err := s.db.Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).Find(&metrics).Error
	if err != nil {
		return nil, err
	}

	snapshot := &model.PerformanceSnapshot{
		ID:        uuid.New(),
		TimeWindow: timeWindow,
		StartTime: startTime,
		EndTime:   endTime,
		CreatedAt: time.Now(),
	}

	if len(metrics) == 0 {
		return snapshot, nil
	}

	// Get unique entity count
	entities := make(map[string]bool)
	cpuUsages := []float64{}
	memoryUsages := []float64{}
	diskUsages := []float64{}

	for _, m := range metrics {
		key := m.EntityType + ":" + m.EntityID
		entities[key] = true

		switch m.MetricType {
		case "cpu":
			cpuUsages = append(cpuUsages, m.Value)
		case "memory":
			memoryUsages = append(memoryUsages, m.Value)
		case "disk":
			diskUsages = append(diskUsages, m.Value)
		}
	}

	snapshot.EntityCount = int32(len(entities))

	snapshot.AvgCPUUsage, snapshot.MaxCPUUsage = calculateStats(cpuUsages)
	snapshot.AvgMemoryUsage, snapshot.MaxMemoryUsage = calculateStats(memoryUsages)
	snapshot.AvgDiskUsage, snapshot.MaxDiskUsage = calculateStats(diskUsages)

	return snapshot, s.db.Create(snapshot).Error
}

// GetLatestSnapshot retrieves the latest performance snapshot
func (s *PerformanceService) GetLatestSnapshot(timeWindow string) (*model.PerformanceSnapshot, error) {
	var snapshot model.PerformanceSnapshot
	query := s.db.Order("created_at DESC")
	if timeWindow != "" {
		query = query.Where("time_window = ?", timeWindow)
	}
	err := query.First(&snapshot).Error
	return &snapshot, err
}

// CalculateSystemHealth calculates overall system health status
func (s *PerformanceService) CalculateSystemHealth() (*model.SystemHealth, error) {
	now := time.Now().Unix()
	fiveMinAgo := now - 300

	// Get recent metrics
	var metrics []model.PerformanceMetric
	err := s.db.Where("timestamp >= ?", fiveMinAgo).Find(&metrics).Error
	if err != nil {
		return nil, err
	}

	health := &model.SystemHealth{
		ID:        uuid.New(),
		Timestamp: now,
		CreatedAt: time.Now(),
	}

	// Count hosts and clusters from metrics
	hosts := make(map[string]bool)
	clusters := make(map[string]bool)
	warningCount := 0
	criticalCount := 0

	for _, m := range metrics {
		switch m.EntityType {
		case "host":
			hosts[m.EntityID] = true
		case "cluster":
			clusters[m.EntityID] = true
		}

		// Check thresholds
		if m.MetricType == "cpu" && m.Value > 90 {
			criticalCount++
		} else if m.MetricType == "cpu" && m.Value > 75 {
			warningCount++
		}
	}

	health.HostCount = int32(len(hosts))
	health.ClusterCount = int32(len(clusters))
	health.HealthyHosts = health.HostCount
	health.HealthyClusters = health.ClusterCount
	health.WarningCount = int32(warningCount)
	health.CriticalCount = int32(criticalCount)

	// Determine overall status
	if criticalCount > 0 {
		health.OverallStatus = "critical"
	} else if warningCount > 0 {
		health.OverallStatus = "warning"
	} else {
		health.OverallStatus = "healthy"
	}

	// Save latest health
	s.db.Where("1 = 1").Delete(&model.SystemHealth{})
	return health, s.db.Create(health).Error
}

// GetSystemHealth retrieves current system health
func (s *PerformanceService) GetSystemHealth() (*model.SystemHealth, error) {
	var health model.SystemHealth
	err := s.db.Order("timestamp DESC").First(&health).Error
	return &health, err
}

// GetTrendData retrieves trend data for a specific metric
func (s *PerformanceService) GetTrendData(metricType, entityType string, points int) ([]TrendPoint, error) {
	var result []TrendPoint

	var metrics []model.PerformanceMetric
	query := s.db.Where("metric_type = ?", metricType)
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	err := query.Order("timestamp DESC").Limit(points).Find(&metrics).Error
	if err != nil {
		return nil, err
	}

	for _, m := range metrics {
		result = append(result, TrendPoint{
			Timestamp: m.Timestamp,
			Value:     m.Value,
		})
	}

	return result, nil
}

// TrendPoint represents a data point in a trend
type TrendPoint struct {
	Timestamp int64
	Value     float64
}

func calculateStats(values []float64) (avg, max float64) {
	if len(values) == 0 {
		return 0, 0
	}

	sum := 0.0
	max = values[0]
	for _, v := range values {
		sum += v
		if v > max {
			max = v
		}
	}
	avg = sum / float64(len(values))
	return avg, max
}
