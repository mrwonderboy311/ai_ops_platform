// Notification API client
import axios from 'axios'
import type {
  NotificationPreference,
  NotificationsResponse,
  UnreadCountResponse,
  NotificationStatsResponse,
  NotificationPreferenceResponse,
} from '../types/notification'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const notificationApi = {
  // Get notifications
  getNotifications: async (params?: {
    limit?: number
    unreadOnly?: boolean
  }): Promise<NotificationsResponse> => {
    const response = await axios.get<NotificationsResponse>(`${API_BASE_URL}/api/v1/notifications`, { params })
    return response.data
  },

  // Get unread count
  getUnreadCount: async (): Promise<UnreadCountResponse> => {
    const response = await axios.get<UnreadCountResponse>(`${API_BASE_URL}/api/v1/notifications/unread-count`)
    return response.data
  },

  // Mark as read
  markAsRead: async (notificationId: string): Promise<{ data: { success: boolean } }> => {
    const response = await axios.post(`${API_BASE_URL}/api/v1/notifications/${notificationId}/read`)
    return response.data
  },

  // Mark all as read
  markAllAsRead: async (): Promise<{ data: { success: boolean } }> => {
    const response = await axios.post(`${API_BASE_URL}/api/v1/notifications/mark-all-read`)
    return response.data
  },

  // Delete notification
  deleteNotification: async (notificationId: string): Promise<{ data: { success: boolean } }> => {
    const response = await axios.delete(`${API_BASE_URL}/api/v1/notifications/${notificationId}`)
    return response.data
  },

  // Get notification preference
  getNotificationPreference: async (): Promise<NotificationPreferenceResponse> => {
    const response = await axios.get<NotificationPreferenceResponse>(`${API_BASE_URL}/api/v1/notifications/preferences`)
    return response.data
  },

  // Update notification preference
  updateNotificationPreference: async (pref: Partial<NotificationPreference>): Promise<NotificationPreferenceResponse> => {
    const response = await axios.put<NotificationPreferenceResponse>(`${API_BASE_URL}/api/v1/notifications/preferences`, pref)
    return response.data
  },

  // Get notification stats
  getNotificationStats: async (): Promise<NotificationStatsResponse> => {
    const response = await axios.get<NotificationStatsResponse>(`${API_BASE_URL}/api/v1/notifications/stats`)
    return response.data
  },
}
