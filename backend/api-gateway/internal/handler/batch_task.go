// Package handler provides HTTP handlers for batch task management
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/api-gateway/internal/service"
	"github.com/wangjialin/myops/pkg/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BatchTaskHandler handles batch task management operations
type BatchTaskHandler struct {
	db                *gorm.DB
	logger            *zap.Logger
	taskExecutor      *service.BatchTaskExecutor
}

// NewBatchTaskHandler creates a new batch task handler
func NewBatchTaskHandler(db *gorm.DB, logger *zap.Logger) *BatchTaskHandler {
	return &BatchTaskHandler{
		db:           db,
		logger:       logger,
		taskExecutor: service.NewBatchTaskExecutor(db, logger),
	}
}

// CreateBatchTask handles batch task creation requests
func (h *BatchTaskHandler) CreateBatchTask(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBatchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Validate request
	if len(req.HostIDs) == 0 {
		respondWithError(w, http.StatusBadRequest, "NO_HOSTS", "At least one host must be specified")
		return
	}

	// Verify hosts exist and are available
	var hosts []model.Host
	if err := h.db.Where("id IN ? AND status IN ?", req.HostIDs,
		[]model.HostStatus{model.HostStatusApproved, model.HostStatusOnline}).Find(&hosts).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify hosts")
		return
	}

	if len(hosts) != len(req.HostIDs) {
		respondWithError(w, http.StatusBadRequest, "INVALID_HOSTS", "Some hosts are not available")
		return
	}

	// Set default values
	if req.Timeout <= 0 {
		req.Timeout = 60
	}
	if req.MaxRetries < 0 {
		req.MaxRetries = 0
	}

	// Create batch task
	task := &model.BatchTask{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Status:      model.BatchTaskStatusPending,
		Strategy:    req.Strategy,
		Command:     req.Command,
		Script:      req.Script,
		Timeout:     req.Timeout,
		MaxRetries:  req.MaxRetries,
		Parallelism: req.Parallelism,
		TotalHosts:  int32(len(req.HostIDs)),
	}

	if err := h.db.Create(task).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create task")
		return
	}

	// Create task host records
	taskHosts := make([]model.BatchTaskHost, len(req.HostIDs))
	for i, hostID := range req.HostIDs {
		taskHosts[i] = model.BatchTaskHost{
			BatchTaskID: task.ID,
			HostID:      hostID,
			Status:      model.BatchTaskStatusPending,
		}
	}
	if err := h.db.Create(&taskHosts).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create task hosts")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"data": task,
	})
}

// ExecuteBatchTask handles batch task execution requests
func (h *BatchTaskHandler) ExecuteBatchTask(w http.ResponseWriter, r *http.Request) {
	var req model.ExecuteBatchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Get task
	var task model.BatchTask
	if err := h.db.Where("id = ? AND user_id = ?", req.TaskID, userID).First(&task).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Task not found")
		return
	}

	// Get host IDs
	hostIDs := req.HostIDs
	if len(hostIDs) == 0 {
		// Get hosts from task
		var taskHosts []model.BatchTaskHost
		if err := h.db.Where("batch_task_id = ?", task.ID).Find(&taskHosts).Error; err != nil {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get task hosts")
			return
		}
		hostIDs = make([]uuid.UUID, len(taskHosts))
		for i, th := range taskHosts {
			hostIDs[i] = th.HostID
		}
	}

	// Execute task asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
		defer cancel()
		if err := h.taskExecutor.ExecuteTask(ctx, task.ID, hostIDs); err != nil {
			h.logger.Error("task execution failed",
				zap.String("taskId", task.ID.String()),
				zap.Error(err))
		}
	}()

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Task execution started",
		"taskId":  task.ID,
	})
}

// GetBatchTask handles batch task retrieval requests
func (h *BatchTaskHandler) GetBatchTask(w http.ResponseWriter, r *http.Request) {
	// Get task ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 4 || pathParts[3] != "batch-tasks" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	taskID, err := uuid.Parse(pathParts[4])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_TASK_ID", "Invalid task ID")
		return
	}

	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Get task with progress
	response, err := h.taskExecutor.GetTaskProgress(taskID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Task not found")
		return
	}

	// Verify ownership
	if response.BatchTask.UserID != userID {
		respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
		return
	}

	// Eager load hosts
	for i := range response.Hosts {
		h.db.Preload("Host").First(&response.Hosts[i], response.Hosts[i].ID)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": response,
	})
}

// ListBatchTasks handles batch task list requests
func (h *BatchTaskHandler) ListBatchTasks(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build query
	query := h.db.Model(&model.BatchTask{}).Where("user_id = ?", userID)

	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if taskType := r.URL.Query().Get("type"); taskType != "" {
		query = query.Where("type = ?", taskType)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get tasks with pagination
	var tasks []model.BatchTask
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&tasks).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve tasks")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"tasks":    tasks,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CancelBatchTask handles batch task cancellation requests
func (h *BatchTaskHandler) CancelBatchTask(w http.ResponseWriter, r *http.Request) {
	var req model.CancelBatchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Verify task ownership
	var task model.BatchTask
	if err := h.db.Where("id = ? AND user_id = ?", req.TaskID, userID).First(&task).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Task not found")
		return
	}

	// Cancel task
	ctx := context.Background()
	if err := h.taskExecutor.CancelTask(ctx, req.TaskID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "CANCEL_FAILED", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Task cancelled successfully",
		"taskId":  req.TaskID,
	})
}

// DeleteBatchTask handles batch task deletion requests
func (h *BatchTaskHandler) DeleteBatchTask(w http.ResponseWriter, r *http.Request) {
	// Get task ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 4 || pathParts[3] != "batch-tasks" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	taskID, err := uuid.Parse(pathParts[4])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_TASK_ID", "Invalid task ID")
		return
	}

	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Verify task ownership and status
	var task model.BatchTask
	if err := h.db.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Task not found")
		return
	}

	if task.Status == model.BatchTaskStatusRunning {
		respondWithError(w, http.StatusBadRequest, "TASK_RUNNING", "Cannot delete running task")
		return
	}

	// Delete task hosts
	if err := h.db.Where("batch_task_id = ?", taskID).Delete(&model.BatchTaskHost{}).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete task hosts")
		return
	}

	// Delete task
	if err := h.db.Delete(&task).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete task")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Task deleted successfully",
		"taskId":  taskID,
	})
}
