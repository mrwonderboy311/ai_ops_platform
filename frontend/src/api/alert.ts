// Alert API client
import axios from 'axios'
import type {
  AlertRule,
  Alert,
  AlertStatistics,
  AlertRulesResponse,
  AlertsResponse,
  EventsResponse,
} from '../types/alert'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const alertApi = {
  // Get alert rules
  getAlertRules: async (params?: {
    page?: number
    pageSize?: number
    enabled?: boolean
    targetType?: string
  }): Promise<AlertRulesResponse> => {
    const response = await axios.get<AlertRulesResponse>(`${API_BASE_URL}/api/v1/alert-rules`, { params })
    return response.data
  },

  // Create alert rule
  createAlertRule: async (rule: Partial<AlertRule>): Promise<AlertRule> => {
    const response = await axios.post<{ data: AlertRule }>(`${API_BASE_URL}/api/v1/alert-rules`, rule)
    return response.data.data
  },

  // Get alert rule
  getAlertRule: async (ruleId: string): Promise<AlertRule> => {
    const response = await axios.get<{ data: AlertRule }>(`${API_BASE_URL}/api/v1/alert-rules/${ruleId}`)
    return response.data.data
  },

  // Update alert rule
  updateAlertRule: async (ruleId: string, rule: Partial<AlertRule>): Promise<AlertRule> => {
    const response = await axios.put<{ data: AlertRule }>(`${API_BASE_URL}/api/v1/alert-rules/${ruleId}`, rule)
    return response.data.data
  },

  // Delete alert rule
  deleteAlertRule: async (ruleId: string): Promise<void> => {
    await axios.delete(`${API_BASE_URL}/api/v1/alert-rules/${ruleId}`)
  },

  // Get alerts
  getAlerts: async (params?: {
    page?: number
    pageSize?: number
    status?: string
    severity?: string
  }): Promise<AlertsResponse> => {
    const response = await axios.get<AlertsResponse>(`${API_BASE_URL}/api/v1/alerts`, { params })
    return response.data
  },

  // Get alert statistics
  getAlertStatistics: async (): Promise<AlertStatistics> => {
    const response = await axios.get<{ data: AlertStatistics }>(`${API_BASE_URL}/api/v1/alerts/statistics`)
    return response.data.data
  },

  // Silence alert
  silenceAlert: async (alertId: string, duration: string): Promise<Alert> => {
    const response = await axios.post<{ data: Alert }>(
      `${API_BASE_URL}/api/v1/alerts/${alertId}/silence`,
      { duration }
    )
    return response.data.data
  },

  // Get events
  getEvents: async (params?: {
    page?: number
    pageSize?: number
    type?: string
    severity?: string
  }): Promise<EventsResponse> => {
    const response = await axios.get<EventsResponse>(`${API_BASE_URL}/api/v1/events`, { params })
    return response.data
  },
}
