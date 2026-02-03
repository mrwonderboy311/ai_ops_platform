// Package handler provides HTTP handlers for Helm repository management
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// HelmHandler handles Helm repository operations
type HelmHandler struct {
	db *gorm.DB
}

// NewHelmHandler creates a new Helm handler
func NewHelmHandler(db *gorm.DB) *HelmHandler {
	return &HelmHandler{db: db}
}

// CreateHelmRepo creates a new Helm repository
func (h *HelmHandler) CreateHelmRepo(w http.ResponseWriter, r *http.Request) {
	var req model.CreateHelmRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Check if repository name already exists for this user
	var existingRepo model.HelmRepository
	if err := h.db.Where("user_id = ? AND name = ?", userUUID, req.Name).First(&existingRepo).Error; err == nil {
		respondWithError(w, http.StatusConflict, "CONFLICT", "Repository name already exists")
		return
	}

	// Create repository
	repo := model.HelmRepository{
		ID:              uuid.New(),
		UserID:          userUUID,
		Name:            req.Name,
		Description:     req.Description,
		Type:            req.Type,
		Status:          model.HelmRepoStatusActive,
		URL:             req.URL,
		Username:        req.Username,
		Password:        req.Password, // Should be encrypted in production
		CAFile:          req.CAFile,
		CertFile:        req.CertFile,
		KeyFile:         req.KeyFile,
		InsecureSkipTLS: req.InsecureSkipTLS,
	}

	if err := h.db.Create(&repo).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create repository")
		return
	}

	respondWithJSON(w, http.StatusCreated, repo)
}

// ListHelmRepos lists all Helm repositories for the authenticated user
func (h *HelmHandler) ListHelmRepos(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Parse query parameters
	page := 1
	pageSize := 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Build query
	query := h.db.Model(&model.HelmRepository{}).Where("user_id = ?", userUUID)

	// Apply filters
	if status := r.URL.Query().Get("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if repoType := r.URL.Query().Get("type"); repoType != "" {
		query = query.Where("type = ?", repoType)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Fetch repositories
	var repos []model.HelmRepository
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&repos).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch repositories")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":       repos,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetHelmRepo gets a specific Helm repository
func (h *HelmHandler) GetHelmRepo(w http.ResponseWriter, r *http.Request) {
	// Extract repo ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid repository ID")
		return
	}

	repoID := parts[3]
	repoUUID, err := uuid.Parse(repoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid repository ID format")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Fetch repository
	var repo model.HelmRepository
	if err := h.db.Where("id = ? AND user_id = ?", repoUUID, userUUID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Repository not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch repository")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, repo)
}

// UpdateHelmRepo updates a Helm repository
func (h *HelmHandler) UpdateHelmRepo(w http.ResponseWriter, r *http.Request) {
	// Extract repo ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid repository ID")
		return
	}

	repoID := parts[3]
	repoUUID, err := uuid.Parse(repoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid repository ID format")
		return
	}

	var req model.UpdateHelmRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Fetch repository
	var repo model.HelmRepository
	if err := h.db.Where("id = ? AND user_id = ?", repoUUID, userUUID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Repository not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch repository")
		}
		return
	}

	// Update fields
	updates := map[string]interface{}{}
	if req.Name != "" {
		// Check if new name conflicts
		var existingRepo model.HelmRepository
		if err := h.db.Where("user_id = ? AND name = ? AND id != ?", userUUID, req.Name, repoUUID).First(&existingRepo).Error; err == nil {
			respondWithError(w, http.StatusConflict, "CONFLICT", "Repository name already exists")
			return
		}
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.URL != "" {
		updates["url"] = req.URL
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	if req.CAFile != "" {
		updates["ca_file"] = req.CAFile
	}
	if req.CertFile != "" {
		updates["cert_file"] = req.CertFile
	}
	if req.KeyFile != "" {
		updates["key_file"] = req.KeyFile
	}
	updates["insecure_skip_tls"] = req.InsecureSkipTLS

	if err := h.db.Model(&repo).Updates(updates).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update repository")
		return
	}

	// Fetch updated repo
	h.db.First(&repo, repoUUID)
	respondWithJSON(w, http.StatusOK, repo)
}

// DeleteHelmRepo deletes a Helm repository
func (h *HelmHandler) DeleteHelmRepo(w http.ResponseWriter, r *http.Request) {
	// Extract repo ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid repository ID")
		return
	}

	repoID := parts[3]
	repoUUID, err := uuid.Parse(repoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid repository ID format")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Check ownership
	var repo model.HelmRepository
	if err := h.db.Where("id = ? AND user_id = ?", repoUUID, userUUID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Repository not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch repository")
		}
		return
	}

	// Delete repository
	if err := h.db.Delete(&repo).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete repository")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Repository deleted successfully",
	})
}

// TestHelmRepo tests a Helm repository connection
func (h *HelmHandler) TestHelmRepo(w http.ResponseWriter, r *http.Request) {
	var req model.HelmRepoTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// TODO: Implement actual Helm repository connection test
	// This would use the Helm Go SDK to connect and test the repository
	// For now, return a mock response

	respondWithJSON(w, http.StatusOK, model.HelmRepoTestResponse{
		Success:    true,
		ChartCount: 0,
		Message:    "Connection test not yet implemented",
	})
}

// SyncHelmRepo syncs charts from a Helm repository
func (h *HelmHandler) SyncHelmRepo(w http.ResponseWriter, r *http.Request) {
	// Extract repo ID from URL path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request")
		return
	}

	repoID := parts[4]
	repoUUID, err := uuid.Parse(repoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid repository ID format")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user ID format")
		return
	}

	// Fetch repository
	var repo model.HelmRepository
	if err := h.db.Where("id = ? AND user_id = ?", repoUUID, userUUID).First(&repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Repository not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch repository")
		}
		return
	}

	// TODO: Implement actual Helm repository sync
	// This would use the Helm Go SDK to fetch chart index
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Sync not yet implemented",
	})
}
