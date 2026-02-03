import React, { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button, Space, Table, Tag, Card, Row, Col, Statistic, Form, Input, Select, DatePicker, Modal } from 'antd'
import {
  FileTextOutlined,
  ReloadOutlined,
  SearchOutlined,
  UserOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  EyeOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { auditApi } from '../api/audit'
import type { AuditLog } from '../types/audit'

const { RangePicker } = DatePicker

export const AuditLogPage: React.FC = () => {
  const [form] = Form.useForm()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [filters, setFilters] = useState<any>({})
  const [detailModal, setDetailModal] = useState<{ visible: boolean; log?: AuditLog }>({
    visible: false,
  })

  // Fetch audit logs
  const { data: logsData, isLoading, refetch } = useQuery({
    queryKey: ['auditLogs', page, pageSize, filters],
    queryFn: () =>
      auditApi.getAuditLogs({
        page,
        pageSize,
        ...filters,
      }),
  })

  // Fetch audit log summary
  const { data: summaryData, refetch: refetchSummary } = useQuery({
    queryKey: ['auditLogSummary'],
    queryFn: () => auditApi.getAuditLogSummary('24h'),
  })

  const summary = summaryData?.data

  // Handle search
  const handleSearch = () => {
    const values = form.getFieldsValue()
    const newFilters: any = {}

    if (values.username) newFilters.username = values.username
    if (values.action) newFilters.action = values.action
    if (values.resource) newFilters.resource = values.resource
    if (values.resourceId) newFilters.resourceId = values.resourceId
    if (values.dateRange) {
      newFilters.startTime = values.dateRange[0].toISOString()
      newFilters.endTime = values.dateRange[1].toISOString()
    }

    setFilters(newFilters)
    setPage(1)
  }

  // Handle reset
  const handleReset = () => {
    form.resetFields()
    setFilters({})
    setPage(1)
  }

  // Handle refresh
  const handleRefresh = () => {
    refetch()
    refetchSummary()
  }

  // Get status config
  const getStatusConfig = (statusCode: number) => {
    if (statusCode >= 200 && statusCode < 300) {
      return { color: 'success', icon: <CheckCircleOutlined />, label: 'Success' }
    } else if (statusCode >= 400 && statusCode < 500) {
      return { color: 'warning', icon: <CloseCircleOutlined />, label: 'Client Error' }
    } else if (statusCode >= 500) {
      return { color: 'error', icon: <CloseCircleOutlined />, label: 'Server Error' }
    }
    return { color: 'default', icon: null, label: statusCode.toString() }
  }

  // Get action config
  const getActionConfig = (action: string) => {
    switch (action) {
      case 'create':
        return { color: 'green', label: 'Create' }
      case 'read':
        return { color: 'blue', label: 'Read' }
      case 'update':
        return { color: 'orange', label: 'Update' }
      case 'delete':
        return { color: 'red', label: 'Delete' }
      case 'login':
        return { color: 'cyan', label: 'Login' }
      case 'logout':
        return { color: 'default', label: 'Logout' }
      default:
        return { color: 'default', label: action }
    }
  }

  // Table columns
  const columns: ColumnsType<AuditLog> = [
    {
      title: 'User',
      dataIndex: 'username',
      key: 'username',
      width: 120,
      render: (username: string) => (
        <Space>
          <UserOutlined />
          {username || 'System'}
        </Space>
      ),
    },
    {
      title: 'Action',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (action: string) => {
        const config = getActionConfig(action)
        return <Tag color={config.color}>{config.label}</Tag>
      },
    },
    {
      title: 'Resource',
      dataIndex: 'resource',
      key: 'resource',
      width: 120,
      render: (resource: string) => <Tag>{resource}</Tag>,
    },
    {
      title: 'Method',
      dataIndex: 'method',
      key: 'method',
      width: 80,
    },
    {
      title: 'Path',
      dataIndex: 'path',
      key: 'path',
      ellipsis: true,
    },
    {
      title: 'Status',
      dataIndex: 'statusCode',
      key: 'statusCode',
      width: 100,
      render: (statusCode: number) => {
        const config = getStatusConfig(statusCode)
        return (
          <Tag color={config.color} icon={config.icon}>
            {statusCode}
          </Tag>
        )
      },
    },
    {
      title: 'IP Address',
      dataIndex: 'ipAddress',
      key: 'ipAddress',
      width: 140,
      ellipsis: true,
    },
    {
      title: 'Time',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Button
          size="small"
          icon={<EyeOutlined />}
          onClick={() => setDetailModal({ visible: true, log: record })}
        >
          View
        </Button>
      ),
    },
  ]

  const logs = logsData?.data.logs || []

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
            <FileTextOutlined /> Audit Logs
          </span>
        </Space>
        <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
          Refresh
        </Button>
      </div>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="Total Operations"
              value={summary?.totalOperations || 0}
              prefix={<FileTextOutlined />}
              valueStyle={{ color: '#888' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Active Users"
              value={summary?.userActivity || 0}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Failed Operations"
              value={summary?.failedOperations || 0}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Success Rate"
              value={
                summary && summary.totalOperations > 0
                  ? (((summary.totalOperations - summary.failedOperations) / summary.totalOperations) * 100).toFixed(1)
                  : '0'
              }
              suffix="%"
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Filters */}
      <Card style={{ marginBottom: '16px' }}>
        <Form form={form} layout="inline">
          <Form.Item name="username" label="Username">
            <Input placeholder="Search by username" allowClear style={{ width: 150 }} />
          </Form.Item>
          <Form.Item name="action" label="Action">
            <Select placeholder="Select action" allowClear style={{ width: 120 }}>
              <Select.Option value="create">Create</Select.Option>
              <Select.Option value="read">Read</Select.Option>
              <Select.Option value="update">Update</Select.Option>
              <Select.Option value="delete">Delete</Select.Option>
              <Select.Option value="login">Login</Select.Option>
              <Select.Option value="logout">Logout</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="resource" label="Resource">
            <Select placeholder="Select resource" allowClear style={{ width: 120 }}>
              <Select.Option value="hosts">Hosts</Select.Option>
              <Select.Option value="clusters">Clusters</Select.Option>
              <Select.Option value="users">Users</Select.Option>
              <Select.Option value="batch-tasks">Batch Tasks</Select.Option>
              <Select.Option value="alert-rules">Alert Rules</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="dateRange" label="Date Range">
            <RangePicker showTime />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
                Search
              </Button>
              <Button onClick={handleReset}>Reset</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {/* Table */}
      <Table
        columns={columns}
        dataSource={logs}
        rowKey="id"
        loading={isLoading}
        pagination={{
          total: logsData?.data.total || 0,
          pageSize: pageSize,
          current: page,
          showSizeChanger: true,
          showTotal: (total) => `Total ${total} items`,
          onChange: (page, pageSize) => {
            setPage(page)
            setPageSize(pageSize)
          },
        }}
        size="small"
      />

      {/* Detail Modal */}
      <Modal
        title="Audit Log Details"
        open={detailModal.visible}
        onCancel={() => setDetailModal({ visible: false })}
        footer={[
          <Button key="close" onClick={() => setDetailModal({ visible: false })}>
            Close
          </Button>,
        ]}
        width={700}
      >
        {detailModal.log && (
          <div>
            <Row gutter={16}>
              <Col span={12}>
                <div style={{ marginBottom: '8px' }}>
                  <strong>User:</strong> {detailModal.log.username || 'System'}
                </div>
              </Col>
              <Col span={12}>
                <div style={{ marginBottom: '8px' }}>
                  <strong>Action:</strong> {detailModal.log.action}
                </div>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <div style={{ marginBottom: '8px' }}>
                  <strong>Resource:</strong> {detailModal.log.resource}
                </div>
              </Col>
              <Col span={12}>
                <div style={{ marginBottom: '8px' }}>
                  <strong>Resource ID:</strong> {detailModal.log.resourceId || '-'}
                </div>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <div style={{ marginBottom: '8px' }}>
                  <strong>Method:</strong> {detailModal.log.method}
                </div>
              </Col>
              <Col span={12}>
                <div style={{ marginBottom: '8px' }}>
                  <strong>Status Code:</strong> {detailModal.log.statusCode}
                </div>
              </Col>
            </Row>
            <div style={{ marginBottom: '8px' }}>
              <strong>Path:</strong> {detailModal.log.path}
            </div>
            <div style={{ marginBottom: '8px' }}>
              <strong>IP Address:</strong> {detailModal.log.ipAddress}
            </div>
            <div style={{ marginBottom: '8px' }}>
              <strong>User Agent:</strong> {detailModal.log.userAgent}
            </div>
            <div style={{ marginBottom: '8px' }}>
              <strong>Time:</strong> {dayjs(detailModal.log.createdAt).format('YYYY-MM-DD HH:mm:ss')}
            </div>
            {detailModal.log.errorMsg && (
              <div style={{ marginBottom: '8px' }}>
                <strong>Error:</strong> <span style={{ color: '#cf1322' }}>{detailModal.log.errorMsg}</span>
              </div>
            )}
            {detailModal.log.newValue && (
              <div style={{ marginBottom: '8px' }}>
                <strong>New Value:</strong>
                <pre style={{ background: '#f5f5f5', padding: '8px', borderRadius: '4px', maxHeight: '200px', overflow: 'auto' }}>
                  {JSON.stringify(JSON.parse(detailModal.log.newValue), null, 2)}
                </pre>
              </div>
            )}
          </div>
        )}
      </Modal>
    </div>
  )
}

export default AuditLogPage
