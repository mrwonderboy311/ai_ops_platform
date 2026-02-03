// Cluster monitoring API client
import axios from 'axios'
import type {
  ClusterMetric,
  ClusterMetricSummary,
  NodeMetric,
  PodMetric,
  ClusterMetricsResponse,
  ClusterMetricsSummaryResponse,
  NodeMetricsResponse,
  PodMetricsResponse,
  NamespacesResponse,
} from '../types/clusterMonitoring'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const clusterMonitoringApi = {
  // Get cluster metrics history
  getClusterMetrics: async (clusterId: string, duration: string = '1h'): Promise<ClusterMetric[]> => {
    const response = await axios.get<ClusterMetricsResponse>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/metrics`,
      {
        params: { duration },
      }
    )
    return response.data.data
  },

  // Get cluster metrics summary (latest)
  getClusterMetricsSummary: async (clusterId: string): Promise<ClusterMetricSummary> => {
    const response = await axios.get<ClusterMetricsSummaryResponse>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/metrics/summary`
    )
    return response.data.data
  },

  // Get live cluster metrics
  getLiveClusterMetrics: async (clusterId: string): Promise<any> => {
    const response = await axios.get<{ data: any }>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/metrics/live`
    )
    return response.data.data
  },

  // Get node metrics
  getNodeMetrics: async (clusterId: string, duration: string = '1h'): Promise<NodeMetric[]> => {
    const response = await axios.get<NodeMetricsResponse>(
      `${API_BASE_URL}/api/v1/nodes/${clusterId}/metrics`,
      {
        params: { duration },
      }
    )
    return response.data.data
  },

  // Get live node metrics
  getLiveNodeMetrics: async (clusterId: string): Promise<any[]> => {
    const response = await axios.get<{ data: any[] }>(
      `${API_BASE_URL}/api/v1/nodes/${clusterId}/live-metrics`
    )
    return response.data.data
  },

  // Get pod metrics by namespace
  getPodMetrics: async (
    clusterId: string,
    namespace: string,
    duration: string = '1h'
  ): Promise<PodMetric[]> => {
    const response = await axios.get<PodMetricsResponse>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/namespaces/${namespace}/metrics`,
      {
        params: { duration },
      }
    )
    return response.data.data
  },

  // Get list of namespaces
  getNamespaces: async (clusterId: string): Promise<string[]> => {
    const response = await axios.get<NamespacesResponse>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/namespaces`
    )
    return response.data.data
  },

  // Refresh metrics
  refreshMetrics: async (clusterId: string): Promise<void> => {
    await axios.post(`${API_BASE_URL}/api/v1/clusters/${clusterId}/refresh`)
  },
}
