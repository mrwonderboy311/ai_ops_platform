// Package handler provides HTTP handlers for performance monitoring
package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/wangjialin/myops/api-gateway/internal/service"
	"github.com/wangjialin/myops/pkg/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PerformanceHandler handles performance monitoring requests
type PerformanceHandler struct {
	db          *gorm.DB
	perfService *service.PerformanceService
	logger      *zap.Logger
}

// NewPerformanceHandler creates a new performance handler
func NewPerformanceHandler(db *gorm.DB, logger *zap.Logger) *PerformanceHandler {
	return &PerformanceHandler{
		db:          db,
		perfService: service.NewPerformanceService(db, logger),
		logger:      logger,
	}
}

// GetMetrics handles performance metrics retrieval
func (h *PerformanceHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metricType := r.URL.Query().Get("metricType")
	entityType := r.URL.Query().Get("entityType")
	entityID := r.URL.Query().Get("entityId")

	var startTime, endTime int64
	if startTimeStr := r.URL.Query().Get("startTime"); startTimeStr != "" {
		startTime, _ = strconv.ParseInt(startTimeStr, 10, 64)
	}
	if endTimeStr := r.URL.Query().Get("endTime"); endTimeStr != "" {
		endTime, _ = strconv.ParseInt(endTimeStr, 10, 64)
	}

	// Default to last hour if no time range specified
	if startTime == 0 {
		startTime = time.Now().Unix() - 3600
	}
	if endTime == 0 {
		endTime = time.Now().Unix()
	}

	metrics, err := h.perfService.GetMetrics(metricType, entityType, entityID, startTime, endTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve metrics")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"metrics": metrics,
			"total":   len(metrics),
		},
	})
}

// GetSystemHealth handles system health status retrieval
func (h *PerformanceHandler) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.perfService.GetSystemHealth()
	if err != nil {
		// Calculate if not exists
		health, err = h.perfService.CalculateSystemHealth()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve system health")
			return
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": health,
	})
}

// RefreshSystemHealth triggers system health recalculation
func (h *PerformanceHandler) RefreshSystemHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.perfService.CalculateSystemHealth()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to calculate system health")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": health,
	})
}

// GetPerformanceSummary handles performance summary retrieval
func (h *PerformanceHandler) GetPerformanceSummary(w http.ResponseWriter, r *http.Request) {
	timeWindow := r.URL.Query().Get("timeWindow")
	if timeWindow == "" {
		timeWindow = "5m"
	}

	var duration time.Duration
	switch timeWindow {
	case "5m":
		duration = 5 * time.Minute
	case "15m":
		duration = 15 * time.Minute
	case "1h":
		duration = 1 * time.Hour
	case "1d":
		duration = 24 * time.Hour
	default:
		duration = 5 * time.Minute
	}

	endTime := time.Now().Unix()
	startTime := time.Now().Add(-duration).Unix()

	snapshot, err := h.perfService.AggregateSnapshot(timeWindow, startTime, endTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to aggregate performance data")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": snapshot,
	})
}

// GetTrendData handles trend data retrieval
func (h *PerformanceHandler) GetTrendData(w http.ResponseWriter, r *http.Request) {
	metricType := r.URL.Query().Get("metricType")
	if metricType == "" {
		metricType = "cpu"
	}

	entityType := r.URL.Query().Get("entityType")

	points := 100
	if pointsStr := r.URL.Query().Get("points"); pointsStr != "" {
		if p, err := strconv.Atoi(pointsStr); err == nil && p > 0 && p <= 1000 {
			points = p
		}
	}

	trend, err := h.perfService.GetTrendData(metricType, entityType, points)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve trend data")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"metricType": metricType,
			"points":     trend,
		},
	})
}

// CollectMetric handles metric collection (internal endpoint)
func (h *PerformanceHandler) CollectMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	body, _ := io.ReadAll(r.Body)
	var metric model.PerformanceMetric
	if err := json.Unmarshal(body, &metric); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.perfService.CollectMetric(&metric); err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to collect metric")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"data": metric,
	})
}

// GetMetricStatistics handles metric statistics retrieval
func (h *PerformanceHandler) GetMetricStatistics(w http.ResponseWriter, r *http.Request) {
	metricType := r.URL.Query().Get("metricType")
	if metricType == "" {
		metricType = "cpu"
	}

	duration := 3600 // default 1 hour
	if durationStr := r.URL.Query().Get("duration"); durationStr != "" {
		if d, err := strconv.Atoi(durationStr); err == nil {
			duration = d
		}
	}

	startTime := time.Now().Unix() - int64(duration)

	var stats struct {
		Avg   float64
		Min   float64
		Max   float64
		Count int64
	}

	err := h.db.Raw(`
		SELECT AVG(value) as avg, MIN(value) as min, MAX(value) as max, COUNT(*) as count
		FROM performance_metrics
		WHERE metric_type = ? AND timestamp >= ?
	`, metricType, startTime).Scan(&stats).Error

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve statistics")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"metricType": metricType,
			"avg":        stats.Avg,
			"min":        stats.Min,
			"max":        stats.Max,
			"count":      stats.Count,
		},
	})
}
