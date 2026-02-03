// Performance monitoring types
export interface PerformanceMetric {
  id: string
  metricType: string
  entityType: string
  entityId: string
  value: number
  unit: string
  timestamp: number
  labels: Record<string, string>
  createdAt: string
  updatedAt: string
}

export interface PerformanceSnapshot {
  id: string
  timeWindow: string
  startTime: number
  endTime: number
  entityCount: number
  avgCPUUsage: number
  maxCPUUsage: number
  avgMemoryUsage: number
  maxMemoryUsage: number
  avgDiskUsage: number
  maxDiskUsage: number
  avgResponseTime: number
  errorRate: number
  throughput: number
  createdAt: string
}

export interface SystemHealth {
  id: string
  overallStatus: 'healthy' | 'warning' | 'critical'
  hostCount: number
  clusterCount: number
  healthyHosts: number
  healthyClusters: number
  warningCount: number
  criticalCount: number
  timestamp: number
  createdAt: string
}

export interface TrendPoint {
  timestamp: number
  value: number
}

export interface MetricsResponse {
  data: {
    metrics: PerformanceMetric[]
    total: number
  }
}

export interface SystemHealthResponse {
  data: SystemHealth
}

export interface PerformanceSummaryResponse {
  data: PerformanceSnapshot
}

export interface TrendDataResponse {
  data: {
    metricType: string
    points: TrendPoint[]
  }
}

export interface MetricStatisticsResponse {
  data: {
    metricType: string
    avg: number
    min: number
    max: number
    count: number
  }
}
