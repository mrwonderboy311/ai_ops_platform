// Package handler provides HTTP handlers for cluster monitoring
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/k8s"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// ClusterMetricsHandler handles cluster metrics operations
type ClusterMetricsHandler struct {
	db *gorm.DB
}

// NewClusterMetricsHandler creates a new cluster metrics handler
func NewClusterMetricsHandler(db *gorm.DB) *ClusterMetricsHandler {
	return &ClusterMetricsHandler{db: db}
}

// GetClusterMetrics handles cluster metrics retrieval requests
func (h *ClusterMetricsHandler) GetClusterMetrics(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 5 || pathParts[4] != "metrics" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
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

	// Verify cluster ownership
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	// Parse query parameters
	durationStr := r.URL.Query().Get("duration")
	duration := time.Hour // default
	if durationStr != "" {
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "INVALID_DURATION", "Invalid duration format")
			return
		}
	}

	// Get start time
	startTime := time.Now().Add(-duration)

	// Query metrics from database
	var metrics []model.ClusterMetric
	if err := h.db.Where("cluster_id = ? AND timestamp >= ?", clusterID, startTime.Unix()).
		Order("timestamp ASC").
		Find(&metrics).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve metrics")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": metrics,
	})
}

// GetClusterMetricsSummary handles cluster metrics summary retrieval requests
func (h *ClusterMetricsHandler) GetClusterMetricsSummary(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 6 || pathParts[5] != "summary" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
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

	// Verify cluster ownership
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	// Get latest metric
	var metric model.ClusterMetric
	if err := h.db.Where("cluster_id = ?", clusterID).
		Order("timestamp DESC").
		First(&metric).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "No metrics available")
		return
	}

	summary := model.ClusterMetricSummary{
		Timestamp:        metric.Timestamp,
		CPUUsagePercent:  metric.CPUUsagePercent,
		PodCount:         metric.PodCount,
		RunningPodCount:  metric.RunningPodCount,
		PendingPodCount:  metric.PendingPodCount,
		FailedPodCount:   metric.FailedPodCount,
		NodeCount:        metric.NodeCount,
		ReadyNodeCount:   metric.ReadyNodeCount,
	}

	if metric.MemoryTotalBytes > 0 {
		summary.MemoryUsagePercent = (float64(metric.MemoryUsageBytes) / float64(metric.MemoryTotalBytes)) * 100
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": summary,
	})
}

// GetNodeMetrics handles node metrics retrieval requests
func (h *ClusterMetricsHandler) GetNodeMetrics(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 6 || pathParts[5] != "metrics" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
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

	// Verify cluster ownership
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	// Parse query parameters
	durationStr := r.URL.Query().Get("duration")
	duration := time.Hour // default
	if durationStr != "" {
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "INVALID_DURATION", "Invalid duration format")
			return
		}
	}

	// Get start time
	startTime := time.Now().Add(-duration)

	// Query metrics from database
	var metrics []model.NodeMetric
	if err := h.db.Where("cluster_id = ? AND timestamp >= ?", clusterID, startTime.Unix()).
		Order("timestamp ASC").
		Find(&metrics).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve metrics")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": metrics,
	})
}

// GetLiveClusterMetrics handles live cluster metrics retrieval from cluster
func (h *ClusterMetricsHandler) GetLiveClusterMetrics(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 6 || pathParts[5] != "live" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
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

	// Verify cluster ownership
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	// Create cluster client and metrics client
	config := &k8s.ClusterConfig{
		Kubeconfig: []byte(cluster.Kubeconfig),
		Endpoint:   cluster.Endpoint,
	}

	client, err := k8s.NewClusterClient(config)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CLIENT_ERROR", "Failed to create cluster client")
		return
	}
	defer client.Close()

	// Get live metrics (requires metrics server)
	// For now, return cluster info without detailed metrics
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.GetClusterInfo(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch cluster metrics")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": info,
	})
}

