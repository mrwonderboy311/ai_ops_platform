// Notification types
export type NotificationType = 'alert' | 'system' | 'task' | 'security' | 'info'

export type NotificationPriority = 'low' | 'medium' | 'high' | 'critical'

export type NotificationStatus = 'pending' | 'sent' | 'failed' | 'delivered'

export interface Notification {
  id: string
  userId: string
  type: NotificationType
  title: string
  message: string
  priority: NotificationPriority
  status: NotificationStatus
  read: boolean
  readAt?: string
  actionUrl?: string
  actionLabel?: string
  metadata?: Record<string, string>
  expiresAt?: string
  createdAt: string
  updatedAt: string
}

export interface NotificationPreference {
  id: string
  userId: string
  emailEnabled: boolean
  webEnabled: boolean
  pushEnabled: boolean
  alertTypes: NotificationType[]
  minPriority: NotificationPriority
  quietHoursStart: string
  quietHoursEnd: string
  timezone: string
  createdAt: string
  updatedAt: string
}

export interface NotificationStats {
  total: number
  unread: number
  alert: number
  system: number
  task: number
  security: number
  info: number
}

export interface NotificationsResponse {
  data: {
    notifications: Notification[]
    count: number
  }
}

export interface UnreadCountResponse {
  data: {
    unreadCount: number
  }
}

export interface NotificationStatsResponse {
  data: NotificationStats
}

export interface NotificationPreferenceResponse {
  data: NotificationPreference
}
