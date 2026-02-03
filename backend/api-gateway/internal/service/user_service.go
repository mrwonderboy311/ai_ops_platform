// Package service provides user and role management services
package service

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserService handles user management operations
type UserService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB, logger *zap.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, user *model.User) error {
	// Check if username or email already exists
	var existingUser model.User
	err := s.db.Where("username = ? OR email = ?", user.Username, user.Email).First(&existingUser).Error
	if err == nil {
		return errors.New("user with this username or email already exists")
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true

	return s.db.Create(user).Error
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(userID uuid.UUID) (*model.User, error) {
	var user model.User
	err := s.db.Preload("Roles").Preload("Roles.Permissions").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := s.db.Preload("Roles").Preload("Roles.Permissions").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsers retrieves all users with pagination
func (s *UserService) ListUsers(page, pageSize int, search string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := s.db.Model(&model.User{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.Preload("Roles").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(userID uuid.UUID, updates map[string]interface{}) error {
	// Prevent updating sensitive fields directly
	delete(updates, "password_hash")
	delete(updates, "id")

	updates["updated_at"] = time.Now()

	result := s.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteUser deletes a user (soft delete by setting is_active to false)
func (s *UserService) DeleteUser(userID uuid.UUID) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Update("is_active", false).Error
}

// UpdatePassword updates a user's password
func (s *UserService) UpdatePassword(userID uuid.UUID, newPasswordHash string) error {
	return s.db.Model(&model.User{}).
		Where("id = ?", userID).
		Update("password_hash", newPasswordHash).Error
}

// RecordLogin records a user login
func (s *UserService) RecordLogin(userID uuid.UUID, ipAddress string) error {
	now := time.Now()
	return s.db.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": &now,
			"updated_at":     time.Now(),
		}).Error
}

// CreateRole creates a new role
func (s *UserService) CreateRole(ctx context.Context, role *model.Role) error {
	// Check if role already exists
	var existingRole model.Role
	err := s.db.Where("name = ?", role.Name).First(&existingRole).Error
	if err == nil {
		return errors.New("role with this name already exists")
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	role.ID = uuid.New()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	return s.db.Create(role).Error
}

// GetRoleByID retrieves a role by ID
func (s *UserService) GetRoleByID(roleID uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := s.db.Preload("Permissions").Where("id = ?", roleID).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// ListRoles retrieves all roles
func (s *UserService) ListRoles() ([]model.Role, error) {
	var roles []model.Role
	err := s.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

// UpdateRole updates a role
func (s *UserService) UpdateRole(roleID uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	result := s.db.Model(&model.Role{}).Where("id = ?", roleID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteRole deletes a role
func (s *UserService) DeleteRole(roleID uuid.UUID) error {
	// Check if it's a system role
	var role model.Role
	err := s.db.Where("id = ?", roleID).First(&role).Error
	if err != nil {
		return err
	}
	if role.IsSystem {
		return errors.New("cannot delete system role")
	}

	// Remove all user associations
	s.db.Where("role_id = ?", roleID).Delete(&model.UserRole{})

	// Remove all permission associations
	s.db.Where("role_id = ?", roleID).Delete(&model.RolePermission{})

	return s.db.Delete(&role).Error
}

// AssignRoleToRole assigns a permission to a role
func (s *UserService) AssignPermissionToRole(roleID, permissionID uuid.UUID, createdBy uuid.UUID) error {
	// Check if association already exists
	var existing model.RolePermission
	err := s.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).First(&existing).Error
	if err == nil {
		return nil // Already assigned
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	association := &model.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		CreatedAt:    time.Now(),
		CreatedBy:    createdBy,
	}
	return s.db.Create(association).Error
}

// RemovePermissionFromRole removes a permission from a role
func (s *UserService) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	return s.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).Delete(&model.RolePermission{}).Error
}

// AssignRoleToUser assigns a role to a user
func (s *UserService) AssignRoleToUser(userID, roleID uuid.UUID, createdBy uuid.UUID) error {
	// Check if association already exists
	var existing model.UserRole
	err := s.db.Where("user_id = ? AND role_id = ?", userID, roleID).First(&existing).Error
	if err == nil {
		return nil // Already assigned
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	association := &model.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: time.Now(),
		CreatedBy: createdBy,
	}
	return s.db.Create(association).Error
}

// RemoveRoleFromUser removes a role from a user
func (s *UserService) RemoveRoleFromUser(userID, roleID uuid.UUID) error {
	return s.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&model.UserRole{}).Error
}

// GetUserPermissions retrieves all permissions for a user (through their roles)
func (s *UserService) GetUserPermissions(userID uuid.UUID) ([]model.Permission, error) {
	var permissions []model.Permission

	err := s.db.Joins(`
		JOIN role_permissions ON role_permissions.permission_id = permissions.id
		JOIN user_roles ON user_roles.role_id = role_permissions.role_id
	`).Where("user_roles.user_id = ?", userID).
		Group("permissions.id").
		Find(&permissions).Error

	return permissions, err
}

// CheckPermission checks if a user has a specific permission
func (s *UserService) CheckPermission(userID uuid.UUID, resource, action string) (bool, error) {
	var count int64
	err := s.db.Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND permissions.resource = ? AND permissions.action = ?", userID, resource, action).
		Count(&count).Error

	return count > 0, err
}

// ListPermissions retrieves all permissions
func (s *UserService) ListPermissions() ([]model.Permission, error) {
	var permissions []model.Permission
	err := s.db.Find(&permissions).Error
	return permissions, err
}

// CreatePermission creates a new permission
func (s *UserService) CreatePermission(ctx context.Context, permission *model.Permission) error {
	permission.ID = uuid.New()
	permission.CreatedAt = time.Now()
	permission.UpdatedAt = time.Now()

	return s.db.Create(permission).Error
}

// CreateSession creates a new user session
func (s *UserService) CreateSession(ctx context.Context, userID uuid.UUID, token, ipAddress, userAgent string, expiresAt time.Time) (*model.UserSession, error) {
	// Generate token hash
	tokenHash := hashToken(token)

	session := &model.UserSession{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	err := s.db.Create(session).Error
	if err != nil {
		return nil, err
	}

	return session, nil
}

// ValidateSession validates a session token
func (s *UserService) ValidateSession(token string) (*model.UserSession, error) {
	tokenHash := hashToken(token)

	var session model.UserSession
	err := s.db.Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		Preload("User").
		First(&session).Error

	if err != nil {
		return nil, err
	}

	// Update last seen
	session.LastSeenAt = time.Now()
	s.db.Save(&session)

	return &session, nil
}

// RevokeSession revokes a session
func (s *UserService) RevokeSession(sessionID uuid.UUID) error {
	return s.db.Delete(&model.UserSession{}, "id = ?", sessionID).Error
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *UserService) RevokeAllUserSessions(userID uuid.UUID) error {
	return s.db.Delete(&model.UserSession{}, "user_id = ?", userID).Error
}

// CleanupExpiredSessions removes expired sessions
func (s *UserService) CleanupExpiredSessions() error {
	return s.db.Where("expires_at < ?", time.Now()).Delete(&model.UserSession{}).Error
}

// hashToken creates a hash of the session token
func hashToken(token string) string {
	bytes := []byte(token)
	hash := make([]byte, len(bytes))
	for i, b := range bytes {
		hash[i] = b ^ 0x55 // Simple XOR obfuscation
	}
	return hex.EncodeToString(hash)
}

// InitializeSystemRoles initializes default system roles and permissions
func (s *UserService) InitializeSystemRoles(ctx context.Context) error {
	// Check if already initialized
	var count int64
	s.db.Model(&model.Role{}).Where("is_system = ?", true).Count(&count)
	if count > 0 {
		return nil // Already initialized
	}

	// Create permissions
	permissions := []model.Permission{
		{Resource: model.ResourceHosts, Action: model.ActionRead, Description: "View hosts"},
		{Resource: model.ResourceHosts, Action: model.ActionCreate, Description: "Create hosts"},
		{Resource: model.ResourceHosts, Action: model.ActionUpdate, Description: "Update hosts"},
		{Resource: model.ResourceHosts, Action: model.ActionDelete, Description: "Delete hosts"},
		{Resource: model.ResourceClusters, Action: model.ActionRead, Description: "View clusters"},
		{Resource: model.ResourceClusters, Action: model.ActionCreate, Description: "Create clusters"},
		{Resource: model.ResourceClusters, Action: model.ActionUpdate, Description: "Update clusters"},
		{Resource: model.ResourceClusters, Action: model.ActionDelete, Description: "Delete clusters"},
		{Resource: model.ResourceUsers, Action: model.ActionRead, Description: "View users"},
		{Resource: model.ResourceUsers, Action: model.ActionCreate, Description: "Create users"},
		{Resource: model.ResourceUsers, Action: model.ActionUpdate, Description: "Update users"},
		{Resource: model.ResourceUsers, Action: model.ActionDelete, Description: "Delete users"},
		{Resource: model.ResourceRoles, Action: model.ActionRead, Description: "View roles"},
		{Resource: model.ResourceRoles, Action: model.ActionCreate, Description: "Create roles"},
		{Resource: model.ResourceRoles, Action: model.ActionUpdate, Description: "Update roles"},
		{Resource: model.ResourceRoles, Action: model.ActionDelete, Description: "Delete roles"},
		{Resource: model.ResourceBatchTasks, Action: model.ActionRead, Description: "View batch tasks"},
		{Resource: model.ResourceBatchTasks, Action: model.ActionCreate, Description: "Create batch tasks"},
		{Resource: model.ResourceBatchTasks, Action: model.ActionUpdate, Description: "Update batch tasks"},
		{Resource: model.ResourceBatchTasks, Action: model.ActionDelete, Description: "Delete batch tasks"},
		{Resource: model.ResourceBatchTasks, Action: model.ActionExecute, Description: "Execute batch tasks"},
		{Resource: model.ResourceAlerts, Action: model.ActionRead, Description: "View alerts"},
		{Resource: model.ResourceAlerts, Action: model.ActionCreate, Description: "Create alerts"},
		{Resource: model.ResourceAlerts, Action: model.ActionUpdate, Description: "Update alerts"},
		{Resource: model.ResourceAlerts, Action: model.ActionDelete, Description: "Delete alerts"},
		{Resource: model.ResourceAuditLogs, Action: model.ActionRead, Description: "View audit logs"},
		{Resource: model.ResourcePerformance, Action: model.ActionRead, Description: "View performance metrics"},
		{Resource: model.ResourceSettings, Action: model.ActionRead, Description: "View settings"},
		{Resource: model.ResourceSettings, Action: model.ActionUpdate, Description: "Update settings"},
	}

	// Create roles
	adminRole := &model.Role{
		Name:        model.RoleAdmin,
		DisplayName:  "Administrator",
		Description: "Full system access",
		IsSystem:    true,
	}

	operatorRole := &model.Role{
		Name:        model.RoleOperator,
		DisplayName:  "Operator",
		Description: "Can manage infrastructure and execute tasks",
		IsSystem:    true,
	}

	viewerRole := &model.Role{
		Name:        model.RoleViewer,
		DisplayName:  "Viewer",
		Description: "Read-only access",
		IsSystem:    true,
	}

	auditorRole := &model.Role{
		Name:        model.RoleAuditor,
		DisplayName:  "Auditor",
		Description: "Can view audit logs and system activity",
		IsSystem:    true,
	}

	// Create permissions and assign to roles
	for i := range permissions {
		perm := &permissions[i]
		if err := s.CreatePermission(ctx, perm); err != nil {
			s.logger.Error("Failed to create permission", zap.Error(err))
			continue
		}

		// Admin gets all permissions
		if err := s.AssignPermissionToRole(adminRole.ID, perm.ID, uuid.Nil); err != nil {
			s.logger.Error("Failed to assign permission to admin role", zap.Error(err))
		}

		// Operator gets most permissions except user/role management
		if perm.Resource != model.ResourceUsers && perm.Resource != model.ResourceRoles {
			s.AssignPermissionToRole(operatorRole.ID, perm.ID, uuid.Nil)
		}

		// Viewer gets read-only permissions
		if perm.Action == model.ActionRead {
			s.AssignPermissionToRole(viewerRole.ID, perm.ID, uuid.Nil)
			s.AssignPermissionToRole(auditorRole.ID, perm.ID, uuid.Nil)
		}
	}

	// Save roles
	roles := []*model.Role{adminRole, operatorRole, viewerRole, auditorRole}
	for _, role := range roles {
		if err := s.CreateRole(ctx, role); err != nil {
			s.logger.Error("Failed to create role", zap.String("role", role.Name), zap.Error(err))
		}
	}

	return nil
}
