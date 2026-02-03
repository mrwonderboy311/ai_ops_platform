// Package db provides database operations
package db

import (
	"context"
	"time"

	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// ScanTaskRepository handles scan task database operations
type ScanTaskRepository struct {
	db *gorm.DB
}

// NewScanTaskRepository creates a new ScanTaskRepository
func NewScanTaskRepository(db *gorm.DB) *ScanTaskRepository {
	return &ScanTaskRepository{db: db}
}

// Create creates a new scan task
func (r *ScanTaskRepository) Create(ctx context.Context, task *model.ScanTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// FindByID finds a scan task by ID
func (r *ScanTaskRepository) FindByID(ctx context.Context, id string) (*model.ScanTask, error) {
	var task model.ScanTask
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update updates a scan task
func (r *ScanTaskRepository) Update(ctx context.Context, task *model.ScanTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// IncrementDiscovered increments the discovered hosts counter
func (r *ScanTaskRepository) IncrementDiscovered(ctx context.Context, taskID string) error {
	return r.db.WithContext(ctx).
		Model(&model.ScanTask{}).
		Where("id = ?", taskID).
		UpdateColumn("discovered_hosts", gorm.Expr("discovered_hosts + 1")).
		Error
}

// MarkCompleted marks a scan task as completed
func (r *ScanTaskRepository) MarkCompleted(ctx context.Context, taskID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.ScanTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":       model.ScanTaskStatusCompleted,
			"completed_at": &now,
		}).Error
}

// MarkFailed marks a scan task as failed
func (r *ScanTaskRepository) MarkFailed(ctx context.Context, taskID string, errMsg string) error {
	return r.db.WithContext(ctx).
		Model(&model.ScanTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":         model.ScanTaskStatusFailed,
			"error_message":  errMsg,
			"completed_at":    time.Now(),
		}).Error
}

// CreateDiscoveredHost creates a discovered host entry
func (r *ScanTaskRepository) CreateDiscoveredHost(ctx context.Context, host *model.DiscoveredHost) error {
	return r.db.WithContext(ctx).Create(host).Error
}

// GetDiscoveredHosts retrieves discovered hosts for a task
func (r *ScanTaskRepository) GetDiscoveredHosts(ctx context.Context, taskID string) ([]*model.DiscoveredHost, error) {
	var hosts []*model.DiscoveredHost
	err := r.db.WithContext(ctx).
		Where("scan_task_id = ?", taskID).
		Find(&hosts).Error
	return hosts, err
}

// ListUserTasks retrieves scan tasks for a user
func (r *ScanTaskRepository) ListUserTasks(ctx context.Context, userID string, limit int) ([]*model.ScanTask, error) {
	var tasks []*model.ScanTask
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&tasks).Error
	return tasks, err
}
