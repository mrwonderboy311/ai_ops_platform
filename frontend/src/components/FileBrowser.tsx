import { useState } from 'react'
import {
  Table,
  Button,
  Input,
  Space,
  Breadcrumb,
  Modal,
  message,
  Popconfirm,
  Tooltip,
} from 'antd'
import {
  FileOutlined,
  FolderOutlined,
  ArrowUpOutlined,
  ReloadOutlined,
  DeleteOutlined,
  DownloadOutlined,
  UploadOutlined,
  FolderAddOutlined,
  HomeOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { FileInfo } from '../types/file'
import { FileAuthModal } from './FileAuthModal'

interface FileBrowserProps {
  hostId: string
  hostName: string
}

export const FileBrowser: React.FC<FileBrowserProps> = ({
  hostId,
  hostName,
}) => {
  const [currentPath, setCurrentPath] = useState('/')
  const [files, setFiles] = useState<FileInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
  const [authModalVisible, setAuthModalVisible] = useState(false)
  const [pendingAction, setPendingAction] = useState<
    'list' | 'upload' | 'download' | 'delete' | 'mkdir' | null
  >(null)
  const [mkdirModalVisible, setMkdirModalVisible] = useState(false)
  const [newDirName, setNewDirName] = useState('')
  const [credentials, setCredentials] = useState<{
    username: string
    password: string
    key: string
  }>({ username: 'root', password: '', key: '' })

  // Load directory contents
  const loadDirectory = async (path: string) => {
    setLoading(true)
    try {
      const { fileApi } = await import('../api/file')
      const response = await fileApi.listDirectory({
        hostId,
        path,
        username: credentials.username,
        password: credentials.password,
        key: credentials.key,
      })
      setFiles(response.files)
      setCurrentPath(response.path)
    } catch (error: any) {
      if (error.response?.status === 401) {
        // Need authentication
        setPendingAction('list')
        setAuthModalVisible(true)
      } else {
        message.error(`Failed to list directory: ${error.message || 'Unknown error'}`)
      }
    } finally {
      setLoading(false)
    }
  }

  // Handle file/folder click
  const handleItemClick = (file: FileInfo) => {
    if (file.isDir) {
      loadDirectory(file.path)
    }
  }

  // Handle breadcrumb navigation
  const handleBreadcrumbNavigate = (path: string) => {
    loadDirectory(path)
  }

  // Handle parent directory
  const handleParent = () => {
    const parentPath = currentPath.split('/').slice(0, -1).join('/') || '/'
    loadDirectory(parentPath)
  }

  // Handle refresh
  const handleRefresh = () => {
    loadDirectory(currentPath)
  }

  // Handle upload
  const handleUpload = () => {
    setPendingAction('upload')
    setAuthModalVisible(true)
  }

  // Handle download
  const handleDownload = async (file: FileInfo) => {
    setLoading(true)
    try {
      const { fileApi } = await import('../api/file')
      const blob = await fileApi.downloadFile({
        hostId,
        remotePath: file.path,
        username: credentials.username,
        password: credentials.password,
        key: credentials.key,
      })

      // Create download link
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = file.name
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      window.URL.revokeObjectURL(url)

      message.success(`Downloaded: ${file.name}`)
    } catch (error: any) {
      message.error(`Failed to download: ${error.message || 'Unknown error'}`)
    } finally {
      setLoading(false)
    }
  }

  // Handle delete
  const handleDelete = async () => {
    setLoading(true)
    try {
      const { fileApi } = await import('../api/file')
      for (const key of selectedRowKeys) {
        const file = files.find(f => f.path === key)
        if (file) {
          await fileApi.deleteFile({
            hostId,
            remotePath: file.path,
            username: credentials.username,
            password: credentials.password,
            key: credentials.key,
          })
        }
      }
      message.success(`Deleted ${selectedRowKeys.length} item(s)`)
      setSelectedRowKeys([])
      loadDirectory(currentPath)
    } catch (error: any) {
      message.error(`Failed to delete: ${error.message || 'Unknown error'}`)
    } finally {
      setLoading(false)
    }
  }

  // Handle create directory
  const handleMkdir = () => {
    setPendingAction('mkdir')
    setAuthModalVisible(true)
  }

  const handleCreateDirectory = async () => {
    if (!newDirName.trim()) {
      message.error('Please enter a directory name')
      return
    }

    setLoading(true)
    try {
      const { fileApi } = await import('../api/file')
      const newPath = currentPath === '/' ? `/${newDirName}` : `${currentPath}/${newDirName}`
      await fileApi.createDirectory({
        hostId,
        path: newPath,
        username: credentials.username,
        password: credentials.password,
        key: credentials.key,
      })
      message.success(`Created directory: ${newDirName}`)
      setMkdirModalVisible(false)
      setNewDirName('')
      loadDirectory(currentPath)
    } catch (error: any) {
      message.error(`Failed to create directory: ${error.message || 'Unknown error'}`)
    } finally {
      setLoading(false)
    }
  }

  // Handle auth submit
  const handleAuthSubmit = (creds: { username: string; password: string; key: string }) => {
    setCredentials(creds)
    setAuthModalVisible(false)

    if (pendingAction === 'upload') {
      // Trigger file upload
      const input = document.createElement('input') as HTMLInputElement
      input.type = 'file'
      input.multiple = true
      input.onchange = async (e) => {
        const target = e.target as HTMLInputElement
        if (target.files) {
          for (let i = 0; i < target.files.length; i++) {
            await uploadFile(target.files[i])
          }
        }
      }
      input.click()
    } else if (pendingAction === 'mkdir') {
      setMkdirModalVisible(true)
    } else {
      loadDirectory(currentPath)
    }

    setPendingAction(null)
  }

  // Upload a single file
  const uploadFile = async (file: File) => {
    setLoading(true)
    try {
      const { fileApi } = await import('../api/file')
      await fileApi.uploadFile({
        hostId,
        remotePath: currentPath,
        username: credentials.username,
        password: credentials.password,
        key: credentials.key,
        file,
      })
      message.success(`Uploaded: ${file.name}`)
      loadDirectory(currentPath)
    } catch (error: any) {
      message.error(`Failed to upload: ${error.message || 'Unknown error'}`)
    } finally {
      setLoading(false)
    }
  }

  // Format file size
  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // Format date
  const formatDate = (dateStr: string): string => {
    const date = new Date(dateStr)
    return date.toLocaleString()
  }

  // Breadcrumb items
  const pathParts = currentPath.split('/').filter(Boolean)
  const breadcrumbItems = [
    {
      title: <span onClick={() => handleBreadcrumbNavigate('/')}><HomeOutlined /></span>,
    },
  ]

  let accumPath = ''
  for (let i = 0; i < pathParts.length; i++) {
    accumPath += '/' + pathParts[i]
    breadcrumbItems.push({
      title: <span onClick={() => handleBreadcrumbNavigate(accumPath)}>{pathParts[i]}</span>,
    })
  }

  // Table columns
  const columns: ColumnsType<FileInfo> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name, record) => (
        <Space>
          {record.isDir ? (
            <FolderOutlined style={{ color: '#1890ff', fontSize: '18px' }} />
          ) : (
            <FileOutlined style={{ color: '#8c8c8c', fontSize: '18px' }} />
          )}
          <a
            onClick={(e) => {
              e.preventDefault()
              handleItemClick(record)
            }}
            style={{ color: record.isDir ? '#1890ff' : 'inherit' }}
          >
            {name}
          </a>
        </Space>
      ),
    },
    {
      title: 'Size',
      dataIndex: 'size',
      key: 'size',
      width: 120,
      render: (size, record) => (record.isDir ? '-' : formatFileSize(size)),
    },
    {
      title: 'Modified',
      dataIndex: 'modTime',
      key: 'modTime',
      width: 180,
      render: (date) => formatDate(date),
    },
    {
      title: 'Permissions',
      dataIndex: 'permissions',
      key: 'permissions',
      width: 100,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 150,
      render: (_, record) => (
        <Space>
          {!record.isDir && (
            <Tooltip title="Download">
              <Button
                type="text"
                icon={<DownloadOutlined />}
                onClick={() => handleDownload(record)}
                disabled={loading}
              />
            </Tooltip>
          )}
          <Tooltip title="Delete">
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              onClick={() => {
                setSelectedRowKeys([record.path])
              }}
              disabled={loading}
            />
          </Tooltip>
        </Space>
      ),
    },
  ]

  // Row selection
  const rowSelection = {
    selectedRowKeys,
    onChange: (newSelectedRowKeys: React.Key[]) => {
      setSelectedRowKeys(newSelectedRowKeys)
    },
  }

  return (
    <div>
      {/* Breadcrumb and toolbar */}
      <div style={{ marginBottom: '16px' }}>
        <Breadcrumb items={breadcrumbItems} style={{ marginBottom: '8px' }} />
        <Space>
          <Button
            icon={<ArrowUpOutlined />}
            onClick={handleParent}
            disabled={currentPath === '/' || loading}
          >
            Parent
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh} loading={loading}>
            Refresh
          </Button>
          <Button icon={<UploadOutlined />} onClick={handleUpload} loading={loading}>
            Upload
          </Button>
          <Button icon={<FolderAddOutlined />} onClick={handleMkdir} loading={loading}>
            New Folder
          </Button>
          {selectedRowKeys.length > 0 && (
            <Popconfirm
              title="Delete selected items?"
              description={`Are you sure you want to delete ${selectedRowKeys.length} item(s)?`}
              onConfirm={handleDelete}
            >
              <Button danger icon={<DeleteOutlined />} loading={loading}>
                Delete ({selectedRowKeys.length})
              </Button>
            </Popconfirm>
          )}
        </Space>
      </div>

      {/* File table */}
      <Table
        columns={columns}
        dataSource={files}
        rowKey="path"
        loading={loading}
        rowSelection={rowSelection}
        pagination={false}
        size="small"
        onRow={(record) => ({
          onDoubleClick: () => handleItemClick(record),
        })}
      />

      {/* Auth modal */}
      <FileAuthModal
        visible={authModalVisible}
        hostName={hostName}
        onSubmit={handleAuthSubmit}
        onCancel={() => {
          setAuthModalVisible(false)
          setPendingAction(null)
        }}
      />

      {/* New folder modal */}
      <Modal
        title="New Folder"
        open={mkdirModalVisible}
        onOk={handleCreateDirectory}
        onCancel={() => {
          setMkdirModalVisible(false)
          setNewDirName('')
        }}
        okText="Create"
        confirmLoading={loading}
      >
        <Input
          placeholder="Folder name"
          value={newDirName}
          onChange={(e) => setNewDirName(e.target.value)}
          onPressEnter={handleCreateDirectory}
        />
      </Modal>
    </div>
  )
}

export default FileBrowser
