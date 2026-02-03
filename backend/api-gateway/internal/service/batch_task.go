// Package service provides business logic for batch task management
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"github.com/wangjialin/myops/pkg/ssh"
	"gorm.io/gorm"
	"go.uber.org/zap"
)

// BatchTaskExecutor handles execution of batch tasks across multiple hosts
type BatchTaskExecutor struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewBatchTaskExecutor creates a new batch task executor
func NewBatchTaskExecutor(db *gorm.DB, logger *zap.Logger) *BatchTaskExecutor {
	return &BatchTaskExecutor{
		db:     db,
		logger: logger,
	}
}

// ExecuteTask executes a batch task on specified hosts
func (e *BatchTaskExecutor) ExecuteTask(ctx context.Context, taskID uuid.UUID, hostIDs []uuid.UUID) error {
	// Get batch task
	var task model.BatchTask
	if err := e.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return fmt.Errorf("batch task not found: %w", err)
	}

	// Update task status
	now := time.Now()
	task.Status = model.BatchTaskStatusRunning
	task.StartedAt = &now
	task.TotalHosts = int32(len(hostIDs))
	if err := e.db.Save(&task).Error; err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Create task host records
	taskHosts := make([]model.BatchTaskHost, len(hostIDs))
	for i, hostID := range hostIDs {
		taskHosts[i] = model.BatchTaskHost{
			BatchTaskID: taskID,
			HostID:      hostID,
			Status:      model.BatchTaskStatusPending,
		}
	}
	if err := e.db.Create(&taskHosts).Error; err != nil {
		return fmt.Errorf("failed to create task hosts: %w", err)
	}

	// Execute based on strategy
	switch task.Strategy {
	case model.StrategyParallel:
		return e.executeParallel(ctx, &task, hostIDs)
	case model.StrategySerial:
		return e.executeSerial(ctx, &task, hostIDs)
	case model.StrategyRolling:
		return e.executeRolling(ctx, &task, hostIDs)
	default:
		return fmt.Errorf("unknown execution strategy: %s", task.Strategy)
	}
}

// executeParallel executes the task on all hosts simultaneously
func (e *BatchTaskExecutor) executeParallel(ctx context.Context, task *model.BatchTask, hostIDs []uuid.UUID) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(hostIDs))

	for _, hostID := range hostIDs {
		wg.Add(1)
		go func(hid uuid.UUID) {
			defer wg.Done()
			if err := e.executeOnHost(ctx, task, hid); err != nil {
				e.logger.Error("task execution failed on host",
					zap.String("taskId", task.ID.String()),
					zap.String("hostId", hid.String()),
					zap.Error(err))
				errChan <- err
			}
		}(hostID)
	}

	wg.Wait()
	close(errChan)

	// Check if any errors occurred
	hasErrors := false
	for range errChan {
		hasErrors = true
	}

	return e.finalizeTask(task, hasErrors)
}

// executeSerial executes the task on hosts one at a time
func (e *BatchTaskExecutor) executeSerial(ctx context.Context, task *model.BatchTask, hostIDs []uuid.UUID) error {
	hasErrors := false

	for _, hostID := range hostIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := e.executeOnHost(ctx, task, hostID); err != nil {
			e.logger.Error("task execution failed on host",
				zap.String("taskId", task.ID.String()),
				zap.String("hostId", hostID.String()),
				zap.Error(err))
			hasErrors = true
		}
	}

	return e.finalizeTask(task, hasErrors)
}

// executeRolling executes the task with a percentage of hosts at a time
func (e *BatchTaskExecutor) executeRolling(ctx context.Context, task *model.BatchTask, hostIDs []uuid.UUID) error {
	// Calculate batch size
	parallelism := task.Parallelism
	if parallelism <= 0 {
		parallelism = 1 // Default to 1 host at a time
	}
	batchSize := int(parallelism)
	if batchSize > len(hostIDs) {
		batchSize = len(hostIDs)
	}

	hasErrors := false

	// Process in batches
	for i := 0; i < len(hostIDs); i += batchSize {
		end := i + batchSize
		if end > len(hostIDs) {
			end = len(hostIDs)
		}

		batch := hostIDs[i:end]

		// Execute batch in parallel
		var wg sync.WaitGroup
		for _, hostID := range batch {
			wg.Add(1)
			go func(hid uuid.UUID) {
				defer wg.Done()
				if err := e.executeOnHost(ctx, task, hid); err != nil {
					e.logger.Error("task execution failed on host",
						zap.String("taskId", task.ID.String()),
						zap.String("hostId", hid.String()),
						zap.Error(err))
					hasErrors = true
				}
			}(hostID)
		}
		wg.Wait()
	}

	return e.finalizeTask(task, hasErrors)
}

