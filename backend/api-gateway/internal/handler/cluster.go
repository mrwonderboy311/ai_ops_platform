// Package handler provides HTTP handlers for Kubernetes cluster management
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/k8s"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// ClusterHandler handles Kubernetes cluster operations
type ClusterHandler struct {
	db *gorm.DB
}

// NewClusterHandler creates a new cluster handler
func NewClusterHandler(db *gorm.DB) *ClusterHandler {
	return &ClusterHandler{db: db}
}

// CreateCluster handles cluster creation requests
func (h *ClusterHandler) CreateCluster(w http.ResponseWriter, r *http.Request) {
	var req model.CreateClusterRequest
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

	// Test connection before creating
	config := &k8s.ClusterConfig{
		Kubeconfig: []byte(req.Kubeconfig),
		Endpoint:   req.Endpoint,
	}

	client, err := k8s.NewClusterClient(config)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CONFIG", "Failed to create cluster client")
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.TestConnection(ctx)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "CONNECTION_FAILED", "Failed to connect to cluster")
		return
	}

	// Create cluster record
	cluster := &model.K8sCluster{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Status:      model.ClusterStatusConnected,
		Endpoint:    req.Endpoint,
		Kubeconfig:  req.Kubeconfig, // TODO: Encrypt this
		Version:     info.Version,
		NodeCount:   info.NodeCount,
		Region:      req.Region,
		Provider:    req.Provider,
	}

	now := time.Now()
	cluster.LastConnectedAt = &now

	if err := h.db.Create(cluster).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create cluster")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"data": cluster,
	})
}

// TestConnection handles cluster connection test requests
func (h *ClusterHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	var req model.ClusterConnectionTestRequest
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

	// Test connection
	config := &k8s.ClusterConfig{
		Kubeconfig: []byte(req.Kubeconfig),
		Endpoint:   req.Endpoint,
	}

	client, err := k8s.NewClusterClient(config)
	if err != nil {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"data": model.ClusterConnectionTestResponse{
				Success: false,
				Error:   err.Error(),
			},
		})
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.TestConnection(ctx)
	if err != nil {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"data": model.ClusterConnectionTestResponse{
				Success: false,
				Error:   err.Error(),
			},
		})
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": model.ClusterConnectionTestResponse{
			Success:   true,
			Version:   info.Version,
			NodeCount: info.NodeCount,
		},
	})
}

// GetCluster handles cluster retrieval requests
func (h *ClusterHandler) GetCluster(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 4 || pathParts[3] != "clusters" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[4])
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

	// Get cluster
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": cluster,
	})
}

// ListClusters handles cluster list requests
func (h *ClusterHandler) ListClusters(w http.ResponseWriter, r *http.Request) {
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
	query := h.db.Model(&model.K8sCluster{}).Where("user_id = ?", userID)

	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if clusterType := r.URL.Query().Get("type"); clusterType != "" {
		query = query.Where("type = ?", clusterType)
	}
	if provider := r.URL.Query().Get("provider"); provider != "" {
		query = query.Where("provider = ?", provider)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get clusters with pagination
	var clusters []model.K8sCluster
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&clusters).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve clusters")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"clusters": clusters,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// UpdateCluster handles cluster update requests
func (h *ClusterHandler) UpdateCluster(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 4 || pathParts[3] != "clusters" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[4])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_CLUSTER_ID", "Invalid cluster ID")
		return
	}

	var req model.UpdateClusterRequest
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

	// Get cluster
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Endpoint != "" {
		updates["endpoint"] = req.Endpoint
	}
	if req.Kubeconfig != "" {
		updates["kubeconfig"] = req.Kubeconfig // TODO: Encrypt this
	}

	if err := h.db.Model(&cluster).Updates(updates).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update cluster")
		return
	}

	// Refresh cluster info
	if req.Kubeconfig != "" || req.Endpoint != "" {
		go h.refreshClusterInfo(clusterID)
	}

	// Get updated cluster
	h.db.Where("id = ?", clusterID).First(&cluster)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": cluster,
	})
}

