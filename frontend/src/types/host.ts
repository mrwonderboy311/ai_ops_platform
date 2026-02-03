// Host types
export type HostStatus = 'pending' | 'approved' | 'rejected' | 'offline' | 'online'

export interface Host {
  id: string
  hostname: string
  ipAddress: string
  port: number
  status: HostStatus
  osType: string
  osVersion: string
  cpuCores: number | null
  memoryGB: number | null
  diskGB: number | null
  labels: Record<string, string>
  tags: string[]
  clusterId: string | null
  registeredBy: string
  approvedBy: string | null
  approvedAt: string | null
  lastSeenAt: string | null
  createdAt: string
  updatedAt: string
  // Eager loaded associations
  registeredByUser?: {
    id: string
    username: string
    email: string
  }
  approvedByUser?: {
    id: string
    username: string
    email: string
  }
}

export interface CreateHostRequest {
  hostname: string
  ipAddress: string
  port: number
  osType: string
  osVersion: string
  cpuCores: number
  memoryGB: number
  diskGB: number
  labels: Record<string, string>
  tags: string[]
  clusterId: string
}

export interface UpdateHostRequest {
  hostname?: string
  port?: number
  osType?: string
  osVersion?: string
  cpuCores?: number
  memoryGB?: number
  diskGB?: number
  labels?: Record<string, string>
  tags?: string[]
  clusterId?: string
}

export interface HostListParams {
  page?: number
  pageSize?: number
  status?: HostStatus
  hostname?: string
  ipAddress?: string
  registeredBy?: string
  labels?: Record<string, string>
  tags?: string[]
  sortBy?: string
  sortDesc?: boolean
}

export interface HostListResponse {
  hosts: Host[]
  total: number
  page: number
  pageSize: number
}

export interface RejectHostRequest {
  reason?: string
}
