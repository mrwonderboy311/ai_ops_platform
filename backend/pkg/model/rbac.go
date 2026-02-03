// Package model provides data models for RBAC (Role-Based Access Control)
package model

import (
	"time"

	"github.com/google.comuuid"
	"gorm.io/gorm"
)

// Permission represents a specific permission in the system
type Permission struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Permission identification
	Name        string `gorm:"size:255;not null;uniqueIndex:idx_permission_name" json:"name"`
	DisplayName string `gorm:"size:255;not null" json:"displayName"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	Category    string `gorm:"size:100;not null;index:idx_permission_category" json:"category"` // host, k8s, observability, ai, system
	Resource    string `gorm:"size:100;not null;index:idx_permission_resource" json:"resource"`    // hosts, clusters, pods, etc.
	Action      string `gorm:"size:100;not null" json:"action"`       // create, read, update, delete, execute, etc.

	// Permission details
	Scope       string `gorm:"size:50;default:PermissionScopeGlobal" json:"scope"` // global, cluster, namespace, host
	Conditions  string `gorm:"type:text" json:"conditions,omitempty"` // JSON: additional conditions

	// Relationships
	RolePermissions []RolePermission `gorm:"foreignKey:PermissionID" json:"rolePermissions,omitempty"`
}

// PermissionScope constants
const (
	PermissionScopeGlobal     = "global"     // Apply to all resources
	PermissionScopeCluster    = "cluster"    // Apply to specific cluster
	PermissionScopeNamespace = "namespace" // Apply to specific namespace
	PermissionScopeHost       = "host"       // Apply to specific host
)

// Role represents a role with associated permissions
type Role struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Role details
	Name        string `gorm:"size:100;not null;uniqueIndex:idx_role_name" json:"name"`
	DisplayName string `gorm:"size:255;not null" json:"displayName"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	IsSystem    bool   `gorm:"default:false;not null" json:"isSystem"` // System roles cannot be deleted
	IsDefault   bool   `gorm:"default:false;not null" json:"isDefault"`  // Default role for new users

	// Role hierarchy
	ParentID    *uuid.UUID `gorm:"type:uuid" json:"parentId,omitempty"` // Parent role for inheritance

	// Relationships
	Parent          *Role              `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children        []Role             `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	RolePermissions []RolePermission   `gorm:"foreignKey:RoleID" json:"rolePermissions,omitempty"`
	UserRoles      []UserRole        `gorm:"foreignKey:RoleID" json:"userRoles,omitempty"`
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;not null" json:"roleId"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null" json:"permissionId"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// Additional constraints for this role-permission pair
	Conditions   string `gorm:"type:text" json:"conditions,omitempty"` // JSON: override permission conditions
	Disabled     bool   `gorm:"default:false" json:"disabled"`

	// Relationships
	Role       *Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission *Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

// UserRole represents the assignment of a role to a user
type UserRole struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_user_role_user_id;index:idx_user_role_unique" json:"userId"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null;index:idx_user_role_role_id;index:idx_user_role_unique" json:"roleId"`
	AssignedBy *uuid.UUID `json:"assignedBy,omitempty"` // User who assigned this role

	// Resource-scoped assignment (for cluster-specific roles, etc.)
	// If set, this role only applies to the specified resource
	ResourceID   *uuid.UUID `gorm:"type:uuid" json:"resourceId,omitempty"`
	ResourceType string     `gorm:"size:50" json:"resourceType,omitempty"` // cluster, host, etc.
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"` // Temporary role assignment

	// Relationships
	User    *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role    *Role     `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// ResourceAccessPolicy represents fine-grained access control for specific resources
type ResourceAccessPolicy struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_resource_policy_user_id" json:"userId"`
	ClusterID *uuid.UUID `gorm:"type:uuid;index:idx_resource_policy_cluster_id" json:"clusterId,omitempty"`
	HostID    *uuid.UUID `gorm:"type:uuid;index:idx_resource_policy_host_id" json:"hostId,omitempty"`

	// Policy details
	Name       string    `gorm:"size:255;not null" json:"name"`
	Effect     string    `gorm:"size:20;not null;default:PolicyEffectAllow" json:"effect"` // allow, deny
	Action     string    `gorm:"size:100;not null" json:"action"` // read, write, delete, execute
	Resource   string    `gorm:"size:100;not null" json:"resource"` // pods, services, configmaps, etc.
	Selector   string    `gorm:"type:text" json:"selector,omitempty"`  // JSON: label selector for matching resources

	// Conditions
	Conditions string    `gorm:"type:text" json:"conditions,omitempty"` // JSON: additional conditions
	Reason     string    `gorm:"type:text" json:"reason,omitempty"`

	// Status
	Enabled    bool      `gorm:"default:true" json:"enabled"`

	// Relationships
	Cluster    *K8sCluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	Host      *Host       `gorm:"foreignKey:HostID" json:"host,omitempty"`
}

