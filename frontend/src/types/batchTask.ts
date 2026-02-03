// Batch task types

export type BatchTaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
export type BatchTaskType = 'command' | 'script' | 'file_op'
export type TaskExecutionStrategy = 'parallel' | 'serial' | 'rolling'

export interface BatchTask {
  id: string
  userId: string
  name: string
  description: string
  type: BatchTaskType
  status: BatchTaskStatus
  strategy: TaskExecutionStrategy
  command: string
  script: string
  timeout: number
  maxRetries: number
  parallelism: number
  totalHosts: number
  completedHosts: number
  failedHosts: number
  startedAt?: string
  completedAt?: string
  createdAt: string
  updatedAt: string
}

export interface BatchTaskHost {
  id: string
  batchTaskId: string
  hostId: string
  status: BatchTaskStatus
  exitCode?: number
  stdout: string
  stderr: string
  duration: number
  errorMessage: string
  retryCount: number
  startedAt?: string
  completedAt?: string
  createdAt: string
  batchTask?: BatchTask
  host?: {
    id: string
    hostname: string
    ipAddress: string
  }
}

export interface BatchTaskResponse {
  batchTask: BatchTask
  hosts: BatchTaskHost[]
  progress: number
}

export interface CreateBatchTaskRequest {
  name: string
  description?: string
  type: BatchTaskType
  strategy: TaskExecutionStrategy
  command?: string
  script?: string
  timeout?: number
  maxRetries?: number
  parallelism?: number
  hostIds: string[]
}

export interface ExecuteBatchTaskRequest {
  taskId: string
  hostIds?: string[]
}

export interface ListBatchTasksResponse {
  tasks: BatchTask[]
  total: number
  page: number
  pageSize: number
}
