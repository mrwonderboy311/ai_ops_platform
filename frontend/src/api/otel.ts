import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export interface OtelCollector {
  id: string
  userId: string
  clusterId: string
  name: string
  namespace: string
  type: 'metrics' | 'logs' | 'traces' | 'all'
  status: 'deploying' | 'running' | 'stopped' | 'error' | 'pending'
  version: string
  replicas: number
  resources?: string
  metricsEndpoint?: string
  logsEndpoint?: string
  tracesEndpoint?: string
  podNames?: string
  lastHealthCheck?: string
  errorMessage?: string
  createdAt: string
  updatedAt: string
  cluster?: {
    id: string
    name: string
  }
}

export interface CreateCollectorRequest {
  clusterId: string
  name: string
  namespace?: string
  type: 'metrics' | 'logs' | 'traces' | 'all'
  config?: string
  replicas?: number
  resources?: string
  metricsEndpoint?: string
  logsEndpoint?: string
  tracesEndpoint?: string
}

export interface UpdateCollectorRequest {
  config?: string
  replicas?: number
  resources?: string
  metricsEndpoint?: string
  logsEndpoint?: string
  tracesEndpoint?: string
}

export interface CollectorDeploymentStatus {
  status: string
  podNames: string[]
  replicas: number
  readyReplicas: number
  metricsUrl?: string
  errorMessage?: string
}

const otelApi = {
  // List all collectors
  getCollectors: async (params?: { clusterId?: string; namespace?: string; type?: string; status?: string; page?: number; pageSize?: number }) => {
    const response = await axios.get<{
      data: OtelCollector[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }>(`${API_BASE_URL}/api/v1/otel/collectors`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // Get a specific collector
  getCollector: async (id: string) => {
    const response = await axios.get<OtelCollector>(`${API_BASE_URL}/api/v1/otel/collectors/${id}`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Create a collector
  createCollector: async (data: CreateCollectorRequest) => {
    const response = await axios.post<OtelCollector>(`${API_BASE_URL}/api/v1/otel/collectors`, data, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Update a collector
  updateCollector: async (id: string, data: UpdateCollectorRequest) => {
    const response = await axios.put<OtelCollector>(`${API_BASE_URL}/api/v1/otel/collectors/${id}`, data, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Delete a collector
  deleteCollector: async (id: string) => {
    const response = await axios.delete<{ message: string }>(`${API_BASE_URL}/api/v1/otel/collectors/${id}`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Get collector status
  getCollectorStatus: async (id: string) => {
    const response = await axios.get<CollectorDeploymentStatus>(`${API_BASE_URL}/api/v1/otel/collectors/${id}/status`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Start a collector
  startCollector: async (id: string) => {
    const response = await axios.post<{ message: string }>(`${API_BASE_URL}/api/v1/otel/collectors/${id}/start`, {}, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Stop a collector
  stopCollector: async (id: string) => {
    const response = await axios.post<{ message: string }>(`${API_BASE_URL}/api/v1/otel/collectors/${id}/stop`, {}, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },

  // Restart a collector
  restartCollector: async (id: string) => {
    const response = await axios.post<{ message: string }>(`${API_BASE_URL}/api/v1/otel/collectors/${id}/restart`, {}, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
    })
    return response.data
  },
}

export default otelApi
