// Audit log types
export interface AuditLog {
  id: string
  userId: string
  username: string
  action: string
  resource: string
  resourceId: string
  method: string
  path: string
  ipAddress: string
  userAgent: string
  statusCode: number
  errorMsg?: string
  oldValue?: string
  newValue?: string
  createdAt: string
}

export interface AuditLogFilter {
  username?: string
  action?: string
  resource?: string
  resourceId?: string
  startTime?: string
  endTime?: string
}

export interface AuditLogSummary {
  totalOperations: number
  userActivity: number
  failedOperations: number
  operationsByType: Record<string, number>
  operationsByResource: Record<string, number>
  topUsers: UserActivityStats[]
  topResources: ResourceActivityStats[]
}

export interface UserActivityStats {
  username: string
  operationCount: number
}

export interface ResourceActivityStats {
  resource: string
  operationCount: number
}

export interface AuditLogsResponse {
  data: {
    logs: AuditLog[]
    total: number
    page: number
    pageSize: number
  }
}

export interface AuditLogSummaryResponse {
  data: AuditLogSummary
}
