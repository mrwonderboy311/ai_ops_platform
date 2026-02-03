// Audit API client
import axios from 'axios'
import type {
  AuditLogsResponse,
  AuditLogSummaryResponse,
} from '../types/audit'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const auditApi = {
  // Get audit logs
  getAuditLogs: async (params?: {
    page?: number
    pageSize?: number
    username?: string
    action?: string
    resource?: string
    resourceId?: string
    startTime?: string
    endTime?: string
  }): Promise<AuditLogsResponse> => {
    const response = await axios.get<AuditLogsResponse>(`${API_BASE_URL}/api/v1/audit-logs`, { params })
    return response.data
  },

  // Get audit log summary
  getAuditLogSummary: async (duration?: string): Promise<AuditLogSummaryResponse> => {
    const response = await axios.get<AuditLogSummaryResponse>(
      `${API_BASE_URL}/api/v1/audit-logs/summary`,
      { params: { duration } }
    )
    return response.data
  },

  // Get user activity
  getUserActivity: async (params?: {
    userId?: string
    duration?: string
  }): Promise<{ data: any[] }> => {
    const response = await axios.get(`${API_BASE_URL}/api/v1/audit-logs/user-activity`, { params })
    return response.data
  },

  // Get resource activity
  getResourceActivity: async (params?: {
    duration?: string
  }): Promise<{ data: any[] }> => {
    const response = await axios.get(`${API_BASE_URL}/api/v1/audit-logs/resource-activity`, { params })
    return response.data
  },
}
