// Package handler provides HTTP handlers for user and role management
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/api-gateway/internal/service"
	"github.com/wangjialin/myops/pkg/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserManagementHandler handles user management requests
type UserManagementHandler struct {
	db          *gorm.DB
	userService *service.UserService
	logger      *zap.Logger
}

// NewUserManagementHandler creates a new user management handler
func NewUserManagementHandler(db *gorm.DB, logger *zap.Logger) *UserManagementHandler {
	return &UserManagementHandler{
		db:          db,
		userService: service.NewUserService(db, logger),
		logger:      logger,
	}
}

// ListUsers handles user list retrieval
func (h *UserManagementHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	search := r.URL.Query().Get("search")

	users, total, err := h.userService.ListUsers(page, pageSize, search)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve users")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"users":    users,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// GetUserByID handles single user retrieval
func (h *UserManagementHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := extractIDFromPath(r.URL.Path, "/api/v1/users/")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve user")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": user,
	})
}

// CreateUser handles user creation
func (h *UserManagementHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	err := h.userService.CreateUser(r.Context(), &user)
	if err != nil {
		if err.Error() == "user with this username or email already exists" {
			respondWithError(w, http.StatusConflict, "CONFLICT", "Username or email already exists")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create user")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"data": user,
	})
}

// UpdateUser handles user update
func (h *UserManagementHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := extractIDFromPath(r.URL.Path, "/api/v1/users/")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	err = h.userService.UpdateUser(id, updates)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user")
		}
		return
	}

	// Fetch updated user
	user, _ := h.userService.GetUserByID(id)
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": user,
	})
}

// DeleteUser handles user deletion
func (h *UserManagementHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := extractIDFromPath(r.URL.Path, "/api/v1/users/")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID")
		return
	}

	err = h.userService.DeleteUser(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete user")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"success": true,
		},
	})
}

// GetUserRoles handles retrieval of user's roles
func (h *UserManagementHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID := extractIDFromPath(r.URL.Path, "/api/v1/users/")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": user.Roles,
	})
}

// AssignRoleToUser handles role assignment to user
func (h *UserManagementHandler) AssignRoleToUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RoleID string `json:"roleId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	userID := extractIDFromPath(r.URL.Path, "/api/v1/users/")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid role ID")
		return
	}

	// Get current user from context
	var createdBy uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			createdBy, _ = uuid.Parse(uid)
		}
	}

	err = h.userService.AssignRoleToUser(uid, roleID, createdBy)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to assign role")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"success": true,
		},
	})
}

// RemoveRoleFromUser handles role removal from user
func (h *UserManagementHandler) RemoveRoleFromUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RoleID string `json:"roleId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	userID := extractIDFromPath(r.URL.Path, "/api/v1/users/")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid role ID")
		return
	}

	err = h.userService.RemoveRoleFromUser(uid, roleID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove role")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"success": true,
		},
	})
}

// ListRoles handles role list retrieval
func (h *UserManagementHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.userService.ListRoles()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve roles")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"roles": roles,
			"total": len(roles),
		},
	})
}

// GetRoleByID handles single role retrieval
func (h *UserManagementHandler) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	roleID := extractIDFromPath(r.URL.Path, "/api/v1/roles/")
	if roleID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Role ID is required")
		return
	}

	id, err := uuid.Parse(roleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid role ID")
		return
	}

	role, err := h.userService.GetRoleByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Role not found")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": role,
	})
}

// CreateRole handles role creation
func (h *UserManagementHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var role model.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	err := h.userService.CreateRole(r.Context(), &role)
	if err != nil {
		if err.Error() == "role with this name already exists" {
			respondWithError(w, http.StatusConflict, "CONFLICT", "Role name already exists")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create role")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"data": role,
	})
}

// UpdateRole handles role update
func (h *UserManagementHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	roleID := extractIDFromPath(r.URL.Path, "/api/v1/roles/")
	if roleID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Role ID is required")
		return
	}

	id, err := uuid.Parse(roleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid role ID")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	err = h.userService.UpdateRole(id, updates)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Role not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update role")
		}
		return
	}

	// Fetch updated role
	role, _ := h.userService.GetRoleByID(id)
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": role,
	})
}

// DeleteRole handles role deletion
func (h *UserManagementHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleID := extractIDFromPath(r.URL.Path, "/api/v1/roles/")
	if roleID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Role ID is required")
		return
	}

	id, err := uuid.Parse(roleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid role ID")
		return
	}

	err = h.userService.DeleteRole(id)
	if err != nil {
		if err.Error() == "cannot delete system role" {
			respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Cannot delete system role")
		} else if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Role not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete role")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"success": true,
		},
	})
}

// ListPermissions handles permission list retrieval
func (h *UserManagementHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.userService.ListPermissions()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve permissions")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"permissions": permissions,
			"total":       len(permissions),
		},
	})
}

// AssignPermissionToRole handles permission assignment to role
func (h *UserManagementHandler) AssignPermissionToRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PermissionID string `json:"permissionId"`
	}

	roleID := extractIDFromPath(r.URL.Path, "/api/v1/roles/")
	if roleID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Role ID is required")
		return
	}

	rid, err := uuid.Parse(roleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid role ID")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	pid, err := uuid.Parse(req.PermissionID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid permission ID")
		return
	}

	// Get current user from context
	var createdBy uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			createdBy, _ = uuid.Parse(uid)
		}
	}

	err = h.userService.AssignPermissionToRole(rid, pid, createdBy)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to assign permission")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"success": true,
		},
	})
}

// RemovePermissionFromRole handles permission removal from role
func (h *UserManagementHandler) RemovePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PermissionID string `json:"permissionId"`
	}

	roleID := extractIDFromPath(r.URL.Path, "/api/v1/roles/")
	if roleID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Role ID is required")
		return
	}

	rid, err := uuid.Parse(roleID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid role ID")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	pid, err := uuid.Parse(req.PermissionID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid permission ID")
		return
	}

	err = h.userService.RemovePermissionFromRole(rid, pid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove permission")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"success": true,
		},
	})
}

// CheckPermission handles permission check
func (h *UserManagementHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
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

	resource := r.URL.Query().Get("resource")
	action := r.URL.Query().Get("action")

	if resource == "" || action == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Resource and action are required")
		return
	}

	hasPermission, err := h.userService.CheckPermission(userID, resource, action)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check permission")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"hasPermission": hasPermission,
		},
	})
}
