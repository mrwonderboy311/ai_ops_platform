import { apiClient } from './client'
import type {
  FileTransfer,
  ListDirectoryRequest,
  ListDirectoryResponse,
  FileUploadRequest,
  FileDownloadRequest,
  FileDeleteRequest,
  CreateDirectoryRequest,
  FileRenameRequest,
} from '../types/file'

// Get auth token
const getAuthHeaders = () => {
  const auth = localStorage.getItem('myops-auth')
  if (auth) {
    try {
      const { token } = JSON.parse(auth)
      return {
        Authorization: `Bearer ${token}`,
      }
    } catch {
      return {}
    }
  }
  return {}
}

export const fileApi = {
  // List directory contents
  listDirectory: async (request: ListDirectoryRequest): Promise<ListDirectoryResponse> => {
    const response = await apiClient.post<{ data: ListDirectoryResponse }>('/api/v1/files/list', request, {
      headers: getAuthHeaders(),
    })
    return response.data.data
  },

  // Upload file
  uploadFile: async (request: FileUploadRequest): Promise<{
    transferId: string
    fileName: string
    size: number
    targetPath: string
  }> => {
    const formData = new FormData()
    formData.append('hostId', request.hostId)
    formData.append('remotePath', request.remotePath)
    formData.append('username', request.username)
    if (request.password) {
      formData.append('password', request.password)
    }
    if (request.key) {
      formData.append('key', request.key)
    }
    if (request.overwrite !== undefined) {
      formData.append('overwrite', String(request.overwrite))
    }
    formData.append('file', request.file)

    const response = await apiClient.post<{
      data: {
        transferId: string
        fileName: string
        size: number
        targetPath: string
      }
    }>('/api/v1/files/upload', formData, {
      headers: {
        ...getAuthHeaders(),
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data.data
  },

  // Download file
  downloadFile: async (request: FileDownloadRequest): Promise<Blob> => {
    const response = await apiClient.post('/api/v1/files/download', request, {
      headers: getAuthHeaders(),
      responseType: 'blob',
    })
    return response.data
  },

  // Delete file
  deleteFile: async (request: FileDeleteRequest): Promise<{ message: string; path: string }> => {
    const response = await apiClient.post<{
      data: { message: string; path: string }
    }>('/api/v1/files/delete', request, {
      headers: getAuthHeaders(),
    })
    return response.data.data
  },

  // Create directory
  createDirectory: async (request: CreateDirectoryRequest): Promise<{ message: string; path: string }> => {
    const response = await apiClient.post<{
      data: { message: string; path: string }
    }>('/api/v1/files/mkdir', request, {
      headers: getAuthHeaders(),
    })
    return response.data.data
  },

  // Rename file
  renameFile: async (request: FileRenameRequest): Promise<{ message: string }> => {
    const response = await apiClient.post<{
      data: { message: string }
    }>('/api/v1/files/rename', request, {
      headers: getAuthHeaders(),
    })
    return response.data.data
  },

  // Get transfer history
  getTransfers: async (params?: {
    hostId?: string
    direction?: string
    status?: string
  }): Promise<FileTransfer[]> => {
    const queryParams = new URLSearchParams()
    if (params?.hostId) queryParams.append('hostId', params.hostId)
    if (params?.direction) queryParams.append('direction', params.direction)
    if (params?.status) queryParams.append('status', params.status)

    const response = await apiClient.get<{ data: FileTransfer[] }>(
      `/api/v1/files/transfers?${queryParams.toString()}`,
      { headers: getAuthHeaders() }
    )
    return response.data.data
  },

  // Get transfer history for a specific host
  getHostTransfers: async (hostId: string): Promise<FileTransfer[]> => {
    const response = await apiClient.get<{ data: FileTransfer[] }>(
      `/api/v1/hosts/${hostId}/transfers`,
      { headers: getAuthHeaders() }
    )
    return response.data.data
  },
}
