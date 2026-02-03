import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export interface PrometheusDataSource {
  id: string
  userId: string
  clusterId?: string
  name: string
  url: string
  username?: string
  status: 'active' | 'inactive' | 'error'
  insecureSkipTLS: boolean
  caCert?: string
  clientCert?: string
  lastTestAt?: string
  lastTestStatus?: string
  lastTestError?: string
  queryCount: number
  lastQueriedAt?: string
  averageQueryTime: number
  createdAt: string
  updatedAt: string
  cluster?: {
    id: string
    name: string
  }
}

export interface CreatePrometheusDataSourceRequest {
  clusterId?: string
  name: string
  url: string
  username?: string
  password?: string
  insecureSkipTLS?: boolean
  caCert?: string
  clientCert?: string
  clientKey?: string
  headers?: string
}

export interface UpdatePrometheusDataSourceRequest {
  name?: string
  url?: string
  username?: string
  password?: string
  insecureSkipTLS?: boolean
  caCert?: string
  clientCert?: string
  clientKey?: string
  headers?: string
  status?: 'active' | 'inactive' | 'error'
}

export interface TestPrometheusDataSourceRequest {
  url: string
  username?: string
  password?: string
  insecureSkipTLS?: boolean
  caCert?: string
  clientCert?: string
  clientKey?: string
}

export interface TestPrometheusDataSourceResponse {
  success: boolean
  version?: string
  message: string
  error?: string
  duration: number
}

export interface PrometheusAlertRule {
  id: string
  userId: string
  dataSourceId: string
  clusterId?: string
  name: string
  expression: string
  duration: number
  severity: 'critical' | 'warning' | 'info'
  summary?: string
  description?: string
  labels?: string
  annotations?: string
  enabled: boolean
  synced: boolean
  syncedAt?: string
  syncError?: string
  triggerCount: number
  lastTriggeredAt?: string
  createdAt: string
  updatedAt: string
  dataSource?: PrometheusDataSource
  cluster?: {
    id: string
    name: string
  }
}

export interface CreatePrometheusAlertRuleRequest {
  dataSourceId: string
  clusterId?: string
  name: string
  expression: string
  duration: number
  severity: 'critical' | 'warning' | 'info'
  summary?: string
  description?: string
  labels?: string
  annotations?: string
}

export interface UpdatePrometheusAlertRuleRequest {
  name?: string
  expression?: string
  duration?: number
  severity?: 'critical' | 'warning' | 'info'
  summary?: string
  description?: string
  labels?: string
  annotations?: string
  enabled?: boolean
}

export interface PrometheusQueryRequest {
  query: string
  queryType: 'instant' | 'range'
  startTime?: string
  endTime?: string
  step?: string
}

export interface PrometheusQueryResponse {
  status: string
  data?: PrometheusSeries[]
  error?: string
  duration: number
}

export interface PrometheusSeries {
  metric: Record<string, string>
  values?: PrometheusValue[]
  value?: PrometheusValue
}

export interface PrometheusValue {
  timestamp: number
  value: string
}

export interface PrometheusDashboard {
  id: string
  userId: string
  clusterId?: string
  name: string
  description?: string
  tags?: string
  config: string
  isPublic: boolean
  starred: boolean
  refreshRate: number
  createdAt: string
  updatedAt: string
  cluster?: {
    id: string
    name: string
  }
}

export interface CreatePrometheusDashboardRequest {
  clusterId?: string
  name: string
  description?: string
  tags?: string
  config: string
  isPublic?: boolean
  refreshRate?: number
}

export interface UpdatePrometheusDashboardRequest {
  name?: string
  description?: string
  tags?: string
  config?: string
  isPublic?: boolean
  refreshRate?: number
  starred?: boolean
}

