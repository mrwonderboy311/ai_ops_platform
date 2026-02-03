// Kubernetes cluster API client
import axios from 'axios'
import type {
  K8sCluster,
  ClusterNode,
  ClusterInfo,
  CreateClusterRequest,
  UpdateClusterRequest,
  ClusterConnectionTestRequest,
  ClusterConnectionTestResponse,
  ListClustersResponse,
} from '../types/cluster'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const clusterApi = {
  // List clusters
  listClusters: async (params?: {
    status?: string
    type?: string
    provider?: string
    page?: number
    pageSize?: number
  }): Promise<ListClustersResponse> => {
    const token = localStorage.getItem('token')
    const response = await axios.get<{ data: ListClustersResponse }>(
      `${API_BASE_URL}/api/v1/clusters`,
      {
        params,
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data.data
  },

  // Get cluster details
  getCluster: async (clusterId: string): Promise<K8sCluster> => {
    const token = localStorage.getItem('token')
    const response = await axios.get<{ data: K8sCluster }>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data.data
  },

  // Create cluster
  createCluster: async (request: CreateClusterRequest): Promise<K8sCluster> => {
    const token = localStorage.getItem('token')
    const response = await axios.post<{ data: K8sCluster }>(
      `${API_BASE_URL}/api/v1/clusters`,
      request,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data.data
  },

  // Update cluster
  updateCluster: async (clusterId: string, request: UpdateClusterRequest): Promise<K8sCluster> => {
    const token = localStorage.getItem('token')
    const response = await axios.put<{ data: K8sCluster }>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}`,
      request,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data.data
  },

  // Delete cluster
  deleteCluster: async (clusterId: string): Promise<{ message: string; clusterId: string }> => {
    const token = localStorage.getItem('token')
    const response = await axios.delete<{ message: string; clusterId: string }>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data
  },

  // Test cluster connection
  testConnection: async (request: ClusterConnectionTestRequest): Promise<ClusterConnectionTestResponse> => {
    const token = localStorage.getItem('token')
    const response = await axios.post<{ data: ClusterConnectionTestResponse }>(
      `${API_BASE_URL}/api/v1/clusters/test-connection`,
      request,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data.data
  },

  // Get cluster nodes
  getClusterNodes: async (clusterId: string): Promise<ClusterNode[]> => {
    const token = localStorage.getItem('token')
    const response = await axios.get<{ data: ClusterNode[] }>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/nodes`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data.data
  },

  // Get cluster info
  getClusterInfo: async (clusterId: string): Promise<ClusterInfo> => {
    const token = localStorage.getItem('token')
    const response = await axios.get<{ data: ClusterInfo }>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/info`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data.data
  },
}
