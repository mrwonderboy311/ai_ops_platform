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
  Statistic,
  Row,
  Col,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  DatabaseOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ApiOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import prometheusApi, { type PrometheusDataSource, type CreatePrometheusDataSourceRequest } from '../api/prometheus'

const { Option } = Select

export const PrometheusDataSourcePage: React.FC = () => {
  const queryClient = useQueryClient()

  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingDataSource, setEditingDataSource] = useState<PrometheusDataSource | null>(null)
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null)
  const [form] = Form.useForm()

  // Fetch data sources
  const { data: dataSourcesData, isLoading, refetch } = useQuery({
    queryKey: ['prometheusDataSources'],
    queryFn: () => prometheusApi.getDataSources({ page: 1, pageSize: 100 }),
  })

  const dataSources = dataSourcesData?.data || []

  // Create mutation
  const createMutation = useMutation({
    mutationFn: prometheusApi.createDataSource,
    onSuccess: () => {
      message.success('Data source created successfully')
      setIsModalOpen(false)
      form.resetFields()
      setTestResult(null)
      queryClient.invalidateQueries({ queryKey: ['prometheusDataSources'] })
    },
    onError: (error: any) => {
      message.error(`Failed to create data source: ${error.response?.data?.message || error.message}`)
    },
  })

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      prometheusApi.updateDataSource(id, data),
    onSuccess: () => {
      message.success('Data source updated successfully')
      setIsModalOpen(false)
      setEditingDataSource(null)
      form.resetFields()
      setTestResult(null)
      queryClient.invalidateQueries({ queryKey: ['prometheusDataSources'] })
    },
    onError: (error: any) => {
      message.error(`Failed to update data source: ${error.response?.data?.message || error.message}`)
    },
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: prometheusApi.deleteDataSource,
    onSuccess: () => {
      message.success('Data source deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['prometheusDataSources'] })
    },
    onError: (error: any) => {
      message.error(`Failed to delete data source: ${error.response?.data?.message || error.message}`)
    },
  })

  // Handle add
  const handleAdd = () => {
    setEditingDataSource(null)
    setTestResult(null)
    form.resetFields()
    form.setFieldsValue({
      insecureSkipTLS: false,
    })
    setIsModalOpen(true)
  }

  // Handle edit
  const handleEdit = (dataSource: PrometheusDataSource) => {
    setEditingDataSource(dataSource)
    setTestResult(null)
    form.setFieldsValue({
      clusterId: dataSource.clusterId,
      name: dataSource.name,
      url: dataSource.url,
      username: dataSource.username,
      insecureSkipTLS: dataSource.insecureSkipTLS,
      status: dataSource.status,
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
      const values = await form.validateFields(['url', 'username', 'password', 'insecureSkipTLS'])
      const result = await prometheusApi.testDataSource({
        url: values.url,
        username: values.username,
        password: values.password,
        insecureSkipTLS: values.insecureSkipTLS,
      })
      setTestResult({
        success: result.success,
        message: result.success ? `Connected successfully (${result.version})` : result.message,
      })
      if (!result.success) {
        message.error(result.message)
      } else {
        message.success('Connection test successful')
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

  // Get status tag
  const getStatusTag = (status: string) => {
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
  const totalQueries = dataSources.reduce((sum, ds) => sum + ds.queryCount, 0)
  const activeCount = dataSources.filter(ds => ds.status === 'active').length
  const avgQueryTime = dataSources.length > 0
    ? dataSources.reduce((sum, ds) => sum + ds.averageQueryTime, 0) / dataSources.length
    : 0

  // Table columns
  const columns: ColumnsType<PrometheusDataSource> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <DatabaseOutlined />
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
        <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>{url}</span>
      ),
    },
    {
      title: 'Cluster',
      key: 'cluster',
      render: (_: any, record: PrometheusDataSource) => record.cluster?.name || '-',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => getStatusTag(status),
    },
    {
      title: 'Queries',
      dataIndex: 'queryCount',
      key: 'queryCount',
      width: 100,
      render: (count: number) => (
        <Tag icon={<ApiOutlined />} color="blue">
          {count.toLocaleString()}
        </Tag>
      ),
    },
    {
      title: 'Last Tested',
      dataIndex: 'lastTestAt',
      key: 'lastTestAt',
      width: 150,
      render: (date: string) => (date ? new Date(date).toLocaleString() : '-'),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 150,
      render: (_: any, record: PrometheusDataSource) => (
        <Space size="small">
          <Tooltip title="Edit">
            <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          </Tooltip>
          <Popconfirm
            title="Delete Data Source"
            description="Are you sure you want to delete this data source?"
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
            <DatabaseOutlined /> Prometheus Data Sources
          </h2>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => refetch()} loading={isLoading}>
              Refresh
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
              Add Data Source
            </Button>
          </Space>
        </div>
      </Card>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '16px' }}>
        <Col span={6}>
          <Card>
            <Statistic title="Total Data Sources" value={dataSources.length} />
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
            <Statistic title="Total Queries" value={totalQueries} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Avg Query Time"
              value={avgQueryTime.toFixed(2)}
              suffix="ms"
            />
          </Card>
        </Col>
      </Row>

      {/* Table */}
      <Card>
        {dataSources.length === 0 && !isLoading ? (
          <Empty description="No data sources configured. Add your first Prometheus data source to start querying metrics." />
        ) : (
          <Table
            columns={columns}
            dataSource={dataSources}
            rowKey="id"
            loading={isLoading}
            pagination={{
              total: dataSourcesData?.total || 0,
              pageSize: dataSourcesData?.pageSize || 20,
              current: dataSourcesData?.page || 1,
              showSizeChanger: false,
            }}
          />
        )}
      </Card>

      {/* Add/Edit Modal */}
      <Modal
        title={editingDataSource ? 'Edit Data Source' : 'Add Data Source'}
        open={isModalOpen}
        onOk={() => form.submit()}
        onCancel={() => {
          setIsModalOpen(false)
          setEditingDataSource(null)
          form.resetFields()
          setTestResult(null)
        }}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        width={700}
      >
        <Form form={form} layout="vertical" onFinish={async (values) => {
          if (editingDataSource) {
            updateMutation.mutate({
              id: editingDataSource.id,
              data: values,
            })
          } else {
            createMutation.mutate(values as CreatePrometheusDataSourceRequest)
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
            label="Name"
            name="name"
            rules={[{ required: true, message: 'Please enter a name' }]}
          >
            <Input placeholder="e.g., Production Prometheus" />
          </Form.Item>

          <Form.Item
            label="URL"
            name="url"
            rules={[{ required: true, message: 'Please enter the Prometheus URL' }]}
          >
            <Input placeholder="e.g., http://prometheus:9090" />
          </Form.Item>

          <Form.Item label="Username (Optional)" name="username">
            <Input placeholder="Basic auth username" />
          </Form.Item>

          <Form.Item label="Password (Optional)" name="password">
            <Input.Password placeholder="Basic auth password" />
          </Form.Item>

          <Form.Item
            label="Skip TLS Verification"
            name="insecureSkipTLS"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          {editingDataSource && (
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

export default PrometheusDataSourcePage