// GetLiveNodeMetrics handles live node metrics retrieval from cluster
func (h *ClusterMetricsHandler) GetLiveNodeMetrics(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 6 || pathParts[5] != "live-metrics" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
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

	// Verify cluster ownership
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	// Create cluster client
	config := &k8s.ClusterConfig{
		Kubeconfig: []byte(cluster.Kubeconfig),
		Endpoint:   cluster.Endpoint,
	}

	client, err := k8s.NewClusterClient(config)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CLIENT_ERROR", "Failed to create cluster client")
		return
	}
	defer client.Close()

	// Get nodes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nodes, err := client.GetNodes(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch node metrics")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": nodes,
	})
}

// GetPodMetrics handles pod metrics retrieval requests
func (h *ClusterMetricsHandler) GetPodMetrics(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID and namespace from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 7 {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
		return
	}

	namespace := pathParts[5]

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

	// Verify cluster ownership
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	// Parse query parameters
	durationStr := r.URL.Query().Get("duration")
	duration := time.Hour // default
	if durationStr != "" {
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "INVALID_DURATION", "Invalid duration format")
			return
		}
	}

	// Get start time
	startTime := time.Now().Add(-duration)

	// Query metrics from database
	var metrics []model.PodMetric
	if err := h.db.Where("cluster_id = ? AND namespace = ? AND timestamp >= ?", clusterID, namespace, startTime.Unix()).
		Order("timestamp ASC").
		Find(&metrics).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve metrics")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": metrics,
	})
}

// ListNamespaces handles namespace list requests
func (h *ClusterMetricsHandler) ListNamespaces(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 5 || pathParts[4] != "namespaces" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
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

	// Verify cluster ownership
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	// Create cluster client
	config := &k8s.ClusterConfig{
		Kubeconfig: []byte(cluster.Kubeconfig),
		Endpoint:   cluster.Endpoint,
	}

	client, err := k8s.NewClusterClient(config)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CLIENT_ERROR", "Failed to create cluster client")
		return
	}
	defer client.Close()

	// Get namespaces
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	namespaces, err := client.GetNamespaces(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch namespaces")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": namespaces,
	})
}

// RefreshMetrics triggers a metrics refresh for a cluster
func (h *ClusterMetricsHandler) RefreshMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 5 || pathParts[4] != "refresh" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
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

	// Verify cluster ownership
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	// Trigger background metrics refresh
	go h.collectMetrics(clusterID)

	respondWithJSON(w, http.StatusAccepted, map[string]interface{}{
		"message": "Metrics refresh started",
	})
}

// collectMetrics collects and stores metrics for a cluster
func (h *ClusterMetricsHandler) collectMetrics(clusterID uuid.UUID) {
	var cluster model.K8sCluster
	if err := h.db.Where("id = ?", clusterID).First(&cluster).Error; err != nil {
		return
	}

	config := &k8s.ClusterConfig{
		Kubeconfig: []byte(cluster.Kubeconfig),
		Endpoint:   cluster.Endpoint,
	}

	client, err := k8s.NewClusterClient(config)
	if err != nil {
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get cluster info for basic metrics
	info, err := client.GetClusterInfo(ctx)
	if err != nil {
		return
	}

	// Create cluster metric record
	timestamp := time.Now().Unix()
	clusterMetric := &model.ClusterMetric{
		ID:              uuid.New(),
		ClusterID:       clusterID,
		Timestamp:       timestamp,
		PodCount:        info.PodCount,
		NodeCount:       info.NodeCount,
	}

	// Store metric (note: CPU/Memory usage requires metrics server which may not be installed)
	h.db.Create(clusterMetric)

	// Get and store node metrics
	nodes, _ := client.GetNodes(ctx)
	for _, node := range nodes {
		nodeMetric := &model.NodeMetric{
			ID:        uuid.New(),
			ClusterID: clusterID,
			NodeName:  node.Name,
			Timestamp: timestamp,
			Status:    node.Status,
			Ready:     node.Status == "Ready",
			PodCount:  0, // Will be calculated if we have pod info
		}
		h.db.Create(nodeMetric)
	}
}
