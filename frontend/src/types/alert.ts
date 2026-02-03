// Alert types
export type AlertSeverity = 'info' | 'warning' | 'critical'
export type AlertStatus = 'pending' | 'firing' | 'resolved' | 'silenced'

export interface AlertRule {
  id: string
  userId: string
  name: string
  description: string
  enabled: boolean
  targetType: 'host' | 'cluster' | 'node' | 'pod'
  targetId: string
  metricType: string
  operator: '>' | '<' | '>=' | '<=' | '==' | '!='
  threshold: number
  duration: number
  severity: AlertSeverity
  notifyEmail: boolean
  notifyWebhook: boolean
  webhookUrl: string
  silencedUntil?: string
  lastEvaluatedAt?: string
  createdAt: string
  updatedAt: string
}

export interface Alert {
  id: string
  ruleId: string
  userId: string
  clusterId?: string
  hostId?: string
  status: AlertStatus
  severity: AlertSeverity
  title: string
  description: string
  value: number
  threshold: number
  startedAt: string
  updatedAt: string
  resolvedAt?: string
  silencedUntil?: string
  labels: string
  annotations: string
}

export interface Event {
  id: string
  clusterId?: string
  hostId?: string
  type: string
  severity: string
  title: string
  message: string
  metadata: string
  createdAt: string
}

export interface AlertStatistics {
  totalAlerts: number
  firingAlerts: number
  resolvedAlerts: number
  criticalAlerts: number
  warningAlerts: number
}

export interface AlertRulesResponse {
  data: {
    rules: AlertRule[]
    total: number
    page: number
    pageSize: number
  }
}

export interface AlertsResponse {
  data: {
    alerts: Alert[]
    total: number
    page: number
    pageSize: number
  }
}

export interface EventsResponse {
  data: {
    events: Event[]
    total: number
    page: number
    pageSize: number
  }
}
