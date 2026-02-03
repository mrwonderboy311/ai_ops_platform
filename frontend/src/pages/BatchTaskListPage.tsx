import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Table, Button, Space, Tag, message, Select, Popconfirm, Progress, Card, Statistic, Row, Col } from 'antd'
import {
  PlusOutlined,
  EyeOutlined,
  DeleteOutlined,
  ReloadOutlined,
  StopOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { batchTaskApi } from '../api/batchTask'
import type { BatchTask, BatchTaskStatus, BatchTaskType } from '../types/batchTask'
import { CreateBatchTaskModal } from '../components/CreateBatchTaskModal'

const { Option } = Select

export const BatchTaskListPage: React.FC = () => {
  const navigate = useNavigate()
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState<string | undefined>()
  const [typeFilter, setTypeFilter] = useState<string | undefined>()

  // Fetch batch tasks
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['batchTasks', page, pageSize, statusFilter, typeFilter],
    queryFn: () =>
      batchTaskApi.listBatchTasks({
        page,
        pageSize,
        status: statusFilter,
        type: typeFilter,
      }),
  })

  // Handle delete
  const handleDelete = async (taskId: string) => {
    try {
      await batchTaskApi.deleteBatchTask(taskId)
      message.success('Batch task deleted successfully')
      refetch()
    } catch (error: any) {
      message.error(`Failed to delete: ${error.response?.data?.message || error.message}`)
    }
  }

  // Handle cancel
  const handleCancel = async (taskId: string) => {
    try {
      await batchTaskApi.cancelBatchTask(taskId)
      message.success('Batch task cancelled successfully')
      refetch()
    } catch (error: any) {
      message.error(`Failed to cancel: ${error.response?.data?.message || error.message}`)
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
      parallel: 'Parallel',
      serial: 'Serial',
      rolling: 'Rolling',
    }
    return labels[strategy] || strategy
  }

  // Calculate statistics
  const stats = React.useMemo(() => {
    if (!data?.tasks) return { total: 0, running: 0, completed: 0, failed: 0 }
    return {
      total: data.total,
      running: data.tasks.filter((t) => t.status === 'running').length,
      completed: data.tasks.filter((t) => t.status === 'completed').length,
      failed: data.tasks.filter((t) => t.status === 'failed').length,
    }
  }, [data])

  // Table columns
  const columns: ColumnsType<BatchTask> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name, record) => (
        <a onClick={() => navigate(`/batch-tasks/${record.id}`)}>{name}</a>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: BatchTaskType) => (
        <Tag color="blue">{type.toUpperCase()}</Tag>
      ),
    },
    {
      title: 'Strategy',
      dataIndex: 'strategy',
      key: 'strategy',
      width: 100,
      render: (strategy: string) => (
        <Tag>{getStrategyLabel(strategy)}</Tag>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: BatchTaskStatus) => (
        <Tag color={getStatusColor(status)}>{status.toUpperCase()}</Tag>
      ),
    },
    {
      title: 'Progress',
      key: 'progress',
      width: 150,
      render: (_, record) => {
        const percent = record.totalHosts > 0
          ? Math.round((record.completedHosts / record.totalHosts) * 100)
          : 0
        return (
          <Progress
            percent={percent}
            size="small"
            status={record.failedHosts > 0 ? 'exception' : undefined}
          />
        )
      },
    },
    {
      title: 'Hosts',
      key: 'hosts',
      width: 120,
      render: (_, record) => (
        <span>
          {record.completedHosts}/{record.totalHosts}
        </span>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/batch-tasks/${record.id}`)}
          >
            View
          </Button>
          {record.status === 'running' && (
            <Popconfirm
              title="Cancel this task?"
              onConfirm={() => handleCancel(record.id)}
            >
              <Button
                type="link"
                size="small"
                danger
                icon={<StopOutlined />}
              >
                Cancel
              </Button>
            </Popconfirm>
          )}
          {record.status !== 'running' && (
            <Popconfirm
              title="Delete this task?"
              onConfirm={() => handleDelete(record.id)}
            >
              <Button
                type="link"
                size="small"
                danger
                icon={<DeleteOutlined />}
              >
                Delete
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0 }}>Batch Tasks</h1>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalVisible(true)}>
          Create Batch Task
        </Button>
      </div>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic title="Total Tasks" value={stats.total} prefix={<ThunderboltOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Running"
              value={stats.running}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Completed"
              value={stats.completed}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Failed"
              value={stats.failed}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Filters */}
      <Space style={{ marginBottom: '16px' }}>
        <Select
          placeholder="Filter by status"
          allowClear
          style={{ width: 150 }}
          onChange={(value) => {
            setStatusFilter(value)
            setPage(1)
          }}
          value={statusFilter}
        >
          <Option value="pending">Pending</Option>
          <Option value="running">Running</Option>
          <Option value="completed">Completed</Option>
          <Option value="failed">Failed</Option>
          <Option value="cancelled">Cancelled</Option>
        </Select>
        <Select
          placeholder="Filter by type"
          allowClear
          style={{ width: 150 }}
          onChange={(value) => {
            setTypeFilter(value)
            setPage(1)
          }}
          value={typeFilter}
        >
          <Option value="command">Command</Option>
          <Option value="script">Script</Option>
          <Option value="file_op">File Operation</Option>
        </Select>
        <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
          Refresh
        </Button>
      </Space>

      {/* Table */}
      <Table
        columns={columns}
        dataSource={data?.tasks || []}
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: page,
          pageSize: pageSize,
          total: data?.total || 0,
          showSizeChanger: true,
          showTotal: (total) => `Total ${total} tasks`,
          onChange: (newPage, newPageSize) => {
            setPage(newPage)
            setPageSize(newPageSize || 20)
          },
        }}
      />

      {/* Create Modal */}
      <CreateBatchTaskModal
        visible={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onSuccess={() => {
          setCreateModalVisible(false)
          refetch()
        }}
      />
    </div>
  )
}

export default BatchTaskListPage
