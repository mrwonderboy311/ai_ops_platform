// Performance API client
import axios from 'axios'
import type {
  MetricsResponse,
  SystemHealthResponse,
  PerformanceSummaryResponse,
  TrendDataResponse,
  MetricStatisticsResponse,
} from '../types/performance'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const performanceApi = {
  // Get performance metrics
  getMetrics: async (params?: {
    metricType?: string
    entityType?: string
    entityId?: string
    startTime?: number
    endTime?: number
  }): Promise<MetricsResponse> => {
    const response = await axios.get<MetricsResponse>(`${API_BASE_URL}/api/v1/performance/metrics`, { params })
    return response.data
  },

  // Get system health
  getSystemHealth: async (): Promise<SystemHealthResponse> => {
    const response = await axios.get<SystemHealthResponse>(`${API_BASE_URL}/api/v1/performance/health`)
    return response.data
  },

  // Refresh system health
  refreshSystemHealth: async (): Promise<SystemHealthResponse> => {
    const response = await axios.post<SystemHealthResponse>(`${API_BASE_URL}/api/v1/performance/health/refresh`)
    return response.data
  },

  // Get performance summary
  getPerformanceSummary: async (timeWindow?: string): Promise<PerformanceSummaryResponse> => {
    const response = await axios.get<PerformanceSummaryResponse>(`${API_BASE_URL}/api/v1/performance/summary`, {
      params: { timeWindow },
    })
    return response.data
  },

  // Get trend data
  getTrendData: async (params?: {
    metricType?: string
    entityType?: string
    points?: number
  }): Promise<TrendDataResponse> => {
    const response = await axios.get<TrendDataResponse>(`${API_BASE_URL}/api/v1/performance/trend`, { params })
    return response.data
  },

  // Get metric statistics
  getMetricStatistics: async (params?: {
    metricType?: string
    duration?: number
  }): Promise<MetricStatisticsResponse> => {
    const response = await axios.get<MetricStatisticsResponse>(`${API_BASE_URL}/api/v1/performance/statistics`, { params })
    return response.data
  },
}