// DeleteCluster handles cluster deletion requests
func (h *ClusterHandler) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 4 || pathParts[3] != "clusters" {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid URL path")
		return
	}

	clusterID, err := uuid.Parse(pathParts[4])
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
	var cluster model.Cluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	// Delete related records
	h.db.Where("cluster_id = ?", clusterID).Delete(&model.ClusterNode{})
	h.db.Where("cluster_id = ?", clusterID).Delete(&model.ClusterNamespace{})

	// Delete cluster
	if err := h.db.Delete(&cluster).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete cluster")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Cluster deleted successfully",
		"clusterId": clusterID,
	})
}

// GetClusterNodes handles cluster nodes retrieval requests
func (h *ClusterHandler) GetClusterNodes(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 5 || pathParts[4] != "nodes" {
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
	var cluster model.Cluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterID, userID).First(&cluster).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster not found")
		return
	}

	// Get nodes from database
	var nodes []model.ClusterNode
	if err := h.db.Where("cluster_id = ?", clusterID).Find(&nodes).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve nodes")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": nodes,
	})
}

// GetClusterInfo handles cluster info retrieval requests
func (h *ClusterHandler) GetClusterInfo(w http.ResponseWriter, r *http.Request) {
	// Get cluster ID from URL path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 5 || pathParts[4] != "info" {
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

	// Create cluster client and get info
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.GetClusterInfo(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch cluster info")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": info,
	})
}

// refreshClusterInfo refreshes cluster information in the background
func (h *ClusterHandler) refreshClusterInfo(clusterID uuid.UUID) {
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
		h.updateClusterStatus(clusterID, model.ClusterStatusError, err.Error())
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.GetClusterInfo(ctx)
	if err != nil {
		h.updateClusterStatus(clusterID, model.ClusterStatusError, err.Error())
		return
	}

	// Update cluster info
	now := time.Now()
	h.db.Model(&model.K8sCluster{}).Where("id = ?", clusterID).Updates(map[string]interface{}{
		"status":          model.ClusterStatusConnected,
		"version":         info.Version,
		"nodeCount":        info.NodeCount,
		"lastConnectedAt":  &now,
		"errorMessage":     "",
	})

	// Update nodes
	nodes, _ := client.GetNodes(ctx)
	h.updateClusterNodes(clusterID, nodes)
}

// updateClusterStatus updates the cluster status
func (h *ClusterHandler) updateClusterStatus(clusterID uuid.UUID, status model.ClusterStatus, errMsg string) {
	h.db.Model(&model.K8sCluster{}).Where("id = ?", clusterID).Updates(map[string]interface{}{
		"status":       status,
		"errorMessage": errMsg,
	})
}

// updateClusterNodes updates cluster nodes
func (h *ClusterHandler) updateClusterNodes(clusterID uuid.UUID, nodes []k8s.NodeInfo) {
	// Delete existing nodes
	h.db.Where("cluster_id = ?", clusterID).Delete(&model.ClusterNode{})

	// Insert new nodes
	for _, node := range nodes {
		clusterNode := &model.ClusterNode{
			ID:               uuid.New(),
			ClusterID:        clusterID,
			Name:             node.Name,
			InternalIP:       node.InternalIP,
			ExternalIP:       node.ExternalIP,
			Status:           node.Status,
			Roles:            node.Roles,
			Version:          node.Version,
			OSImage:          node.OSImage,
			KernelVersion:    node.KernelVersion,
			ContainerRuntime: node.ContainerRuntime,
			CPUCapacity:      node.CPUCapacity,
			MemoryCapacity:   node.MemoryCapacity,
			StorageCapacity:  node.StorageCapacity,
			CPUAllocatable:   node.CPUAllocatable,
			MemoryAllocatable: node.MemoryAllocatable,
			StorageAllocatable: node.StorageAllocatable,
		}
		h.db.Create(clusterNode)
	}
}
