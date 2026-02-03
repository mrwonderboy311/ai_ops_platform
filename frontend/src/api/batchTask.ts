// Batch task API client
import axios from 'axios'
import type {
  BatchTask,
  BatchTaskResponse,
  CreateBatchTaskRequest,
  ExecuteBatchTaskRequest,
  ListBatchTasksResponse,
} from '../types/batchTask'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const batchTaskApi = {
  // Create a new batch task
  createBatchTask: async (request: CreateBatchTaskRequest): Promise<BatchTask> => {
    const token = localStorage.getItem('token')
    const response = await axios.post<{ data: BatchTask }>(
      `${API_BASE_URL}/api/v1/batch-tasks`,
      request,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data.data
  },

  // Execute a batch task
  executeBatchTask: async (request: ExecuteBatchTaskRequest): Promise<{ message: string; taskId: string }> => {
    const token = localStorage.getItem('token')
    const response = await axios.post<{ message: string; taskId: string }>(
      `${API_BASE_URL}/api/v1/batch-tasks/execute`,
      request,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data
  },

  // Get batch task details
  getBatchTask: async (taskId: string): Promise<BatchTaskResponse> => {
    const token = localStorage.getItem('token')
    const response = await axios.get<{ data: BatchTaskResponse }>(
      `${API_BASE_URL}/api/v1/batch-tasks/${taskId}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data.data
  },

  // List batch tasks
  listBatchTasks: async (params?: {
    status?: string
    type?: string
    page?: number
    pageSize?: number
  }): Promise<ListBatchTasksResponse> => {
    const token = localStorage.getItem('token')
    const response = await axios.get<{ data: ListBatchTasksResponse }>(
      `${API_BASE_URL}/api/v1/batch-tasks`,
      {
        params,
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data.data
  },

  // Cancel a batch task
  cancelBatchTask: async (taskId: string): Promise<{ message: string; taskId: string }> => {
    const token = localStorage.getItem('token')
    const response = await axios.post<{ message: string; taskId: string }>(
      `${API_BASE_URL}/api/v1/batch-tasks/cancel`,
      { taskId },
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data
  },

  // Delete a batch task
  deleteBatchTask: async (taskId: string): Promise<{ message: string; taskId: string }> => {
    const token = localStorage.getItem('token')
    const response = await axios.delete<{ message: string; taskId: string }>(
      `${API_BASE_URL}/api/v1/batch-tasks/${taskId}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return response.data
  },
}
