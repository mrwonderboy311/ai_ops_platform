import { apiClient } from './client'

const api = apiClient

// Types
export interface Permission {
  id: string
  name: string
  displayName: string
  description?: string
  category: string
  resource: string
  action: string
  scope: string
  createdAt: string
  updatedAt: string
}

export interface Role {
  id: string
  name: string
  displayName: string
  description?: string
  isSystem: boolean
  isDefault: boolean
  parentId?: string
  permissions?: Permission[]
  createdAt: string
  updatedAt: string
}

export interface UserRole {
  id: string
  userId: string
  roleId: string
  resourceId?: string
  resourceType?: string
  expiresAt?: string
  createdAt: string
  role?: Role
}

export interface ResourceAccessPolicy {
  id: string
  name: string
  description?: string
  effect: 'allow' | 'deny'
  action: string
  resource: string
  selector?: string
  enabled: boolean
  roleId?: string
  createdAt: string
  updatedAt: string
}

export interface CurrentUser {
  user: any
  roles: UserRole[]
  permissions: Permission[]
  permissionsMap: Record<string, string[]>
}

// Permission APIs
export const getPermissions = (params?: {
  category?: string
  resource?: string
  action?: string
  scope?: string
  page?: number
  pageSize?: number
}) =>
  api.get<{ data: Permission[]; total: number; page: number; pageSize: number }>('/rbac/permissions', { params })

export const createPermission = (data: {
  name: string
  displayName: string
  description?: string
  category: string
  resource: string
  action: string
  scope: string
}) => api.post<Permission>('/rbac/permissions', data)

export const getPermission = (id: string) => api.get<Permission>(`/rbac/permissions/${id}`)

export const updatePermission = (id: string, data: {
  displayName?: string
  description?: string
}) => api.patch<Permission>(`/rbac/permissions/${id}`, data)

export const deletePermission = (id: string) => api.delete(`/rbac/permissions/${id}`)

// Role APIs
export const getRoles = (params?: {
  isSystem?: boolean
  page?: number
  pageSize?: number
}) =>
  api.get<{ data: Role[]; total: number; page: number; pageSize: number }>('/rbac/roles', { params })

export const createRole = (data: {
  name: string
  displayName: string
  description?: string
  isDefault?: boolean
  parentId?: string
}) => api.post<Role>('/rbac/roles', data)

export const getRole = (id: string) => api.get<Role & { userCount: number }>(`/rbac/roles/${id}`)

export const updateRole = (id: string, data: {
  displayName?: string
  description?: string
  isDefault?: boolean
}) => api.patch<Role>(`/rbac/roles/${id}`, data)

export const deleteRole = (id: string) => api.delete(`/rbac/roles/${id}`)

export const getRolePermissions = (roleId: string) =>
  api.get<{ data: Permission[]; total: number }>(`/rbac/roles/${roleId}/permissions`)

export const assignRolePermissions = (roleId: string, data: {
  permissionIds: string[]
  override?: boolean
}) => api.post(`/rbac/roles/${roleId}/permissions`, data)

export const removeRolePermission = (roleId: string, permissionId: string) =>
  api.delete(`/rbac/roles/${roleId}/permissions/${permissionId}`)

export const seedDefaultRoles = () => api.post('/rbac/roles/seed')

// User Role APIs
export const getUserRoles = (userId: string) =>
  api.get<{ data: UserRole[] }>(`/rbac/users/${userId}/roles`)

export const assignUserRole = (userId: string, data: {
  roleId: string
  resourceId?: string
  resourceType?: string
  expiresAt?: string
}) => api.post<UserRole>(`/rbac/users/${userId}/roles`, data)

export const removeUserRole = (userId: string, data: { roleId: string }) =>
  api.delete(`/rbac/users/${userId}/roles`, { data })

export const getUserPermissions = (userId: string) =>
  api.get<{
    permissions: Permission[]
    grouped: Record<string, Permission[]>
    total: number
  }>(`/rbac/users/${userId}/permissions`)

