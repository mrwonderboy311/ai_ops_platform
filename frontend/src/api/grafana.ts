import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export interface GrafanaInstance {
  id: string
  userId: string
  clusterId?: string
  name: string
  url: string
  username?: string
  status: 'active' | 'inactive' | 'error'
  serviceAccountId?: string
  autoSync: boolean
  syncInterval: number
  lastSyncAt?: string
  syncStatus: string
  syncError?: string
  dashboardCount: number
  dataSourceCount: number
  createdAt: string
  updatedAt: string
  cluster?: {
    id: string
    name: string
  }
}

export interface CreateGrafanaInstanceRequest {
  clusterId?: string
  name: string
  url: string
  apiKey?: string
  username?: string
  password?: string
  serviceAccountId?: string
  serviceAccountToken?: string
  autoSync?: boolean
  syncInterval?: number
}

export interface UpdateGrafanaInstanceRequest {
  name?: string
  url?: string
  apiKey?: string
  username?: string
  password?: string
  serviceAccountId?: string
  serviceAccountToken?: string
  status?: 'active' | 'inactive' | 'error'
  autoSync?: boolean
  syncInterval?: number
}

export interface TestGrafanaInstanceRequest {
  url: string
  apiKey?: string
  username?: string
  password?: string
  serviceAccountId?: string
  serviceAccountToken?: string
}

export interface TestGrafanaInstanceResponse {
  success: boolean
  version?: string
  message: string
  error?: string
  duration: number
}

export interface SyncGrafanaInstanceRequest {
  syncDashboards?: boolean
  syncDataSources?: boolean
  syncFolders?: boolean
}

export interface SyncGrafanaInstanceResponse {
  success: boolean
  message: string
  dashboardsAdded?: number
  dashboardsUpdated?: number
  dashboardsRemoved?: number
  dataSourcesAdded?: number
  foldersAdded?: number
  duration: number
}

export interface GrafanaDashboard {
  id: string
  userId: string
  instanceId: string
  clusterId?: string
  grafanaUid: string
  grafanaId: number
  title: string
  slug?: string
  tags?: string
  folderTitle?: string
  folderUid?: string
  folderId?: number
  config: string
  version: number
  isStarred: boolean
  synced: boolean
  syncedAt?: string
  createdAt: string
  updatedAt: string
  instance?: GrafanaInstance
  cluster?: {
    id: string
    name: string
  }
}

export interface GrafanaDataSource {
  id: string
  userId: string
  instanceId: string
  grafanaUid: string
  grafanaId: number
  name: string
  type: string
  isDefault: boolean
  url?: string
  database?: string
  jsonData?: string
  healthStatus?: string
  createdAt: string
  updatedAt: string
  instance?: GrafanaInstance
}

export interface GrafanaFolder {
  id: string
  userId: string
  instanceId: string
  grafanaUid: string
  grafanaId: number
  title: string
  synced: boolean
  syncedAt?: string
  createdAt: string
  updatedAt: string
  instance?: GrafanaInstance
}

const grafanaApi = {
  // ============== Instance Management ==============

  // List all Grafana instances
  getInstances: async (params?: {
    clusterId?: string
    status?: string
    page?: number
    pageSize?: number
  }) => {
    const response = await axios.get<{
      data: GrafanaInstance[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }>(`${API_BASE_URL}/api/v1/grafana/instances`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // Get a specific instance
  getInstance: async (id: string) => {
    const response = await axios.get<GrafanaInstance>(
      `${API_BASE_URL}/api/v1/grafana/instances/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Create a new instance
  createInstance: async (data: CreateGrafanaInstanceRequest) => {
    const response = await axios.post<GrafanaInstance>(
      `${API_BASE_URL}/api/v1/grafana/instances`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Update an instance
  updateInstance: async (id: string, data: UpdateGrafanaInstanceRequest) => {
    const response = await axios.put<GrafanaInstance>(
      `${API_BASE_URL}/api/v1/grafana/instances/${id}`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Delete an instance
  deleteInstance: async (id: string) => {
    const response = await axios.delete<{ message: string }>(
      `${API_BASE_URL}/api/v1/grafana/instances/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Test an instance connection
  testInstance: async (data: TestGrafanaInstanceRequest) => {
    const response = await axios.post<TestGrafanaInstanceResponse>(
      `${API_BASE_URL}/api/v1/grafana/instances/test`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Sync an instance
  syncInstance: async (id: string, data: SyncGrafanaInstanceRequest) => {
    const response = await axios.post<SyncGrafanaInstanceResponse>(
      `${API_BASE_URL}/api/v1/grafana/instances/${id}/sync`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // ============== Dashboard Management ==============

  // List all dashboards
  getDashboards: async (params?: {
    instanceId?: string
    clusterId?: string
    tag?: string
    search?: string
    page?: number
    pageSize?: number
  }) => {
    const response = await axios.get<{
      data: GrafanaDashboard[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }>(`${API_BASE_URL}/api/v1/grafana/dashboards`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // Get a specific dashboard
  getDashboard: async (id: string) => {
    const response = await axios.get<GrafanaDashboard>(
      `${API_BASE_URL}/api/v1/grafana/dashboards/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // ============== Data Source Management ==============

  // List all data sources
  getDataSources: async (params?: {
    instanceId?: string
    type?: string
    page?: number
    pageSize?: number
  }) => {
    const response = await axios.get<{
      data: GrafanaDataSource[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }>(`${API_BASE_URL}/api/v1/grafana/datasources`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // Get a specific data source
  getDataSource: async (id: string) => {
    const response = await axios.get<GrafanaDataSource>(
      `${API_BASE_URL}/api/v1/grafana/datasources/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // ============== Folder Management ==============

  // List all folders
  getFolders: async (params?: {
    instanceId?: string
    page?: number
    pageSize?: number
  }) => {
    const response = await axios.get<{
      data: GrafanaFolder[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }>(`${API_BASE_URL}/api/v1/grafana/folders`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // Get a specific folder
  getFolder: async (id: string) => {
    const response = await axios.get<GrafanaFolder>(
      `${API_BASE_URL}/api/v1/grafana/folders/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },
}

export default grafanaApi
