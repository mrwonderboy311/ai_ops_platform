// Package handler provides HTTP handlers for Kubernetes workload management
package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/k8s"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// WorkloadHandler handles Kubernetes workload operations
type WorkloadHandler struct {
	db *gorm.DB
}

// NewWorkloadHandler creates a new workload handler
func NewWorkloadHandler(db *gorm.DB) *WorkloadHandler {
	return &WorkloadHandler{db: db}
}

// ListNamespaces handles namespace list requests
func (h *WorkloadHandler) ListNamespaces(w http.ResponseWriter, r *http.Request) {
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

// ListDeployments handles deployment list requests
func (h *WorkloadHandler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID and namespace from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 6 || pathParts[4] != "namespaces" || pathParts[6] != "deployments" {
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

	// Get deployments
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deployments, err := getDeployments(ctx, client, namespace)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch deployments")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": deployments,
	})
}

// ListPods handles pod list requests
func (h *WorkloadHandler) ListPods(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID and namespace from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 6 || pathParts[4] != "namespaces" || pathParts[6] != "pods" {
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

	// Get pods
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pods, err := getPods(ctx, client, namespace)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch pods")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": pods,
	})
}

// ListServices handles service list requests
func (h *WorkloadHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID and namespace from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 6 || pathParts[4] != "namespaces" || pathParts[6] != "services" {
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

	// Get services
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	services, err := getServices(ctx, client, namespace)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch services")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": services,
	})
}

// GetPodLogs handles pod log requests
func (h *WorkloadHandler) GetPodLogs(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID, namespace, and pod name from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 8 || pathParts[6] != "pods" || pathParts[8] != "logs" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
		return
	}

	namespace := pathParts[5]
	podName := pathParts[7]

	// Get query parameters
	tailLines := int64(100)
	if tailStr := r.URL.Query().Get("tailLines"); tailStr != "" {
		if tail, err := strconv.Atoi(tailStr); err == nil && tail > 0 {
			tailLines = int64(tail)
		}
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

	// Get pod logs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logs, err := getPodLogs(ctx, client, namespace, podName, tailLines)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch pod logs")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"podName": podName,
			"logs":    logs,
		},
	})
}

// DeletePod handles pod deletion requests
func (h *WorkloadHandler) DeletePod(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Get cluster ID, namespace, and pod name from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 8 || pathParts[6] != "pods" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
		return
	}

	namespace := pathParts[5]
	podName := pathParts[7]

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

	// Delete pod
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := deletePod(ctx, client, namespace, podName); err != nil {
		respondWithError(w, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete pod")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Pod deleted successfully",
		"podName": podName,
	})
}

// Helper functions for Kubernetes operations

func getDeployments(ctx context.Context, client *k8s.ClusterClient, namespace string) ([]map[string]interface{}, error) {
	// This would use the Kubernetes client-go to list deployments
	// For now, return a simplified response
	return []map[string]interface{}{}, nil
}

func getPods(ctx context.Context, client *k8s.ClusterClient, namespace string) ([]map[string]interface{}, error) {
	// This would use the Kubernetes client-go to list pods
	// For now, return a simplified response
	return []map[string]interface{}{}, nil
}

func getServices(ctx context.Context, client *k8s.ClusterClient, namespace string) ([]map[string]interface{}, error) {
	// This would use the Kubernetes client-go to list services
	// For now, return a simplified response
	return []map[string]interface{}{}, nil
}

func getPodLogs(ctx context.Context, client *k8s.ClusterClient, namespace, podName string, tailLines int64) (string, error) {
	// This would use the Kubernetes client-go to get pod logs
	// For now, return a simplified response
	return "Logs not available - client-go integration needed", nil
}

func deletePod(ctx context.Context, client *k8s.ClusterClient, namespace, podName string) error {
	// This would use the Kubernetes client-go to delete the pod
	// For now, return nil as placeholder
	return nil
}
