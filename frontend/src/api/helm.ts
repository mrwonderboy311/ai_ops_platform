import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export interface HelmRepository {
  id: string
  userId: string
  name: string
  description: string
  type: 'http' | 'https' | 'oci'
  status: 'active' | 'inactive' | 'error'
  url: string
  username?: string
  insecureSkipTLS: boolean
  lastSyncedAt?: string
  lastSyncStatus?: string
  lastSyncError?: string
  chartCount: number
  createdAt: string
  updatedAt: string
}

export interface CreateHelmRepoRequest {
  name: string
  description?: string
  type: 'http' | 'https' | 'oci'
  url: string
  username?: string
  password?: string
  caFile?: string
  certFile?: string
  keyFile?: string
  insecureSkipTLS?: boolean
}

export interface UpdateHelmRepoRequest {
  name?: string
  description?: string
  url?: string
  username?: string
  password?: string
  caFile?: string
  certFile?: string
  keyFile?: string
  insecureSkipTLS?: boolean
}

export interface HelmRepoTestRequest {
  type: 'http' | 'https' | 'oci'
  url: string
  username?: string
  password?: string
  caFile?: string
  certFile?: string
  keyFile?: string
  insecureSkipTLS?: boolean
}

export interface HelmRepoTestResponse {
  success: boolean
  chartCount: number
  message: string
  error?: string
}

export interface HelmRepoListResponse {
  data: HelmRepository[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

const helmApi = {
  // List all Helm repositories
  getRepositories: async (params?: { status?: string; type?: string; page?: number; pageSize?: number }) => {
    const response = await axios.get<HelmRepoListResponse>(`${API_BASE_URL}/api/v1/helm/repositories`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // Get a specific Helm repository
  getRepository: async (id: string) => {
    const response = await axios.get<HelmRepository>(`${API_BASE_URL}/api/v1/helm/repositories/${id}`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Create a new Helm repository
  createRepository: async (data: CreateHelmRepoRequest) => {
    const response = await axios.post<HelmRepository>(`${API_BASE_URL}/api/v1/helm/repositories`, data, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Update a Helm repository
  updateRepository: async (id: string, data: UpdateHelmRepoRequest) => {
    const response = await axios.put<HelmRepository>(`${API_BASE_URL}/api/v1/helm/repositories/${id}`, data, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Delete a Helm repository
  deleteRepository: async (id: string) => {
    const response = await axios.delete<{ message: string }>(`${API_BASE_URL}/api/v1/helm/repositories/${id}`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Test a Helm repository connection
  testRepository: async (data: HelmRepoTestRequest) => {
    const response = await axios.post<HelmRepoTestResponse>(`${API_BASE_URL}/api/v1/helm/repositories/test`, data, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Sync a Helm repository
  syncRepository: async (id: string) => {
    const response = await axios.post<{ message: string }>(`${API_BASE_URL}/api/v1/helm/repositories/${id}/sync`, {}, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },
}

export default helmApi
