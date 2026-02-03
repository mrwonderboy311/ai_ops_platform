// Package model provides data models for the application
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username     string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"username"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255)" json:"-"` // LDAP users may have empty password
	UserType     string    `gorm:"type:varchar(50);not null;default:'local'" json:"user_type"`
	DisplayName  string    `gorm:"type:varchar(255)" json:"display_name"`
	Avatar       string    `gorm:"type:varchar(500)" json:"avatar"`
	Phone        string    `gorm:"type:varchar(50)" json:"phone"`
	Department   string    `gorm:"type:varchar(255)" json:"department"`
	Position     string    `gorm:"type:varchar(255)" json:"position"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:now()" json:"updated_at"`
	Roles        []Role    `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeUpdate hook updates the UpdatedAt timestamp
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// UserType constants
const (
	UserTypeLocal = "local"
	UserTypeLDAP  = "ldap"
)

// Role represents a role in the system
type Role struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	DisplayName string     `gorm:"type:varchar(255);not null" json:"display_name"`
	Description string     `gorm:"type:text" json:"description"`
	IsSystem    bool       `gorm:"default:false" json:"is_system"` // System roles cannot be deleted
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	Users       []User     `gorm:"many2many:user_roles;" json:"users,omitempty"`
	CreatedAt   time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"default:now()" json:"updated_at"`
}

// TableName specifies the table name for Role model
func (Role) TableName() string {
	return "roles"
}

// Permission represents a permission in the system
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Resource    string    `gorm:"type:varchar(100);not null" json:"resource"`    // e.g., hosts, clusters, users
	Action      string    `gorm:"type:varchar(50);not null" json:"action"`       // e.g., create, read, update, delete
	Description string    `gorm:"type:text" json:"description"`
	Roles       []Role    `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:now()" json:"updated_at"`
}

// TableName specifies the table name for Permission model
func (Permission) TableName() string {
	return "permissions"
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null" json:"role_id"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	CreatedBy uuid.UUID `gorm:"type:uuid" json:"created_by"`
}

// TableName specifies the table name for UserRole model
func (UserRole) TableName() string {
	return "user_roles"
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;not null" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null" json:"permission_id"`
	CreatedAt    time.Time `gorm:"default:now()" json:"created_at"`
	CreatedBy    uuid.UUID `gorm:"type:uuid" json:"created_by"`
}

// TableName specifies the table name for RolePermission model
func (RolePermission) TableName() string {
	return "role_permissions"
}

// UserSession represents a user session
type UserSession struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash    string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"token_hash"`
	IPAddress    string    `gorm:"type:varchar(50)" json:"ip_address"`
	UserAgent    string    `gorm:"type:text" json:"user_agent"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	LastSeenAt   time.Time `gorm:"default:now()" json:"last_seen_at"`
	CreatedAt    time.Time `gorm:"default:now()" json:"created_at"`
}

// TableName specifies the table name for UserSession model
func (UserSession) TableName() string {
	return "user_sessions"
}

// SystemRole constants
const (
	RoleAdmin      = "admin"
	RoleOperator   = "operator"
	RoleViewer     = "viewer"
	RoleAuditor    = "auditor"
)

// System permission resources
const (
	ResourceHosts       = "hosts"
	ResourceClusters    = "clusters"
	ResourceUsers       = "users"
	ResourceRoles       = "roles"
	ResourceBatchTasks  = "batch-tasks"
	ResourceAlerts      = "alerts"
	ResourceAuditLogs   = "audit-logs"
	ResourcePerformance = "performance"
	ResourceSettings    = "settings"
)

// Permission actions
const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionExecute = "execute"
)