// executeOnHost executes the task on a single host
func (e *BatchTaskExecutor) executeOnHost(ctx context.Context, task *model.BatchTask, hostID uuid.UUID) error {
	// Get host
	var host model.Host
	if err := e.db.Where("id = ?", hostID).First(&host).Error; err != nil {
		return fmt.Errorf("host not found: %w", err)
	}

	// Get task host record
	var taskHost model.BatchTaskHost
	if err := e.db.Where("batch_task_id = ? AND host_id = ?", task.ID, hostID).First(&taskHost).Error; err != nil {
		return fmt.Errorf("task host record not found: %w", err)
	}

	// Update status to running
	now := time.Now()
	taskHost.Status = model.BatchTaskStatusRunning
	taskHost.StartedAt = &now
	if err := e.db.Save(&taskHost).Error; err != nil {
		return fmt.Errorf("failed to update task host status: %w", err)
	}

	// Create SSH config
	config := &ssh.SSHConfig{
		HostID:     host.ID.String(),
		IPAddress:  host.IPAddress,
		Port:       host.Port,
		Username:   "root", // TODO: Get from credentials
		Password:   "",     // TODO: Get from credentials
		PrivateKey: []byte{},
		Timeout:    30 * time.Second,
	}

	// Execute command
	client, err := ssh.NewClient(config)
	if err != nil {
		return e.markHostFailed(&taskHost, fmt.Sprintf("failed to connect: %v", err))
	}
	defer client.Close()

	timeout := time.Duration(task.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	response, err := client.ExecuteCommand(task.Command, timeout)
	if err != nil {
		return e.markHostFailed(&taskHost, fmt.Sprintf("execution failed: %v", err))
	}

	// Update task host with results
	completedAt := time.Now()
	duration := completedAt.Sub(*taskHost.StartedAt).Milliseconds()

	taskHost.Status = model.BatchTaskStatusCompleted
	taskHost.ExitCode = response.ExitCode
	taskHost.Stdout = response.Stdout
	taskHost.Stderr = response.Stderr
	taskHost.Duration = duration
	taskHost.CompletedAt = &completedAt

	if err := e.db.Save(&taskHost).Error; err != nil {
		return fmt.Errorf("failed to update task host: %w", err)
	}

	// Update task progress
	e.updateProgress(task)

	return nil
}

// markHostFailed marks a host task as failed
func (e *BatchTaskExecutor) markHostFailed(taskHost *model.BatchTaskHost, errMsg string) error {
	now := time.Now()
	taskHost.Status = model.BatchTaskStatusFailed
	taskHost.ErrorMessage = errMsg
	taskHost.CompletedAt = &now
	if err := e.db.Save(taskHost).Error; err != nil {
		return err
	}

	// Get parent task and update failed count
	var task model.BatchTask
	if err := e.db.Where("id = ?", taskHost.BatchTaskID).First(&task).Error; err == nil {
		task.FailedHosts++
		e.db.Save(&task)
	}

	return nil
}

// updateProgress updates the progress of a batch task
func (e *BatchTaskExecutor) updateProgress(task *model.BatchTask) {
	var completedCount int64
	e.db.Model(&model.BatchTaskHost{}).
		Where("batch_task_id = ? AND status IN (?)", task.ID,
			[]string{string(model.BatchTaskStatusCompleted), string(model.BatchTaskStatusFailed)}).
		Count(&completedCount)

	task.CompletedHosts = int32(completedCount)
	e.db.Save(task)
}

// finalizeTask finalizes a batch task after all hosts are processed
func (e *BatchTaskExecutor) finalizeTask(task *model.BatchTask, hasErrors bool) error {
	now := time.Now()
	task.CompletedAt = &now

	if hasErrors {
		task.Status = model.BatchTaskStatusFailed
	} else {
		task.Status = model.BatchTaskStatusCompleted
	}

	return e.db.Save(task).Error
}

// CancelTask cancels a running batch task
func (e *BatchTaskExecutor) CancelTask(ctx context.Context, taskID uuid.UUID) error {
	var task model.BatchTask
	if err := e.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return fmt.Errorf("batch task not found: %w", err)
	}

	if task.Status != model.BatchTaskStatusRunning && task.Status != model.BatchTaskStatusPending {
		return fmt.Errorf("task cannot be cancelled in current state: %s", task.Status)
	}

	// Update task status
	now := time.Now()
	task.Status = model.BatchTaskStatusCancelled
	task.CompletedAt = &now
	if err := e.db.Save(&task).Error; err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	// Cancel pending/running task hosts
	e.db.Model(&model.BatchTaskHost{}).
		Where("batch_task_id = ? AND status IN (?)", taskID,
			[]string{string(model.BatchTaskStatusPending), string(model.BatchTaskStatusRunning)}).
		Updates(map[string]interface{}{
			"status":       model.BatchTaskStatusCancelled,
			"completed_at": now,
		})

	return nil
}

// GetTaskProgress returns the progress of a batch task
func (e *BatchTaskExecutor) GetTaskProgress(taskID uuid.UUID) (*model.BatchTaskResponse, error) {
	var task model.BatchTask
	if err := e.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, fmt.Errorf("batch task not found: %w", err)
	}

	// Get task hosts
	var hosts []model.BatchTaskHost
	if err := e.db.Where("batch_task_id = ?", taskID).Find(&hosts).Error; err != nil {
		return nil, fmt.Errorf("failed to get task hosts: %w", err)
	}

	// Calculate progress
	progress := 0.0
	if task.TotalHosts > 0 {
		progress = float64(task.CompletedHosts) / float64(task.TotalHosts) * 100
	}

	return &model.BatchTaskResponse{
		BatchTask: &task,
		Hosts:     hosts,
		Progress:  progress,
	}, nil
}