// PolicyEffect constants
const (
	PolicyEffectAllow = "allow"
	PolicyEffectDeny  = "deny"
)

// CreateRoleRequest represents a request to create a role
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"displayName" binding:"required"`
	Description string `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parentId,omitempty"`
	PermissionIDs []uuid.UUID `json:"permissionIds" binding:"required,min=1"`
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	DisplayName *string  `json:"displayName,omitempty"`
	Description *string  `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parentId,omitempty"`
	PermissionIDs []uuid.UUID `json:"permissionIds,omitempty"`
}

// AssignRoleRequest represents a request to assign a role to a user
type AssignRoleRequest struct {
	RoleID       uuid.UUID  `json:"roleId" binding:"required"`
	ResourceID   *uuid.UUID `json:"resourceId,omitempty"`
	ResourceType string     `json:"resourceType,omitempty"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
}

// CreatePermissionRequest represents a request to create a permission
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"displayName" binding:"required"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Scope       string `json:"scope" binding:"required,oneof=global cluster namespace host"`
	Conditions  string `json:"conditions,omitempty"`
}

// UpdatePermissionRequest represents a request to update a permission
type UpdatePermissionRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	Description *string `json:"description,omitempty"`
	Conditions  *string `json:"conditions,omitempty"`
}

// CheckPermissionRequest represents a request to check if a user has a permission
type CheckPermissionRequest struct {
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	ResourceID   *uuid.UUID `json:"resourceId,omitempty"`
	ResourceType string     `json:"resourceType,omitempty"`
}

// CheckPermissionResponse represents the response from a permission check
type CheckPermissionResponse struct {
	HasPermission bool   `json:"hasPermission"`
	Reason       string `json:"reason,omitempty"`
}

// CreateResourceAccessPolicyRequest represents a request to create a resource access policy
type CreateResourceAccessPolicyRequest struct {
	ClusterID   *uuid.UUID `json:"clusterId,omitempty"`
	HostID      *uuid.UUID `json:"hostId,omitempty"`
	Name        string     `json:"name" binding:"required"`
	Effect      string     `json:"effect" binding:"required,oneof=allow deny"`
	Action      string     `json:"action" binding:"required"`
	Resource    string     `json:"resource" binding:"required"`
	Selector    string     `json:"selector,omitempty"`
	Conditions  string     `json:"conditions,omitempty"`
	Reason      string     `json:"reason,omitempty"`
}

// UpdateResourceAccessPolicyRequest represents a request to update a resource access policy
type UpdateResourceAccessPolicyRequest struct {
	Effect     *string  `json:"effect,omitempty"`
	Action     *string  `json:"action,omitempty"`
	Selector   *string  `json:"selector,omitempty"`
	Conditions *string  `json:"conditions,omitempty"`
	Reason     *string  `json:"reason,omitempty"`
	Enabled    *bool    `json:"enabled,omitempty"`
}

// GetUserPermissionsResponse represents all permissions for a user
type GetUserPermissionsResponse struct {
	DirectPermissions []PermissionSummary `json:"directPermissions"`
	RolePermissions    []PermissionSummary `json:"rolePermissions"`
	ResourcePolicies    []ResourcePolicySummary `json:"resourcePolicies"`
}

// PermissionSummary represents a simplified permission object
type PermissionSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Category    string `json:"category"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Scope       string `json:"scope"`
}

