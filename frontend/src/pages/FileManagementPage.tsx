import React from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Card, Button, Space, Alert, Spin } from 'antd'
import { ArrowLeftOutlined, FolderOutlined } from '@ant-design/icons'
import { FileBrowser } from '../components/FileBrowser'
import { hostApi } from '../api/host'

const FileManagementPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  // Fetch host details
  const { data: host, isLoading, error } = useQuery({
    queryKey: ['host', id],
    queryFn: () => hostApi.getHost(id!),
    enabled: !!id,
  })

  // Loading state
  if (isLoading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center', height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin size="large" tip="Loading host details..." />
      </div>
    )
  }

  // Error state
  if (error || !host) {
    return (
      <div style={{ padding: '24px' }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/hosts')}>
          Back to Hosts
        </Button>
        <Alert
          style={{ marginTop: '16px' }}
          message="Failed to load host"
          description={(error as Error)?.message || 'Host not found'}
          type="error"
          showIcon
        />
      </div>
    )
  }

  // Check if host is approved/online
  const isAvailable = host.status === 'approved' || host.status === 'online'

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/hosts/${id}`)}>
            Back to Host
          </Button>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
            <FolderOutlined /> File Manager: {host.hostname || host.ipAddress}
          </span>
        </Space>
      </div>

      {/* Info alert for file operations */}
      <Alert
        style={{ marginBottom: '16px' }}
        message={
          !isAvailable
            ? 'Host is not available for file operations. Please approve the host and ensure the agent is running.'
            : 'File operations are performed via SFTP. Make sure you have the proper credentials.'
        }
        type={isAvailable ? 'info' : 'warning'}
        showIcon
      />

      {/* File browser */}
      <Card
        bodyStyle={{ padding: '16px' }}
        style={{ minHeight: 'calc(100vh - 200px)' }}
      >
        {isAvailable ? (
          <FileBrowser
            hostId={id!}
            hostName={host.hostname || host.ipAddress}
          />
        ) : (
          <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>
            File management is only available for approved and online hosts.
          </div>
        )}
      </Card>
    </div>
  )
}

export default FileManagementPage
