// Package handler provides HTTP handlers for RBAC operations
package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"myops.k8s.io/backend/pkg/model"
)

// RBACHandler handles RBAC operations
type RBACHandler struct {
	db *gorm.DB
}

// NewRBACHandler creates a new RBAC handler
func NewRBACHandler(db *gorm.DB) *RBACHandler {
	return &RBACHandler{db: db}
}

// Permission Requests/Responses

type ListPermissionsRequest struct {
	Category string `form:"category"`
	Resource string `form:"resource"`
	Action   string `form:"action"`
	Scope    string `form:"scope"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"pageSize" binding:"min=1,max=100"`
}

type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"displayName" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category" binding:"required,oneof=host k8s observability ai system"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required,oneof=create read update delete execute admin"`
	Scope       string `json:"scope" binding:"required,oneof=global cluster namespace host"`
}

type UpdatePermissionRequest struct {
	DisplayName *string `json:"displayName"`
	Description *string `json:"description"`
}

// Role Requests/Responses

type ListRolesRequest struct {
	IsSystem *bool  `form:"isSystem"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"pageSize" binding:"min=1,max=100"`
}

type CreateRoleRequest struct {
	Name        string     `json:"name" binding:"required"`
	DisplayName string     `json:"displayName" binding:"required"`
	Description string     `json:"description"`
	IsDefault   bool       `json:"isDefault"`
	ParentID    *uuid.UUID `json:"parentId"`
}

type UpdateRoleRequest struct {
	DisplayName *string `json:"displayName"`
	Description *string `json:"description"`
	IsDefault   *bool   `json:"isDefault"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []uuid.UUID `json:"permissionIds" binding:"required,min=1"`
	Override      bool        `json:"override"` // If true, replace existing permissions
}

// User Role Requests/Responses

type AssignUserRoleRequest struct {
	RoleID       uuid.UUID  `json:"roleId" binding:"required"`
	ResourceID   *uuid.UUID `json:"resourceId"`
	ResourceType string     `json:"resourceType" binding:"omitempty,oneof=cluster namespace host"`
	ExpiresAt    *time.Time `json:"expiresAt"`
}

type RemoveUserRoleRequest struct {
	RoleID uuid.UUID `json:"roleId" binding:"required"`
}

type CheckPermissionRequest struct {
	Resource     string     `json:"resource" binding:"required"`
	Action       string     `json:"action" binding:"required"`
	ResourceID   *uuid.UUID `json:"resourceId"`
	ResourceType string     `json:"resourceType"`
}

// Resource Access Policy Requests/Responses

type CreateResourceAccessPolicyRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	Effect      string     `json:"effect" binding:"required,oneof=allow deny"`
	Action      string     `json:"action" binding:"required"`
	Resource    string     `json:"resource" binding:"required"`
	Selector    string     `json:"selector"`
	Enabled     bool       `json:"enabled"`
}

