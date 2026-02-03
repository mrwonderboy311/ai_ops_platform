// User API client
import axios from 'axios'
import type {
  User,
  UsersResponse,
  Role,
  RolesResponse,
  PermissionsResponse,
  UserResponse,
  CheckPermissionResponse,
} from '../types/user'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const userApi = {
  // Get users
  getUsers: async (params?: {
    page?: number
    pageSize?: number
    search?: string
  }): Promise<UsersResponse> => {
    const response = await axios.get<UsersResponse>(`${API_BASE_URL}/api/v1/users`, { params })
    return response.data
  },

  // Get user by ID
  getUserByID: async (userId: string): Promise<UserResponse> => {
    const response = await axios.get<UserResponse>(`${API_BASE_URL}/api/v1/users/${userId}`)
    return response.data
  },

  // Create user
  createUser: async (user: Partial<User>): Promise<UserResponse> => {
    const response = await axios.post<UserResponse>(`${API_BASE_URL}/api/v1/users`, user)
    return response.data
  },

  // Update user
  updateUser: async (userId: string, user: Partial<User>): Promise<UserResponse> => {
    const response = await axios.put<UserResponse>(`${API_BASE_URL}/api/v1/users/${userId}`, user)
    return response.data
  },

  // Delete user
  deleteUser: async (userId: string): Promise<{ data: { success: boolean } }> => {
    const response = await axios.delete(`${API_BASE_URL}/api/v1/users/${userId}`)
    return response.data
  },

  // Get user roles
  getUserRoles: async (userId: string): Promise<{ data: Role[] }> => {
    const response = await axios.get<{ data: Role[] }>(`${API_BASE_URL}/api/v1/users/${userId}/roles`)
    return response.data
  },

  // Assign role to user
  assignRoleToUser: async (userId: string, roleId: string): Promise<{ data: { success: boolean } }> => {
    const response = await axios.post(`${API_BASE_URL}/api/v1/users/${userId}/roles`, { roleId })
    return response.data
  },

  // Remove role from user
  removeRoleFromUser: async (userId: string, roleId: string): Promise<{ data: { success: boolean } }> => {
    const response = await axios.delete(`${API_BASE_URL}/api/v1/users/${userId}/roles`, {
      data: { roleId },
    })
    return response.data
  },

  // Get roles
  getRoles: async (): Promise<RolesResponse> => {
    const response = await axios.get<RolesResponse>(`${API_BASE_URL}/api/v1/roles`)
    return response.data
  },

  // Get role by ID
  getRoleByID: async (roleId: string): Promise<{ data: Role }> => {
    const response = await axios.get<{ data: Role }>(`${API_BASE_URL}/api/v1/roles/${roleId}`)
    return response.data
  },

  // Create role
  createRole: async (role: Partial<Role>): Promise<{ data: Role }> => {
    const response = await axios.post<{ data: Role }>(`${API_BASE_URL}/api/v1/roles`, role)
    return response.data
  },

  // Update role
  updateRole: async (roleId: string, role: Partial<Role>): Promise<{ data: Role }> => {
    const response = await axios.put<{ data: Role }>(`${API_BASE_URL}/api/v1/roles/${roleId}`, role)
    return response.data
  },

  // Delete role
  deleteRole: async (roleId: string): Promise<{ data: { success: boolean } }> => {
    const response = await axios.delete(`${API_BASE_URL}/api/v1/roles/${roleId}`)
    return response.data
  },

  // Get permissions
  getPermissions: async (): Promise<PermissionsResponse> => {
    const response = await axios.get<PermissionsResponse>(`${API_BASE_URL}/api/v1/roles/permissions`)
    return response.data
  },

  // Assign permission to role
  assignPermissionToRole: async (roleId: string, permissionId: string): Promise<{ data: { success: boolean } }> => {
    const response = await axios.post(`${API_BASE_URL}/api/v1/roles/${roleId}/permissions`, { permissionId })
    return response.data
  },

  // Remove permission from role
  removePermissionFromRole: async (roleId: string, permissionId: string): Promise<{ data: { success: boolean } }> => {
    const response = await axios.delete(`${API_BASE_URL}/api/v1/roles/${roleId}/permissions`, {
      data: { permissionId },
    })
    return response.data
  },

  // Check permission
  checkPermission: async (resource: string, action: string): Promise<CheckPermissionResponse> => {
    const response = await axios.get<CheckPermissionResponse>(`${API_BASE_URL}/api/v1/users/check-permission`, {
      params: { resource, action },
    })
    return response.data
  },
}
