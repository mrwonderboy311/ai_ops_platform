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

// GetPodDetail handles pod detail requests
func (h *WorkloadHandler) GetPodDetail(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID, namespace, and pod name from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 7 || pathParts[6] != "detail" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[3])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
		return
	}

	namespace := pathParts[5]
	podName := pathParts[7] // Skip "detail" and get pod name

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

	// Get pod detail
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	podDetail, err := client.GetPodDetail(ctx, namespace, podName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch pod details")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": podDetail,
	})
}

// Helper functions for Kubernetes operations

func getDeployments(ctx context.Context, client *k8s.ClusterClient, namespace string) ([]map[string]interface{}, error) {
	deployments, err := client.GetDeployments(ctx, namespace)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(deployments))
	for i, d := range deployments {
		result[i] = map[string]interface{}{
			"name":              d.Name,
			"namespace":         d.Namespace,
			"replicas":          d.Replicas,
			"readyReplicas":     d.ReadyReplicas,
			"updatedReplicas":   d.UpdatedReplicas,
			"availableReplicas": d.AvailableReplicas,
			"image":             d.Image,
			"createdAt":         d.CreatedAt,
		}
	}
	return result, nil
}

func getPods(ctx context.Context, client *k8s.ClusterClient, namespace string) ([]map[string]interface{}, error) {
	pods, err := client.GetPods(ctx, namespace)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(pods))
	for i, p := range pods {
		result[i] = map[string]interface{}{
			"name":         p.Name,
			"namespace":    p.Namespace,
			"status":       p.Status,
			"phase":        p.Phase,
			"podIp":        p.PodIP,
			"nodeName":     p.NodeName,
			"ready":        p.Ready,
			"restartCount": p.RestartCount,
			"ownerType":    p.OwnerType,
			"ownerName":    p.OwnerName,
			"createdAt":    p.CreatedAt,
		}
	}
	return result, nil
}

func getServices(ctx context.Context, client *k8s.ClusterClient, namespace string) ([]map[string]interface{}, error) {
	services, err := client.GetServices(ctx, namespace)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(services))
	for i, s := range services {
		result[i] = map[string]interface{}{
			"name":       s.Name,
			"namespace":  s.Namespace,
			"type":       s.Type,
			"clusterIp":  s.ClusterIP,
			"ports":      s.Ports,
			"createdAt":  s.CreatedAt,
		}
	}
	return result, nil
}

func getPodLogs(ctx context.Context, client *k8s.ClusterClient, namespace, podName string, tailLines int64) (string, error) {
	return client.GetPodLogs(ctx, namespace, podName, tailLines)
}

func deletePod(ctx context.Context, client *k8s.ClusterClient, namespace, podName string) error {
	return client.DeletePod(ctx, namespace, podName)
}
