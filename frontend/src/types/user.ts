// User management types
export interface User {
  id: string
  username: string
  email: string
  userType: 'local' | 'ldap'
  displayName: string
  avatar?: string
  phone?: string
  department?: string
  position?: string
  isActive: boolean
  lastLoginAt?: string
  createdAt: string
  updatedAt: string
  roles?: Role[]
}

export type UserRole = 'admin' | 'operator' | 'viewer' | 'auditor'

export interface Role {
  id: string
  name: UserRole
  displayName: string
  description: string
  isSystem: boolean
  permissions?: Permission[]
  createdAt: string
  updatedAt: string
}

export interface Permission {
  id: string
  resource: string
  action: string
  description: string
  createdAt: string
  updatedAt: string
}

export interface UserResponse {
  data: User
}

export interface UsersResponse {
  data: {
    users: User[]
    total: number
    page: number
    pageSize: number
  }
}

export interface RolesResponse {
  data: {
    roles: Role[]
    total: number
  }
}

export interface PermissionsResponse {
  data: {
    permissions: Permission[]
    total: number
  }
}

export interface CheckPermissionResponse {
  data: {
    hasPermission: boolean
  }
}