const prometheusApi = {
  // ============== Data Source Management ==============

  // List all Prometheus data sources
  getDataSources: async (params?: {
    clusterId?: string
    status?: string
    page?: number
    pageSize?: number
  }) => {
    const response = await axios.get<{
      data: PrometheusDataSource[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }>(`${API_BASE_URL}/api/v1/prometheus/datasources`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // Get a specific data source
  getDataSource: async (id: string) => {
    const response = await axios.get<PrometheusDataSource>(
      `${API_BASE_URL}/api/v1/prometheus/datasources/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Create a new data source
  createDataSource: async (data: CreatePrometheusDataSourceRequest) => {
    const response = await axios.post<PrometheusDataSource>(
      `${API_BASE_URL}/api/v1/prometheus/datasources`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Update a data source
  updateDataSource: async (
    id: string,
    data: UpdatePrometheusDataSourceRequest
  ) => {
    const response = await axios.put<PrometheusDataSource>(
      `${API_BASE_URL}/api/v1/prometheus/datasources/${id}`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Delete a data source
  deleteDataSource: async (id: string) => {
    const response = await axios.delete<{ message: string }>(
      `${API_BASE_URL}/api/v1/prometheus/datasources/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Test a data source connection
  testDataSource: async (data: TestPrometheusDataSourceRequest) => {
    const response = await axios.post<TestPrometheusDataSourceResponse>(
      `${API_BASE_URL}/api/v1/prometheus/datasources/test`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // ============== Alert Rule Management ==============

  // List all alert rules
  getAlertRules: async (params?: {
    dataSourceId?: string
    clusterId?: string
    severity?: string
    enabled?: string
    page?: number
    pageSize?: number
  }) => {
    const response = await axios.get<{
      data: PrometheusAlertRule[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }>(`${API_BASE_URL}/api/v1/prometheus/alert-rules`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // Get a specific alert rule
  getAlertRule: async (id: string) => {
    const response = await axios.get<PrometheusAlertRule>(
      `${API_BASE_URL}/api/v1/prometheus/alert-rules/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Create a new alert rule
  createAlertRule: async (data: CreatePrometheusAlertRuleRequest) => {
    const response = await axios.post<PrometheusAlertRule>(
      `${API_BASE_URL}/api/v1/prometheus/alert-rules`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Update an alert rule
  updateAlertRule: async (
    id: string,
    data: UpdatePrometheusAlertRuleRequest
  ) => {
    const response = await axios.put<PrometheusAlertRule>(
      `${API_BASE_URL}/api/v1/prometheus/alert-rules/${id}`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Delete an alert rule
  deleteAlertRule: async (id: string) => {
    const response = await axios.delete<{ message: string }>(
      `${API_BASE_URL}/api/v1/prometheus/alert-rules/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // ============== Query Execution ==============

  // Execute a Prometheus query
  executeQuery: async (dataSourceId: string, data: PrometheusQueryRequest) => {
    const response = await axios.post<PrometheusQueryResponse>(
      `${API_BASE_URL}/api/v1/prometheus/datasources/${dataSourceId}/query`,
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
    clusterId?: string
    starred?: string
    page?: number
    pageSize?: number
  }) => {
    const response = await axios.get<{
      data: PrometheusDashboard[]
      total: number
      page: number
      pageSize: number
      totalPages: number
    }>(`${API_BASE_URL}/api/v1/prometheus/dashboards`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      params,
    })
    return response.data
  },

  // Get a specific dashboard
  getDashboard: async (id: string) => {
    const response = await axios.get<PrometheusDashboard>(
      `${API_BASE_URL}/api/v1/prometheus/dashboards/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Create a new dashboard
  createDashboard: async (data: CreatePrometheusDashboardRequest) => {
    const response = await axios.post<PrometheusDashboard>(
      `${API_BASE_URL}/api/v1/prometheus/dashboards`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Update a dashboard
  updateDashboard: async (
    id: string,
    data: UpdatePrometheusDashboardRequest
  ) => {
    const response = await axios.put<PrometheusDashboard>(
      `${API_BASE_URL}/api/v1/prometheus/dashboards/${id}`,
      data,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },

  // Delete a dashboard
  deleteDashboard: async (id: string) => {
    const response = await axios.delete<{ message: string }>(
      `${API_BASE_URL}/api/v1/prometheus/dashboards/${id}`,
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      }
    )
    return response.data
  },
}

export default prometheusApi