type UpdateResourceAccessPolicyRequest struct {
	Description *string `json:"description"`
	Effect      *string `json:"effect,omitempty"`
	Action      *string `json:"action,omitempty"`
	Resource    *string `json:"resource,omitempty"`
	Selector    *string `json:"selector,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

// ListPermissions returns all permissions with filtering
func (h *RBACHandler) ListPermissions(c *gin.Context) {
	var req ListPermissionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	query := h.db.Model(&model.Permission{})

	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Resource != "" {
		query = query.Where("resource = ?", req.Resource)
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.Scope != "" {
		query = query.Where("scope = ?", req.Scope)
	}

	var total int64
	query.Count(&total)

	var permissions []model.Permission
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("category, resource, action").Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     permissions,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	})
}

// CreatePermission creates a new permission
func (h *RBACHandler) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if permission name already exists
	var existing model.Permission
	if err := h.db.Where("name = ?", req.Name).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Permission with this name already exists"})
		return
	}

	permission := model.Permission{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Category:    req.Category,
		Resource:    req.Resource,
		Action:      req.Action,
		Scope:       req.Scope,
	}

	if err := h.db.Create(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, permission)
}

// GetPermission returns a single permission by ID
func (h *RBACHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")
	permissionID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var permission model.Permission
	if err := h.db.First(&permission, "id = ?", permissionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// UpdatePermission updates a permission
func (h *RBACHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	permissionID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var permission model.Permission
	if err := h.db.First(&permission, "id = ?", permissionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	updates := map[string]interface{}{}
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if err := h.db.Model(&permission).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// DeletePermission deletes a permission
func (h *RBACHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	permissionID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	// Check if permission is in use
	var rolePermCount int64
	h.db.Model(&model.RolePermission{}).Where("permission_id = ?", permissionID).Count(&rolePermCount)
	if rolePermCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Permission is in use by roles", "inUse": rolePermCount})
		return
	}

	if err := h.db.Delete(&model.Permission{}, "id = ?", permissionID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListRoles returns all roles
func (h *RBACHandler) ListRoles(c *gin.Context) {
	var req ListRolesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	query := h.db.Model(&model.Role{})

	if req.IsSystem != nil {
		query = query.Where("is_system = ?", *req.IsSystem)
	}

	var total int64
	query.Count(&total)

	var roles []model.Role
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("is_system DESC, name").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load permissions for each role
	for i := range roles {
		h.db.Where("role_id = ?", roles[i].ID).Find(&roles[i].Permissions)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     roles,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	})
}

// CreateRole creates a new role
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if role name already exists
	var existing model.Role
	if err := h.db.Where("name = ?", req.Name).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Role with this name already exists"})
		return
	}

	// Validate parent role exists if specified
	if req.ParentID != nil {
		var parent model.Role
		if err := h.db.First(&parent, "id = ?", *req.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent role not found"})
			return
		}
	}

	role := model.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsDefault:   req.IsDefault,
		ParentID:    req.ParentID,
	}

	if err := h.db.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// GetRole returns a single role with its permissions
func (h *RBACHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var role model.Role
	if err := h.db.Preload("Permissions").Preload("Parent").First(&role, "id = ?", roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Count users with this role
	var userCount int64
	h.db.Model(&model.UserRole{}).Where("role_id = ?", roleID).Count(&userCount)

	c.JSON(http.StatusOK, gin.H{
		"role":      role,
		"userCount": userCount,
	})
}

// UpdateRole updates a role
func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role model.Role
	if err := h.db.First(&role, "id = ?", roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Cannot modify system roles
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify system roles"})
		return
	}

	updates := map[string]interface{}{}
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}

	if err := h.db.Model(&role).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

// DeleteRole deletes a role
func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var role model.Role
	if err := h.db.First(&role, "id = ?", roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Cannot delete system roles
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete system roles"})
		return
	}

	// Check if role is in use
	var userCount int64
	h.db.Model(&model.UserRole{}).Where("role_id = ?", roleID).Count(&userCount)
	if userCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Role is in use by users", "userCount": userCount})
		return
	}

	// Delete role permissions and role
	h.db.Where("role_id = ?", roleID).Delete(&model.RolePermission{})
	if err := h.db.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AssignRolePermissions assigns permissions to a role
func (h *RBACHandler) AssignRolePermissions(c *gin.Context) {
	id := c.Param("id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role model.Role
	if err := h.db.First(&role, "id = ?", roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Cannot modify system roles
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify system role permissions"})
		return
	}

	// Verify all permissions exist
	var permissions []model.Permission
	if err := h.db.Where("id IN ?", req.PermissionIDs).Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(permissions) != len(req.PermissionIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Some permissions not found"})
		return
	}

	// If override is true, remove existing permissions
	if req.Override {
		h.db.Where("role_id = ?", roleID).Delete(&model.RolePermission{})
	}

	// Assign new permissions
	for _, permID := range req.PermissionIDs {
		// Check if already assigned
		var existing model.RolePermission
		if err := h.db.Where("role_id = ? AND permission_id = ?", roleID, permID).First(&existing).Error; err == gorm.ErrRecordNotFound {
			rolePerm := model.RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
			}
			h.db.Create(&rolePerm)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions assigned successfully", "count": len(req.PermissionIDs)})
}

// RemoveRolePermission removes a permission from a role
func (h *RBACHandler) RemoveRolePermission(c *gin.Context) {
	roleID := c.Param("id")
	permissionID := c.Param("permissionId")

	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	permUUID, err := uuid.Parse(permissionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var role model.Role
	if err := h.db.First(&role, "id = ?", roleUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Cannot modify system roles
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify system role permissions"})
		return
	}

	if err := h.db.Where("role_id = ? AND permission_id = ?", roleUUID, permUUID).Delete(&model.RolePermission{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission removed successfully"})
}

// ListUserRoles returns all roles assigned to a user
func (h *RBACHandler) ListUserRoles(c *gin.Context) {
	userID := c.Param("userId")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var userRoles []model.UserRole
	if err := h.db.Preload("Role").Preload("Role.Permissions").Where("user_id = ? AND (expires_at IS NULL OR expires_at > ?)", userUUID, time.Now()).Find(&userRoles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userRoles})
}

// AssignUserRole assigns a role to a user
func (h *RBACHandler) AssignUserRole(c *gin.Context) {
	userID := c.Param("userId")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req AssignUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify role exists
	var role model.Role
	if err := h.db.First(&role, "id = ?", req.RoleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Check for duplicate assignment
	var existing model.UserRole
	err = h.db.Where("user_id = ? AND role_id = ? AND resource_id = ? AND resource_type = ?",
		userUUID, req.RoleID, req.ResourceID, req.ResourceType).First(&existing).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Role already assigned to user"})
		return
	}

	userRole := model.UserRole{
		UserID:       userUUID,
		RoleID:       req.RoleID,
		ResourceID:   req.ResourceID,
		ResourceType: req.ResourceType,
		ExpiresAt:    req.ExpiresAt,
	}

	if err := h.db.Create(&userRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, userRole)
}

// RemoveUserRole removes a role from a user
func (h *RBACHandler) RemoveUserRole(c *gin.Context) {
	userID := c.Param("userId")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req RemoveUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Where("user_id = ? AND role_id = ?", userUUID, req.RoleID).Delete(&model.UserRole{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed successfully"})
}

// GetUserPermissions returns all effective permissions for a user
func (h *RBACHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("userId")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	permissions, err := model.GetEffectivePermissions(h.db, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Group by category
	grouped := make(map[string][]model.Permission)
	for _, perm := range permissions {
		grouped[perm.Category] = append(grouped[perm.Category], perm)
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
		"grouped":     grouped,
		"total":       len(permissions),
	})
}

// CheckPermission checks if a user has a specific permission
func (h *RBACHandler) CheckPermission(c *gin.Context) {
	userID := c.Param("userId")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := model.UserHasPermission(h.db, userUUID, req.Resource, req.Action, req.ResourceID, req.ResourceType)

	c.JSON(http.StatusOK, gin.H{
		"allowed": result.Allowed,
		"reason":  result.Reason,
		"source":  result.Source,
	})
}

// ListResourceAccessPolicies returns all resource access policies
func (h *RBACHandler) ListResourceAccessPolicies(c *gin.Context) {
	var policies []model.ResourceAccessPolicy

	if err := h.db.Preload("Role").Order("created_at DESC").Find(&policies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": policies})
}

// CreateResourceAccessPolicy creates a new resource access policy
func (h *RBACHandler) CreateResourceAccessPolicy(c *gin.Context) {
	var req CreateResourceAccessPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policy := model.ResourceAccessPolicy{
		Name:        req.Name,
		Description: req.Description,
		Effect:      req.Effect,
		Action:      req.Action,
		Resource:    req.Resource,
		Selector:    req.Selector,
		Enabled:     req.Enabled,
	}

	if err := h.db.Create(&policy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, policy)
}

// UpdateResourceAccessPolicy updates a resource access policy
func (h *RBACHandler) UpdateResourceAccessPolicy(c *gin.Context) {
	id := c.Param("id")
	policyID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	var req UpdateResourceAccessPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var policy model.ResourceAccessPolicy
	if err := h.db.First(&policy, "id = ?", policyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}

	updates := map[string]interface{}{}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Effect != nil {
		updates["effect"] = *req.Effect
	}
	if req.Action != nil {
		updates["action"] = *req.Action
	}
	if req.Resource != nil {
		updates["resource"] = *req.Resource
	}
	if req.Selector != nil {
		updates["selector"] = *req.Selector
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := h.db.Model(&policy).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// DeleteResourceAccessPolicy deletes a resource access policy
func (h *RBACHandler) DeleteResourceAccessPolicy(c *gin.Context) {
	id := c.Param("id")
	policyID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	if err := h.db.Delete(&model.ResourceAccessPolicy{}, "id = ?", policyID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListRolePermissions returns all permissions for a role
func (h *RBACHandler) ListRolePermissions(c *gin.Context) {
	id := c.Param("id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var rolePermissions []model.RolePermission
	if err := h.db.Preload("Permission").Where("role_id = ?", roleID).Find(&rolePermissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	permissions := make([]model.Permission, len(rolePermissions))
	for i, rp := range rolePermissions {
		permissions[i] = rp.Permission
	}

	c.JSON(http.StatusOK, gin.H{"data": permissions, "total": len(permissions)})
}

// SeedDefaultRoles seeds default roles and permissions
func (h *RBACHandler) SeedDefaultRoles(c *gin.Context) {
	// Seed permissions
	if err := model.SeedDefaultPermissions(h.db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Seed roles
	if err := model.SeedDefaultRoles(h.db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default roles and permissions seeded successfully"})
}

// GetCurrentUser returns current user information with roles and permissions
func (h *RBACHandler) GetCurrentUser(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID := userIDVal.(uuid.UUID)

	// Get user with roles
	var user model.User
	if err := h.db.Preload("Roles.Role").First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get effective permissions
	permissions, err := model.GetEffectivePermissions(h.db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse permissions for easier frontend use
	permissionsMap := make(map[string][]string) // resource -> actions
	for _, perm := range permissions {
		if _, ok := permissionsMap[perm.Resource]; !ok {
			permissionsMap[perm.Resource] = []string{}
		}
		permissionsMap[perm.Resource] = append(permissionsMap[perm.Resource], perm.Action)
	}

	c.JSON(http.StatusOK, gin.H{
		"user":        user,
		"roles":       user.Roles,
		"permissions": permissions,
		"permissionsMap": permissionsMap,
	})
}

// BatchCheckPermissions checks multiple permissions at once
func (h *RBACHandler) BatchCheckPermissions(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userID := userIDVal.(uuid.UUID)

	var req struct {
		Checks []CheckPermissionRequest `json:"checks" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results := make([]gin.H, len(req.Checks))
	for i, check := range req.Checks {
		result := model.UserHasPermission(h.db, userID, check.Resource, check.Action, check.ResourceID, check.ResourceType)
		results[i] = gin.H{
			"resource": check.Resource,
			"action":   check.Action,
			"allowed":  result.Allowed,
			"reason":   result.Reason,
		}
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// AuditLog represents a permission check audit log entry
type AuditLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	Resource  string    `gorm:"size:100;not null;index" json:"resource"`
	Action    string    `gorm:"size:50;not null" json:"action"`
	Allowed   bool      `gorm:"not null;index" json:"allowed"`
	Reason    string    `gorm:"type:text" json:"reason"`
	IP        string    `gorm:"size:50" json:"ip"`
	UserAgent string    `gorm:"size:500" json:"userAgent"`
}

// LogPermissionCheck logs a permission check for audit purposes
func (h *RBACHandler) LogPermissionCheck(userID uuid.UUID, resource, action string, allowed bool, reason string, c *gin.Context) {
	auditLog := AuditLog{
		UserID:  userID,
		Resource: resource,
		Action:   action,
		Allowed:  allowed,
		Reason:   reason,
		IP:       c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}
	h.db.Create(&auditLog)
}

// GetAuditLogs returns permission check audit logs
func (h *RBACHandler) GetAuditLogs(c *gin.Context) {
	userID := c.Param("userId")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "50")
	resource := c.Query("resource")
	action := c.Query("action")

	var logs []AuditLog
	query := h.db.Model(&AuditLog{}).Where("user_id = ?", userUUID)

	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	var total int64
	query.Count(&total)

	offset := (parseInt(page)-1) * parseInt(pageSize)
	if err := query.Order("created_at DESC").Offset(offset).Limit(parseInt(pageSize)).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     logs,
		"total":    total,
		"page":     parseInt(page),
		"pageSize": parseInt(pageSize),
	})
}

func parseInt(s string) int {
	var i int
	if _, err := json.Unmarshal([]byte(`"`+s+`"`), &i); err != nil {
		return 1
	}
	return i
}
