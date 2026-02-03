import React, { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Button,
  Table,
  Space,
  Modal,
  Form,
  Input,
  Select,
  message,
  Popconfirm,
  Tag,
  Card,
  Tooltip,
  Switch,
  Empty,
  InputNumber,
  Row,
  Col,
  Statistic,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  SyncOutlined,
  DashboardOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  LoadingOutlined,
  ApiOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import grafanaApi, { type GrafanaInstance, type CreateGrafanaInstanceRequest } from '../api/grafana'

const { Option } = Select

export const GrafanaInstancesPage: React.FC = () => {
  const queryClient = useQueryClient()

  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingInstance, setEditingInstance] = useState<GrafanaInstance | null>(null)
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<{ success: boolean; message: string; version?: string } | null>(null)
  const [syncing, setSyncing] = useState<Record<string, boolean>>({})
  const [form] = Form.useForm()

  // Fetch instances
  const { data: instancesData, isLoading, refetch } = useQuery({
    queryKey: ['grafanaInstances'],
    queryFn: () => grafanaApi.getInstances({ page: 1, pageSize: 100 }),
  })

  const instances = instancesData?.data || []

  // Create mutation
  const createMutation = useMutation({
    mutationFn: grafanaApi.createInstance,
    onSuccess: () => {
      message.success('Grafana instance created successfully')
      setIsModalOpen(false)
      form.resetFields()
      setTestResult(null)
      queryClient.invalidateQueries({ queryKey: ['grafanaInstances'] })
    },
    onError: (error: any) => {
      message.error(`Failed to create Grafana instance: ${error.response?.data?.message || error.message}`)
    },
  })

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      grafanaApi.updateInstance(id, data),
    onSuccess: () => {
      message.success('Grafana instance updated successfully')
      setIsModalOpen(false)
      setEditingInstance(null)
      form.resetFields()
      setTestResult(null)
      queryClient.invalidateQueries({ queryKey: ['grafanaInstances'] })
    },
    onError: (error: any) => {
      message.error(`Failed to update Grafana instance: ${error.response?.data?.message || error.message}`)
    },
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: grafanaApi.deleteInstance,
    onSuccess: () => {
      message.success('Grafana instance deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['grafanaInstances'] })
    },
    onError: (error: any) => {
      message.error(`Failed to delete Grafana instance: ${error.response?.data?.message || error.message}`)
    },
  })

  // Handle add
  const handleAdd = () => {
    setEditingInstance(null)
    setTestResult(null)
    form.resetFields()
    form.setFieldsValue({
      autoSync: true,
      syncInterval: 300,
    })
    setIsModalOpen(true)
  }

  // Handle edit
  const handleEdit = (instance: GrafanaInstance) => {
    setEditingInstance(instance)
    setTestResult(null)
    form.setFieldsValue({
      clusterId: instance.clusterId,
      name: instance.name,
      url: instance.url,
      username: instance.username,
      autoSync: instance.autoSync,
      syncInterval: instance.syncInterval,
      status: instance.status,
    })
    setIsModalOpen(true)
  }

  // Handle delete
  const handleDelete = async (id: string) => {
    deleteMutation.mutate(id)
  }

  // Handle test connection
  const handleTest = async () => {
    try {
      setTesting(true)
      setTestResult(null)
      const values = await form.validateFields(['url', 'apiKey', 'username', 'password'])
      const result = await grafanaApi.testInstance({
        url: values.url,
        apiKey: values.apiKey,
        username: values.username,
        password: values.password,
      })
      setTestResult({
        success: result.success,
        message: result.success ? `Connected successfully (${result.version})` : result.message,
        version: result.version,
      })
      if (result.success) {
        message.success('Connection test successful')
      } else {
        message.error(result.message)
      }
    } catch (error: any) {
      setTestResult({
        success: false,
        message: error.response?.data?.message || error.message || 'Connection test failed',
      })
      message.error('Connection test failed')
    } finally {
      setTesting(false)
    }
  }

  // Handle sync
  const handleSync = async (instance: GrafanaInstance) => {
    try {
      setSyncing({ ...syncing, [instance.id]: true })
      const result = await grafanaApi.syncInstance(instance.id, {
        syncDashboards: true,
        syncDataSources: true,
        syncFolders: true,
      })
      message.success(result.message)
      queryClient.invalidateQueries({ queryKey: ['grafanaInstances'] })
    } catch (error: any) {
      message.error(`Sync failed: ${error.response?.data?.message || error.message}`)
    } finally {
      setSyncing({ ...syncing, [instance.id]: false })
    }
  }

  // Get status tag
  const getStatusTag = (status: string, syncStatus: string) => {
    if (syncStatus === 'running') {
      return (
        <Tag color="processing" icon={<LoadingOutlined />}>
          Syncing
        </Tag>
      )
    }
    const statusMap: Record<string, { color: string; text: string; icon: React.ReactNode }> = {
      active: { color: 'success', text: 'Active', icon: <CheckCircleOutlined /> },
      inactive: { color: 'default', text: 'Inactive', icon: <CloseCircleOutlined /> },
      error: { color: 'error', text: 'Error', icon: <CloseCircleOutlined /> },
    }
    const { color, text, icon } = statusMap[status] || { color: 'default', text: status, icon: null }
    return (
      <Tag color={color} icon={icon}>
        {text}
      </Tag>
    )
  }

  // Calculate statistics
  const totalDashboards = instances.reduce((sum, inst) => sum + inst.dashboardCount, 0)
  const activeCount = instances.filter(inst => inst.status === 'active').length
  const autoSyncCount = instances.filter(inst => inst.autoSync).length

  // Table columns
  const columns: ColumnsType<GrafanaInstance> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <DashboardOutlined />
          <span style={{ fontWeight: 'bold' }}>{name}</span>
        </Space>
      ),
    },
    {
      title: 'URL',
      dataIndex: 'url',
      key: 'url',
      ellipsis: true,
      render: (url: string) => (
        <a href={url} target="_blank" rel="noopener noreferrer" style={{ fontFamily: 'monospace', fontSize: '12px' }}>
          {url}
        </a>
      ),
    },
    {
      title: 'Cluster',
      key: 'cluster',
      render: (_: any, record: GrafanaInstance) => record.cluster?.name || '-',
    },
    {
      title: 'Status',
      key: 'status',
      width: 120,
      render: (_: any, record: GrafanaInstance) => getStatusTag(record.status, record.syncStatus),
    },
    {
      title: 'Auto Sync',
      dataIndex: 'autoSync',
      key: 'autoSync',
      width: 100,
      render: (autoSync: boolean) => (
        <Tag color={autoSync ? 'blue' : 'default'}>
          {autoSync ? 'On' : 'Off'}
        </Tag>
      ),
    },
    {
      title: 'Dashboards',
      dataIndex: 'dashboardCount',
      key: 'dashboardCount',
      width: 100,
      render: (count: number) => (
        <Tag icon={<DashboardOutlined />} color="purple">
          {count}
        </Tag>
      ),
    },
    {
      title: 'Data Sources',
      dataIndex: 'dataSourceCount',
      key: 'dataSourceCount',
      width: 120,
      render: (count: number) => (
        <Tag icon={<ApiOutlined />} color="cyan">
          {count}
        </Tag>
      ),
    },
    {
      title: 'Last Sync',
      dataIndex: 'lastSyncAt',
      key: 'lastSyncAt',
      width: 150,
      render: (date: string) => (date ? new Date(date).toLocaleString() : '-'),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (_: any, record: GrafanaInstance) => (
        <Space size="small">
          <Tooltip title="Sync Now">
            <Button
              size="small"
              icon={<SyncOutlined />}
              onClick={() => handleSync(record)}
              loading={syncing[record.id]}
            />
          </Tooltip>
          <Tooltip title="Edit">
            <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          </Tooltip>
          <Popconfirm
            title="Delete Instance"
            description="Are you sure you want to delete this Grafana instance?"
            onConfirm={() => handleDelete(record.id)}
            okText="Yes"
            cancelText="No"
          >
            <Tooltip title="Delete">
              <Button size="small" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <Card style={{ marginBottom: '16px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2 style={{ margin: 0 }}>
            <DashboardOutlined /> Grafana Instances
          </h2>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => refetch()} loading={isLoading}>
              Refresh
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
              Add Instance
            </Button>
          </Space>
        </div>
      </Card>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '16px' }}>
        <Col span={6}>
          <Card>
            <Statistic title="Total Instances" value={instances.length} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Active"
              value={activeCount}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="Total Dashboards" value={totalDashboards} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="Auto-Sync Enabled" value={autoSyncCount} />
          </Card>
        </Col>
      </Row>

      {/* Table */}
      <Card>
        {instances.length === 0 && !isLoading ? (
          <Empty description="No Grafana instances configured. Add your first Grafana instance to start syncing dashboards." />
        ) : (
          <Table
            columns={columns}
            dataSource={instances}
            rowKey="id"
            loading={isLoading}
            pagination={{
              total: instancesData?.total || 0,
              pageSize: instancesData?.pageSize || 20,
              current: instancesData?.page || 1,
              showSizeChanger: false,
            }}
          />
        )}
      </Card>

      {/* Add/Edit Modal */}
      <Modal
        title={editingInstance ? 'Edit Grafana Instance' : 'Add Grafana Instance'}
        open={isModalOpen}
        onOk={() => form.submit()}
        onCancel={() => {
          setIsModalOpen(false)
          setEditingInstance(null)
          form.resetFields()
          setTestResult(null)
        }}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        width={700}
      >
        <Form form={form} layout="vertical" onFinish={async (values) => {
          if (editingInstance) {
            updateMutation.mutate({
              id: editingInstance.id,
              data: values,
            })
          } else {
            createMutation.mutate(values as CreateGrafanaInstanceRequest)
          }
        }}>
          <Form.Item
            label="Cluster (Optional)"
            name="clusterId"
          >
            <Select placeholder="Select a cluster" allowClear>
              {/* TODO: Fetch clusters from API */}
              <Option value="">Select a cluster...</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Instance Name"
            name="name"
            rules={[{ required: true, message: 'Please enter an instance name' }]}
          >
            <Input placeholder="e.g., Production Grafana" />
          </Form.Item>

          <Form.Item
            label="Grafana URL"
            name="url"
            rules={[{ required: true, message: 'Please enter the Grafana URL' }]}
          >
            <Input placeholder="e.g., http://grafana:3000" />
          </Form.Item>

          <Form.Item label="API Key" name="apiKey">
            <Input.Password placeholder="Grafana API Key (recommended)" />
          </Form.Item>

          <Form.Item label="Username (Optional)" name="username">
            <Input placeholder="Basic auth username" />
          </Form.Item>

          <Form.Item label="Password (Optional)" name="password">
            <Input.Password placeholder="Basic auth password" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Auto Sync"
                name="autoSync"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Sync Interval (seconds)"
                name="syncInterval"
              >
                <InputNumber min={60} max={86400} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          {editingInstance && (
            <Form.Item
              label="Status"
              name="status"
              rules={[{ required: true }]}
            >
              <Select>
                <Option value="active">Active</Option>
                <Option value="inactive">Inactive</Option>
                <Option value="error">Error</Option>
              </Select>
            </Form.Item>
          )}

          {testResult && (
            <div style={{ marginBottom: 16 }}>
              <Tag color={testResult.success ? 'success' : 'error'} style={{ marginBottom: 8 }}>
                {testResult.message}
              </Tag>
            </div>
          )}

          <Form.Item>
            <Space>
              <Button
                onClick={handleTest}
                loading={testing}
                disabled={!form.getFieldValue('url')}
              >
                Test Connection
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default GrafanaInstancesPage
