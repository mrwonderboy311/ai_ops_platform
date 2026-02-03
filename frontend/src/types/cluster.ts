// Kubernetes cluster types

export type ClusterStatus = 'pending' | 'connected' | 'error' | 'disabled'
export type ClusterType = 'managed' | 'self-hosted'

export interface K8sCluster {
  id: string
  userId: string
  name: string
  description: string
  type: ClusterType
  status: ClusterStatus
  endpoint: string
  version: string
  nodeCount: number
  region: string
  provider: string
  lastConnectedAt?: string
  errorMessage: string
  createdAt: string
  updatedAt: string
}

export interface ClusterNode {
  id: string
  clusterId: string
  name: string
  internalIp: string
  externalIp: string
  status: string
  roles: string
  version: string
  osImage: string
  kernelVersion: string
  containerRuntime: string
  cpuCapacity: string
  memoryCapacity: string
  storageCapacity: string
  cpuAllocatable: string
  memoryAllocatable: string
  storageAllocatable: string
  podCount: number
  conditions: string
  createdAt: string
  updatedAt: string
}

export interface ClusterNamespace {
  id: string
  clusterId: string
  name: string
  status: string
  createdAt: string
}

export interface ClusterConnectionTestRequest {
  kubeconfig: string
  endpoint?: string
}

export interface ClusterConnectionTestResponse {
  success: boolean
  version?: string
  nodeCount?: number
  error?: string
}

export interface CreateClusterRequest {
  name: string
  description?: string
  type: ClusterType
  endpoint?: string
  kubeconfig: string
  region?: string
  provider?: string
}

export interface UpdateClusterRequest {
  name?: string
  description?: string
  endpoint?: string
  kubeconfig?: string
}

export interface ClusterInfo {
  version: string
  nodeCount: number
  namespaceCount: number
  podCount: number
  deploymentCount: number
  serviceCount: number
  ingressCount: number
  configMapCount: number
  secretCount: number
}

export interface ListClustersResponse {
  clusters: K8sCluster[]
  total: number
  page: number
  pageSize: number
}
