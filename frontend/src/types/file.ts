// File transfer types
export type FileTransferStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
export type FileTransferDirection = 'upload' | 'download'

export interface FileInfo {
  name: string
  path: string
  size: number
  mode: string
  modTime: string
  isDir: boolean
  owner?: string
  group?: string
  permissions: string
}

export interface FileTransfer {
  id: string
  hostId: string
  host?: Host
  userId: string
  user?: User
  direction: FileTransferDirection
  sourcePath: string
  targetPath: string
  fileName: string
  fileSize: number
  transferred: number
  status: FileTransferStatus
  errorMessage?: string
  startedAt?: string
  completedAt?: string
  createdAt: string
  updatedAt: string
}

export interface ListDirectoryRequest {
  hostId: string
  path: string
  username: string
  password?: string
  key?: string
}

export interface ListDirectoryResponse {
  path: string
  parent?: string
  files: FileInfo[]
  totalSize: number
  fileCount: number
  dirCount: number
}

export interface FileUploadRequest {
  hostId: string
  remotePath: string
  username: string
  password?: string
  key?: string
  overwrite?: boolean
  file: File
}

export interface FileDownloadRequest {
  hostId: string
  remotePath: string
  username: string
  password?: string
  key?: string
}

export interface FileDeleteRequest {
  hostId: string
  remotePath: string
  username: string
  password?: string
  key?: string
}

export interface CreateDirectoryRequest {
  hostId: string
  path: string
  mode?: string
  username: string
  password?: string
  key?: string
}

export interface FileRenameRequest {
  hostId: string
  oldPath: string
  newPath: string
  username: string
  password?: string
  key?: string
}

// Import Host type
import type { Host } from './host'

// Simple User interface (expand as needed)
interface User {
  id: string
  username: string
  email?: string
}
