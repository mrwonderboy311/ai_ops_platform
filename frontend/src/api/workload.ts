// Workload API client for Kubernetes resources
import axios from 'axios'
import type {
  K8sDeployment,
  K8sPod,
  K8sService,
  PodLogsResponse,
  NamespacesResponse,
  DeploymentsResponse,
  PodsResponse,
  ServicesResponse,
} from '../types/workload'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const workloadApi = {
  // Get namespaces
  getNamespaces: async (clusterId: string): Promise<string[]> => {
    const response = await axios.get<NamespacesResponse>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/namespaces`
    )
    return response.data.data
  },

  // Get deployments
  getDeployments: async (clusterId: string, namespace: string): Promise<K8sDeployment[]> => {
    const response = await axios.get<DeploymentsResponse>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/namespaces/${namespace}/deployments`
    )
    return response.data.data
  },

  // Get pods
  getPods: async (clusterId: string, namespace: string): Promise<K8sPod[]> => {
    const response = await axios.get<PodsResponse>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/namespaces/${namespace}/pods`
    )
    return response.data.data
  },

  // Get pod detail
  getPodDetail: async (clusterId: string, namespace: string, podName: string): Promise<any> => {
    const response = await axios.get<{ data: any }>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/namespaces/${namespace}/pods/${podName}/detail`
    )
    return response.data.data
  },

  // Get services
  getServices: async (clusterId: string, namespace: string): Promise<K8sService[]> => {
    const response = await axios.get<ServicesResponse>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/namespaces/${namespace}/services`
    )
    return response.data.data
  },

  // Get pod logs
  getPodLogs: async (clusterId: string, namespace: string, podName: string, tailLines?: number): Promise<string> => {
    const response = await axios.get<PodLogsResponse>(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/namespaces/${namespace}/pods/${podName}/logs`,
      {
        params: { tailLines },
      }
    )
    return response.data.data.logs
  },

  // Delete pod
  deletePod: async (clusterId: string, namespace: string, podName: string): Promise<void> => {
    await axios.delete(
      `${API_BASE_URL}/api/v1/clusters/${clusterId}/namespaces/${namespace}/pods/${podName}`
    )
  },
}
