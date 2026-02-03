import { apiClient } from './client'
import type {
  Host,
  CreateHostRequest,
  UpdateHostRequest,
  HostListParams,
  HostListResponse,
  RejectHostRequest,
} from '../types/host'

export const hostApi = {
  // List hosts with pagination and filters
  listHosts: async (params: HostListParams = {}): Promise<HostListResponse> => {
    const queryParams = new URLSearchParams()

    if (params.page) queryParams.append('page', String(params.page))
    if (params.pageSize) queryParams.append('page_size', String(params.pageSize))
    if (params.status) queryParams.append('status', params.status)
    if (params.hostname) queryParams.append('hostname', params.hostname)
    if (params.ipAddress) queryParams.append('ip_address', params.ipAddress)
    if (params.registeredBy) queryParams.append('registered_by', params.registeredBy)
    if (params.sortBy) queryParams.append('sort_by', params.sortBy)
    if (params.sortDesc !== undefined) queryParams.append('sort_desc', String(params.sortDesc))

    const response = await apiClient.get<{ data: HostListResponse }>(`/api/v1/hosts?${queryParams.toString()}`)
    return response.data.data
  },

  // Get a single host by ID
  getHost: async (id: string): Promise<Host> => {
    const response = await apiClient.get<{ data: Host }>(`/api/v1/hosts/${id}`)
    return response.data.data
  },

  // Create a new host
  createHost: async (data: CreateHostRequest): Promise<Host> => {
    const response = await apiClient.post<{ data: Host }>('/api/v1/hosts', data)
    return response.data.data
  },

  // Update an existing host
  updateHost: async (id: string, data: UpdateHostRequest): Promise<Host> => {
    const response = await apiClient.put<{ data: Host }>(`/api/v1/hosts/${id}`, data)
    return response.data.data
  },

  // Delete a host
  deleteHost: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/v1/hosts/${id}`)
  },

  // Approve a host
  approveHost: async (id: string): Promise<Host> => {
    const response = await apiClient.patch<{ data: Host }>(`/api/v1/hosts/${id}/approve`)
    return response.data.data
  },

  // Reject a host
  rejectHost: async (id: string, data: RejectHostRequest = {}): Promise<Host> => {
    const response = await apiClient.patch<{ data: Host }>(`/api/v1/hosts/${id}/reject`, data)
    return response.data.data
  },
}