export const checkPermission = (userId: string, data: {
  resource: string
  action: string
  resourceId?: string
  resourceType?: string
}) =>
  api.post<{
    allowed: boolean
    reason: string
    source: string
  }>(`/rbac/users/${userId}/check-permission`, data)

export const getAuditLogs = (userId: string, params?: {
  resource?: string
  action?: string
  page?: number
  pageSize?: number
}) =>
  api.get<{
    data: any[]
    total: number
    page: number
    pageSize: number
  }>(`/rbac/users/${userId}/audit-logs`, { params })

// Current User APIs
export const getCurrentUser = () =>
  api.get<CurrentUser>('/rbac/me')

export const batchCheckPermissions = (checks: Array<{
  resource: string
  action: string
  resourceId?: string
  resourceType?: string
}>) =>
  api.post<{ results: Array<{ resource: string; action: string; allowed: boolean; reason: string }> }>('/rbac/me/check-permissions', { checks })

// Resource Access Policy APIs
export const getResourceAccessPolicies = () =>
  api.get<{ data: ResourceAccessPolicy[] }>('/rbac/policies')

export const createResourceAccessPolicy = (data: {
  name: string
  description?: string
  effect: 'allow' | 'deny'
  action: string
  resource: string
  selector?: string
  enabled?: boolean
}) => api.post<ResourceAccessPolicy>('/rbac/policies', data)

export const updateResourceAccessPolicy = (id: string, data: {
  description?: string
  effect?: 'allow' | 'deny'
  action?: string
  resource?: string
  selector?: string
  enabled?: boolean
}) => api.patch<ResourceAccessPolicy>(`/rbac/policies/${id}`, data)

export const deleteResourceAccessPolicy = (id: string) => api.delete(`/rbac/policies/${id}`)

// Permission categories
export const PERMISSION_CATEGORIES = {
  host: 'Host Management',
  k8s: 'Kubernetes',
  observability: 'Observability',
  ai: 'AI Analysis',
  system: 'System',
} as const

// Permission actions
export const PERMISSION_ACTIONS = {
  create: 'Create',
  read: 'Read',
  update: 'Update',
  delete: 'Delete',
  execute: 'Execute',
  admin: 'Admin',
} as const

// Permission scopes
export const PERMISSION_SCOPES = {
  global: 'Global',
  cluster: 'Cluster',
  namespace: 'Namespace',
  host: 'Host',
} as const

// Resources
export const RESOURCES = {
  hosts: 'Hosts',
  clusters: 'Clusters',
  workloads: 'Workloads',
  pods: 'Pods',
  services: 'Services',
  datasources: 'Data Sources',
  dashboards: 'Dashboards',
  alert_rules: 'Alert Rules',
  anomaly_rules: 'Anomaly Rules',
  users: 'Users',
  roles: 'Roles',
  permissions: 'Permissions',
  settings: 'Settings',
} as const

const rbacApi = {
  // Permissions
  getPermissions,
  createPermission,
  getPermission,
  updatePermission,
  deletePermission,

  // Roles
  getRoles,
  createRole,
  getRole,
  updateRole,
  deleteRole,
  getRolePermissions,
  assignRolePermissions,
  removeRolePermission,
  seedDefaultRoles,

  // User Roles
  getUserRoles,
  assignUserRole,
  removeUserRole,
  getUserPermissions,
  checkPermission,
  getAuditLogs,

  // Current User
  getCurrentUser,
  batchCheckPermissions,

  // Policies
  getResourceAccessPolicies,
  createResourceAccessPolicy,
  updateResourceAccessPolicy,
  deleteResourceAccessPolicy,

  // Constants
  PERMISSION_CATEGORIES,
  PERMISSION_ACTIONS,
  PERMISSION_SCOPES,
  RESOURCES,
}

export default rbacApi
