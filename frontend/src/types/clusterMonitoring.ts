// Cluster monitoring types

export interface ClusterMetric {
  id: string
  clusterId: string
  timestamp: number
  cpuUsagePercent: number
  memoryUsageBytes: number
  memoryTotalBytes: number
  podCount: number
  runningPodCount: number
  pendingPodCount: number
  failedPodCount: number
  nodeCount: number
  readyNodeCount: number
  createdAt: string
}

export interface ClusterMetricSummary {
  timestamp: number
  cpuUsagePercent: number
  memoryUsagePercent: number
  podCount: number
  runningPodCount: number
  pendingPodCount: number
  failedPodCount: number
  nodeCount: number
  readyNodeCount: number
}

export interface NodeMetric {
  id: string
  clusterId: string
  nodeName: string
  timestamp: number
  cpuUsagePercent: number
  memoryUsageBytes: number
  memoryTotalBytes: number
  diskUsageBytes: number
  diskTotalBytes: number
  podCount: number
  networkRxBytes: number
  networkTxBytes: number
  status: string
  ready: boolean
  createdAt: string
}

export interface PodMetric {
  id: string
  clusterId: string
  namespace: string
  podName: string
  timestamp: number
  cpuUsageCores: number
  memoryUsageBytes: number
  restartCount: number
  status: string
  ready: boolean
  nodeName: string
  createdAt: string
}

export interface ClusterMetricsResponse {
  data: ClusterMetric[]
}

export interface ClusterMetricsSummaryResponse {
  data: ClusterMetricSummary
}

export interface NodeMetricsResponse {
  data: NodeMetric[]
}

export interface PodMetricsResponse {
  data: PodMetric[]
}

export interface NamespacesResponse {
  data: string[]
}