// ResourcePolicySummary represents a simplified resource access policy
type ResourcePolicySummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Effect    string `json:"effect"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Enabled   bool   `json:"enabled"`
}

// PermissionCheckResult represents the result of checking a specific permission
type PermissionCheckResult struct {
	Allowed   bool   `json:"allowed"`
	Reason    string `json:"reason,omitempty"`
	Source    string `json:"source"` // "role", "policy", "denied"
}

// Before creating a new role, check if the permission name already exists
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	var existing Permission
	if err := tx.Where("name = ?", p.Name).First(&existing).Error); err == nil {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

// Before creating a new role, check if the role name already exists
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	var existing Role
	if err := tx.Where("name = ?", r.Name).First(&existing).Error; err == nil {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

// UserHasPermission checks if a user has a specific permission
// This is the main authorization function
func UserHasPermission(db *gorm.DB, userID uuid.UUID, resource, action string, resourceID *uuid.UUID, resourceType string) PermissionCheckResult {
	// First check if user is super admin (has a special role)
	var adminRole Role
	if err := db.Where("name = ? AND is_system = ?", "super_admin", true).First(&adminRole).Error; err == nil {
		// Check if user has super admin role
		var userRole UserRole
		if err := db.Where("user_id = ? AND role_id = ?", userID, adminRole.ID).First(&userRole).Error; err == nil {
			return PermissionCheckResult{Allowed: true, Source: "super_admin"}
		}
	}

	// Check direct resource access policies (deny takes precedence)
	var policies []ResourceAccessPolicy
	query := db.Model(&ResourceAccessPolicy{}).Where("user_id = ? AND enabled = ?", userID, true)

	if resourceID != nil {
		query = query.Where("resource_id = ?", resourceID)
	}
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	query = query.Where("resource = ? AND action = ?", resource, action)

	if err := query.Find(&policies).Error; err != nil {
		return PermissionCheckResult{Allowed: false, Reason: "policy_check_error"}
	}

	// Check for explicit deny policies first
	for _, policy := range policies {
		if policy.Effect == PolicyEffectDeny {
			return PermissionCheckResult{
				Allowed: false,
				Reason:  fmt.Sprintf("Denied by policy: %s", policy.Name),
				Source:  "policy",
			}
		}
	}

	// Check for explicit allow policies
	for _, policy := range policies {
		if policy.Effect == PolicyEffectAllow {
			return PermissionCheckResult{
				Allowed: true,
				Reason:  fmt.Sprintf("Allowed by policy: %s", policy.Name),
				Source:  "policy",
			}
		}
	}

	// Check role-based permissions
	// Get all roles assigned to the user
	var userRoles []UserRole
	if err := db.Preload("Role.RolePermissions.Permission").Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return PermissionCheckResult{Allowed: false, Reason: "role_check_error"}
	}

	for _, userRole := range userRoles {
		// Check if role assignment is expired
		if userRole.ExpiresAt != nil && userRole.ExpiresAt.Before(time.Now()) {
			continue
		}

		role := userRole.Role
		if role == nil {
			continue
		}

		// Check all permissions in the role
		for _, rolePerm := range role.RolePermissions {
			if rolePerm.Disabled {
				continue
			}

			perm := rolePerm.Permission
			if perm == nil {
				continue
			}

			// Check if permission matches the request
			if perm.Resource == resource && perm.Action == action {
				// Check scope
				switch perm.Scope {
				case PermissionScopeGlobal:
					return PermissionCheckResult{
						Allowed: true,
						Reason:  fmt.Sprintf("Allowed by role: %s", role.Name),
						Source:  "role",
					}
				case PermissionScopeCluster:
					// Check if resourceID matches cluster
					if resourceID != nil {
						var cluster Cluster
						if err := db.Where("id = ?", resourceID).First(&cluster).Error; err == nil {
							return PermissionCheckResult{
								Allowed: true,
								Reason:  fmt.Sprintf("Allowed by role: %s", role.Name),
								Source:  "role",
							}
						}
					}
				case PermissionScopeNamespace:
					// Would need to check namespace membership
					// For now, return false as namespace-scoped permissions need cluster context
				case PermissionScopeHost:
					// Check if resourceID matches host
					if resourceID != nil {
						var host Host
						if err := db.Where("id = ?", resourceID).First(&host).Error; err == nil {
							return PermissionCheckResult{
								Allowed: true,
								Reason:  fmt.Sprintf("Allowed by role: role: %s", role.Name),
								Source:  "role",
							}
						}
					}
				}
			}
		}
	}

	return PermissionCheckResult{
		Allowed: false,
		Reason:  "no_permission",
		Source:  "none",
	}
}

// GetEffectivePermissions returns all effective permissions for a user
func GetEffectivePermissions(db *gorm.DB, userID uuid.UUID) GetUserPermissionsResponse {
	var response GetUserPermissionsResponse

	// Get role-based permissions
	var userRoles []UserRole
	db.Preload("Role.RolePermissions.Permission.Permission").Where("user_id = ?", userID).Find(&userRoles)

	permissionMap := make(map[string]PermissionSummary)
	for _, userRole := range userRoles {
		// Check if role assignment is expired
		if userRole.ExpiresAt != nil && userRole.ExpiresAt.Before(time.Now()) {
			continue
		}

		role := userRole.Role
		if role == nil {
			continue
		}

		for _, rolePerm := range role.RolePermissions {
			if rolePerm.Disabled {
				continue
			}

			perm := rolePerm.Permission
			if perm == nil {
				continue
			}

			key := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
			permissionMap[key] = PermissionSummary{
				ID:          perm.ID.String(),
				Name:        perm.Name,
				DisplayName: perm.DisplayName,
				Category:    perm.Category,
				Resource:    perm.Resource,
				Action:      perm.Action,
				Scope:       perm.Scope,
			}
		}
	}

	// Convert map to slice
	for _, perm := range permissionMap {
		response.RolePermissions = append(response.RolePermissions, perm)
	}

	return response
}

// SeedDefaultPermissions seeds the database with default permissions
func SeedDefaultPermissions(db *gorm.DB) error {
	// Check if permissions already exist
	var count int64
	db.Model(&Permission{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	permissions := []Permission{
		// Host management permissions
		{Name: "hosts.list", DisplayName: "List Hosts", Category: "host", Resource: "hosts", Action: "list", Scope: PermissionScopeGlobal},
		{Name: "hosts.get", DisplayName: "View Host Details", Category: "host", Resource: "hosts", Action: "get", Scope: PermissionScopeGlobal},
		{Name: "hosts.create", DisplayName: "Add Hosts", Category: "host", Resource: "hosts", Action: "create", Scope: PermissionScopeGlobal},
	{Name: "hosts.update", DisplayName: "Update Hosts", Category: "host", Resource: "hosts", Action: "update", Scope: PermissionScopeGlobal},
		{Name: "hosts.delete", DisplayName: "Delete Hosts", Category: "host", Resource: "hosts", Action: "delete", Scope: PermissionScopeGlobal},
	{Name: "hosts.ssh", DisplayName: "SSH Access", Category: "host", Resource: "hosts", Action: "ssh", Scope: PermissionScopeGlobal},
		{Name: "hosts.files", DisplayName: "File Management", Category: "host", Resource: "hosts", Action: "files", Scope: PermissionScopeGlobal},
	{Name: "hosts.processes", DisplayName: "Process Management", Category: "host", Resource: "hosts", Action: "processes", Scope: PermissionScopeGlobal},

		// Cluster management permissions
		{Name: "clusters.list", DisplayName: "List Clusters", Category: "k8s", Resource: "clusters", Action: "list", Scope: PermissionScopeGlobal},
		{Name: "clusters.get", DisplayName: "View Cluster Details", Category: "k8s", Resource: "clusters", Action: "get", Scope: PermissionScopeGlobal},
		{Name: "clusters.create", DisplayName: "Add Clusters", Category: "k8s", Resource: "clusters", Action: "create", Scope: PermissionScopeGlobal},
		{Name: "clusters.update", DisplayName: "Update Clusters", Category: "k8s", Resource: "clusters", Action: "update", Scope: PermissionScopeGlobal},
		{Name: "clusters.delete", DisplayName: "Delete Clusters", Category: "k8s", Resource: "clusters", Action: "delete", Scope: PermissionScopeGlobal},

		// Workload permissions
		{Name: "workloads.list", DisplayName: "List Workloads", Category: "k8s", Resource: "workloads", Action: "list", Scope: PermissionScopeCluster},
		{Name: "workloads.get", DisplayName: "View Workload Details", Category: "k8s", Resource: "workloads", Action: "get", Scope: PermissionScopeCluster},
		{Name: "workloads.create", DisplayName: "Deploy Workloads", Category: "k8s", Resource: "workloads", Action: "create", Scope: PermissionScopeCluster},
		{Name: "workloads.update", DisplayName: "Update Workloads", Category: "k8s", Resource: "workloads", Action: "update", Scope: PermissionScopeCluster},
		{Name: "workloads.delete", DisplayName: "Delete Workloads", Category: "k8s", Resource: "workloads", Action: "delete", Scope: PermissionScopeCluster},

		// Pod permissions
		{Name: "pods.list", DisplayName: "List Pods", Category: "k8s", Resource: "pods", Action: "list", Scope: PermissionScopeNamespace},
		{Name: "pods.get", DisplayName: "View Pod Details", Category: "k8s", Resource: "pods", Action: "get", Scope: PermissionScopeNamespace},
		{Name: "pods.logs", DisplayName: "View Pod Logs", Category: "k8s", Resource: "pods", Action: "logs", Scope: PermissionScopeNamespace},
		{Name: "pods.terminal", DisplayName: "Pod Terminal Access", Category: "k8s", Resource: "pods", Action: "terminal", Scope: PermissionScopeNamespace},
		{Name: "pods.delete", DisplayName: "Delete Pods", Category: "k8s", Resource: "pods", Action: "delete", Scope: PermissionScopeNamespace},

		// Observability permissions
		{Name: "otel.list", DisplayName: "List OTEL Collectors", Category: "observability", Resource: "otel", Action: "list", Scope: PermissionScopeGlobal},
		{Name: "otel.manage", DisplayName: "Manage OTEL Collectors", Category: "observability", Resource: "otel", Action: "manage", Scope: PermissionScopeGlobal},
		{Name: "prometheus.list", DisplayName: "View Prometheus Data Sources", Category: "observability", Resource: "prometheus", Action: "list", Scope: PermissionScopeGlobal},
		{Name: "prometheus.manage", DisplayName: "Manage Prometheus Data Sources", Category: "observability", Resource: "prometheus", Action: "manage", Scope: PermissionScopeGlobal},
		{Name: "prometheus.query", DisplayName: "Execute Prometheus Queries", Category: "observability", Resource: "prometheus", Action: "query", Scope: PermissionScopeGlobal},
		{Name: "prometheus.alerts", DisplayName: "Manage Prometheus Alert Rules", Category: "observability", Resource: "prometheus", Action: "alerts", Scope: PermissionScopeGlobal},
		{Name: "grafana.list", DisplayName: "View Grafana Instances", Category: "observability", Resource: "grafana", Action: "list", Scope: PermissionScopeGlobal},
		{Name: "grafana.manage", DisplayName: "Manage Grafana Instances", Category: "observability", Resource: "grafana", Action: "manage", Scope: PermissionScopeGlobal},

		// AI Analysis permissions
		{Name: "ai.anomaly.list", DisplayName: "View Anomaly Detection", Category: "ai", Resource: "ai", Action: "anomaly_list", Scope: PermissionScopeGlobal},
		{Name: "ai.anomaly.manage", DisplayName: "Manage Anomaly Detection", Category: "ai", Resource: "ai", Action: "anomaly_manage", Scope: PermissionScopeGlobal},
		{Name: "ai.llm.chat", DisplayName: "AI Chat Access", Category: "ai", Resource: "llm", Action: "chat", Scope: PermissionScopeGlobal},

		// System permissions
		{Name: "users.list", DisplayName: "List Users", Category: "system", Resource: "users", Action: "list", Scope: PermissionScopeGlobal},
		{Name: "users.manage", DisplayName: "Manage Users", Category: "system", Resource: "users", Action: "manage", Scope: PermissionScopeGlobal},
		{Name: "roles.list", DisplayName: "List Roles", Category: "system", Resource: "roles", Action: "list", Scope: PermissionScopeGlobal},
		{Name: "roles.manage", DisplayName: "Manage Roles", Category: "system", Resource: "roles", Action: "manage", Scope: PermissionScopeGlobal},
		{Name: "policies.list", DisplayName: "View Access Policies", Category: "system", Resource: "policies", Action: "list", Scope: PermissionScopeGlobal},
		{Name: "policies.manage", DisplayName: "Manage Access Policies", Category: "system", Resource: "policies", Action: "manage", Scope: PermissionScopeGlobal},
		{Name: "audit.view", DisplayName: "View Audit Logs", Category: "system", Resource: "audit", Action: "view", Scope: PermissionScopeGlobal},
	}

	for _, perm := range permissions {
		if err := db.Create(&perm).Error; err != nil {
			return err
		}
	}

	return nil
}

// SeedDefaultRoles seeds the database with default roles
func SeedDefaultRoles(db *gorm.DB) error {
	// Check if roles already exist
	var count int64
	db.Model(&Role{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	// Get all permissions
	var allPermissions []Permission
	if err := db.Find(&allPermissions).Error; err != nil {
		return err
	}

	permMap := make(map[string]uuid.UUID)
	for _, perm := range allPermissions {
		permMap[perm.Name] = perm.ID
	}

	// Create default roles
	roles := []Role{
		{
			Name:        "super_admin",
			DisplayName:  "Super Administrator",
			Description: "Full access to all resources",
			IsSystem:     true,
			IsDefault:    false,
		},
		{
			Name:        "admin",
			DisplayName:  "Administrator",
			Description: "Administrative access to most resources",
			IsSystem:     true,
			IsDefault:    true,
		},
		{
			Name:        "operator",
			DisplayName:  "Operator",
			Description: "Operational access - can view and manage resources",
			IsSystem:     true,
			IsDefault:    false,
		},
		{
			Name:        "viewer",
			DisplayName:  "Viewer",
			Description: "Read-only access to resources",
			IsSystem:     true,
			IsDefault:    false,
		},
	}

	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			return err
		}

		// Assign permissions based on role type
		var permissionIDs []uuid.UUID
		switch role.Name {
		case "super_admin":
			// All permissions
			for _, perm := range allPermissions {
				permissionIDs = append(permissionIDs, perm.ID)
			}
		case "admin":
			// All permissions except user/role/policy management
			for _, perm := range allPermissions {
				if !strings.HasPrefix(perm.Name, "users.") &&
				   !strings.HasPrefix(perm.Name, "roles.") &&
				   !strings.HasPrefix(perm.Name, "policies.") {
					permissionIDs = append(permissionIDs, perm.ID)
				}
			}
		case "operator":
			// Operational permissions
			for _, perm := range allPermissions {
				if !strings.HasPrefix(perm.Name, "users.") &&
				   !strings.HasPrefix(perm.Name, "roles.") &&
				   !strings.HasPrefix(perm.Name, "policies.") {
					permissionIDs = append(permissionIDs, perm.ID)
				}
			}
		case "viewer":
			// Read-only permissions
			for _, perm := range allPermissions {
				if perm.Action == "list" || perm.Action == "get" || perm.Action == "view" || perm.Action == "logs" {
					permissionIDs = append(permissionIDs, perm.ID)
				}
			}
		}

		// Create role permissions
		for _, permID := range permissionIDs {
			rolePerm := RolePermission{
				RoleID:       role.ID,
				PermissionID: permID,
			}
			if err := db.Create(&rolePerm).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
