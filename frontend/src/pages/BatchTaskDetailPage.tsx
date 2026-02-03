import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Button, Space, Tag, message, Card, Descriptions, Progress, Table, Tabs, Alert, Spin, Modal } from 'antd'
import { ArrowLeftOutlined, ReloadOutlined, StopOutlined, EyeOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { batchTaskApi } from '../api/batchTask'
import type { BatchTaskHost, BatchTaskStatus } from '../types/batchTask'

export const BatchTaskDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [autoRefresh, setAutoRefresh] = useState(true)

  // Fetch batch task details
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['batchTask', id],
    queryFn: () => batchTaskApi.getBatchTask(id!),
    enabled: !!id,
    refetchInterval: autoRefresh ? 3000 : false, // Auto-refresh every 3 seconds when running
  })

  // Auto-refresh control
  const isRunning = data?.batchTask?.status === 'running' || data?.batchTask?.status === 'pending'
  useEffect(() => {
    setAutoRefresh(isRunning)
  }, [isRunning])

  // Handle cancel
  const handleCancel = async () => {
    try {
      await batchTaskApi.cancelBatchTask(id!)
      message.success('Batch task cancelled successfully')
      refetch()
    } catch (error: any) {
      message.error(`Failed to cancel: ${error.response?.data?.message || error.message}`)
    }
  }

  // Handle retry
  const handleRetry = async () => {
    try {
      await batchTaskApi.executeBatchTask({ taskId: id! })
      message.success('Batch task re-executed successfully')
      refetch()
    } catch (error: any) {
      message.error(`Failed to retry: ${error.response?.data?.message || error.message}`)
    }
  }

  // Get status color
  const getStatusColor = (status: BatchTaskStatus): string => {
    const colors: Record<BatchTaskStatus, string> = {
      pending: 'default',
      running: 'processing',
      completed: 'success',
      failed: 'error',
      cancelled: 'warning',
    }
    return colors[status] || 'default'
  }

  // Get strategy label
  const getStrategyLabel = (strategy: string): string => {
    const labels: Record<string, string> = {
      parallel: 'Parallel (All hosts at once)',
      serial: 'Serial (One at a time)',
      rolling: 'Rolling (In batches)',
    }
    return labels[strategy] || strategy
  }

  // Format duration
  const formatDuration = (ms: number): string => {
    if (ms < 1000) return `${ms}ms`
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
    return `${(ms / 60000).toFixed(1)}m`
  }

  // Host table columns
  const hostColumns: ColumnsType<BatchTaskHost> = [
    {
      title: 'Host',
      key: 'host',
      render: (_, record) => record.host?.hostname || record.host?.ipAddress || record.hostId,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: BatchTaskStatus) => (
        <Tag color={getStatusColor(status)}>{status.toUpperCase()}</Tag>
      ),
    },
    {
      title: 'Exit Code',
      dataIndex: 'exitCode',
      key: 'exitCode',
      render: (code?: number) => code ?? '-',
    },
    {
      title: 'Duration',
      dataIndex: 'duration',
      key: 'duration',
      render: (duration: number) => formatDuration(duration),
    },
    {
      title: 'Started',
      dataIndex: 'startedAt',
      key: 'startedAt',
      render: (date?: string) => date ? new Date(date).toLocaleString() : '-',
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_, record) => (
        <Space>
          {(record.stdout || record.stderr) && (
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => {
                Modal.info({
                  title: `Output for ${record.host?.hostname || record.hostId}`,
                  width: 800,
                  content: (
                    <div>
                      {record.stdout && (
                        <div>
                          <h4>Stdout:</h4>
                          <pre style={{ background: '#f5f5f5', padding: '8px', maxHeight: '300px', overflow: 'auto' }}>
                            {record.stdout}
                          </pre>
                        </div>
                      )}
                      {record.stderr && (
                        <div>
                          <h4>Stderr:</h4>
                          <pre style={{ background: '#fff1f0', padding: '8px', maxHeight: '300px', overflow: 'auto' }}>
                            {record.stderr}
                          </pre>
                        </div>
                      )}
                      {record.errorMessage && (
                        <div>
                          <h4>Error:</h4>
                          <p style={{ color: '#f5222d' }}>{record.errorMessage}</p>
                        </div>
                      )}
                    </div>
                  ),
                })
              }}
            >
              View Output
            </Button>
          )}
        </Space>
      ),
    },
  ]

  if (isLoading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center', height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin size="large" tip="Loading batch task details..." />
      </div>
    )
  }

  if (error || !data) {
    return (
      <div style={{ padding: '24px' }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/batch-tasks')}>
          Back to Batch Tasks
        </Button>
        <Alert
          style={{ marginTop: '16px' }}
          message="Failed to load batch task"
          description={(error as Error)?.message || 'Batch task not found'}
          type="error"
          showIcon
        />
      </div>
    )
  }

  const { batchTask, hosts, progress } = data

  const tabItems = [
    {
      key: 'overview',
      label: 'Overview',
      children: (
        <Card title="Task Details" styles={{ body: { padding: '24px' } }}>
          <Descriptions column={2} bordered>
            <Descriptions.Item label="Task Name" span={2}>
              {batchTask.name}
            </Descriptions.Item>
            <Descriptions.Item label="Description" span={2}>
              {batchTask.description || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="Type">
              <Tag color="blue">{batchTask.type.toUpperCase()}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Strategy">
              {getStrategyLabel(batchTask.strategy)}
            </Descriptions.Item>
            <Descriptions.Item label="Status">
              <Tag color={getStatusColor(batchTask.status)}>{batchTask.status.toUpperCase()}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Progress">
              <Progress percent={Math.round(progress)} size="small" />
            </Descriptions.Item>
            <Descriptions.Item label="Hosts">
              {batchTask.completedHosts} / {batchTask.totalHosts}
              {batchTask.failedHosts > 0 && (
                <Tag color="error" style={{ marginLeft: '8px' }}>
                  {batchTask.failedHosts} failed
                </Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="Timeout">
              {batchTask.timeout}s
            </Descriptions.Item>
            <Descriptions.Item label="Created">
              {new Date(batchTask.createdAt).toLocaleString()}
            </Descriptions.Item>
            <Descriptions.Item label="Started">
              {batchTask.startedAt ? new Date(batchTask.startedAt).toLocaleString() : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="Completed">
              {batchTask.completedAt ? new Date(batchTask.completedAt).toLocaleString() : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="Command" span={2}>
              <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                {batchTask.command || batchTask.script || '-'}
              </pre>
            </Descriptions.Item>
          </Descriptions>
        </Card>
      ),
    },
    {
      key: 'hosts',
      label: `Host Results (${hosts.length})`,
      children: (
        <Table
          columns={hostColumns}
          dataSource={hosts}
          rowKey="id"
          pagination={false}
          size="small"
          expandable={{
            expandedRowRender: (record) => (
              <div style={{ padding: '16px' }}>
                {record.stdout && (
                  <div>
                    <strong>Stdout:</strong>
                    <pre style={{ background: '#f5f5f5', padding: '8px', marginTop: '8px' }}>
                      {record.stdout}
                    </pre>
                  </div>
                )}
                {record.stderr && (
                  <div>
                    <strong>Stderr:</strong>
                    <pre style={{ background: '#fff1f0', padding: '8px', marginTop: '8px' }}>
                      {record.stderr}
                    </pre>
                  </div>
                )}
                {record.errorMessage && (
                  <div>
                    <strong>Error:</strong>
                    <p style={{ color: '#f5222d', marginTop: '8px' }}>{record.errorMessage}</p>
                  </div>
                )}
              </div>
            ),
          }}
        />
      ),
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/batch-tasks')}>
            Back to Batch Tasks
          </Button>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
            {batchTask.name}
          </span>
          <Tag color={getStatusColor(batchTask.status)}>{batchTask.status.toUpperCase()}</Tag>
        </Space>
        <Space>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => refetch()}
          >
            Refresh
          </Button>
          {(batchTask.status === 'running' || batchTask.status === 'pending') && (
            <Button danger icon={<StopOutlined />} onClick={handleCancel}>
              Cancel
            </Button>
          )}
          {(batchTask.status === 'failed' || batchTask.status === 'cancelled') && (
            <Button type="primary" onClick={handleRetry}>
              Retry Failed Hosts
            </Button>
          )}
        </Space>
      </div>

      {/* Auto-refresh indicator */}
      {autoRefresh && (
        <Alert
          style={{ marginBottom: '16px' }}
          message="Auto-refreshing"
          description="Task status is being updated automatically. Click 'Refresh' to manually update."
          type="info"
          showIcon
          closable
          onClose={() => setAutoRefresh(false)}
        />
      )}

      {/* Tabs */}
      <Tabs items={tabItems} />
    </div>
  )
}

export default BatchTaskDetailPage
