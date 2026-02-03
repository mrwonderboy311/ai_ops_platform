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
  InputNumber,
  Empty,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  StopOutlined,
  ReloadOutlined,
  CloudServerOutlined,
  ApiOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import otelApi, { type OtelCollector, type CreateCollectorRequest } from '../api/otel'

const { Option } = Select

export const OtelCollectorPage: React.FC = () => {
  const queryClient = useQueryClient()

  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingCollector, setEditingCollector] = useState<OtelCollector | null>(null)
  const [form] = Form.useForm()

  // Fetch collectors
  const { data: collectorsData, isLoading, refetch } = useQuery({
    queryKey: ['otelCollectors'],
    queryFn: () => otelApi.getCollectors({ page: 1, pageSize: 100 }),
  })

  const collectors = collectorsData?.data || []

  // Create mutation
  const createMutation = useMutation({
    mutationFn: otelApi.createCollector,
    onSuccess: () => {
      message.success('Collector created successfully')
      setIsModalOpen(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['otelCollectors'] })
    },
    onError: (error: any) => {
      message.error(`Failed to create collector: ${error.response?.data?.message || error.message}`)
    },
  })

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      otelApi.updateCollector(id, data),
    onSuccess: () => {
      message.success('Collector updated successfully')
      setIsModalOpen(false)
      setEditingCollector(null)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['otelCollectors'] })
    },
    onError: (error: any) => {
      message.error(`Failed to update collector: ${error.response?.data?.message || error.message}`)
    },
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: otelApi.deleteCollector,
    onSuccess: () => {
      message.success('Collector deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['otelCollectors'] })
    },
    onError: (error: any) => {
      message.error(`Failed to delete collector: ${error.response?.data?.message || error.message}`)
    },
  })

  // Start mutation
  const startMutation = useMutation({
    mutationFn: otelApi.startCollector,
    onSuccess: () => {
      message.success('Collector started')
      queryClient.invalidateQueries({ queryKey: ['otelCollectors'] })
    },
    onError: (error: any) => {
      message.error(`Failed to start collector: ${error.response?.data?.message || error.message}`)
    },
  })

  // Stop mutation
  const stopMutation = useMutation({
    mutationFn: otelApi.stopCollector,
    onSuccess: () => {
      message.success('Collector stopped')
      queryClient.invalidateQueries({ queryKey: ['otelCollectors'] })
    },
    onError: (error: any) => {
      message.error(`Failed to stop collector: ${error.response?.data?.message || error.message}`)
    },
  })

  // Restart mutation
  const restartMutation = useMutation({
    mutationFn: otelApi.restartCollector,
    onSuccess: () => {
      message.success('Collector restarted')
      queryClient.invalidateQueries({ queryKey: ['otelCollectors'] })
    },
    onError: (error: any) => {
      message.error(`Failed to restart collector: ${error.response?.data?.message || error.message}`)
    },
  })

  // Handle add
  const handleAdd = () => {
    setEditingCollector(null)
    form.resetFields()
    form.setFieldsValue({
      namespace: 'observability',
      type: 'all',
      replicas: 1,
    })
    setIsModalOpen(true)
  }

  // Handle edit
  const handleEdit = (collector: OtelCollector) => {
    setEditingCollector(collector)
    form.setFieldsValue({
      name: collector.name,
      namespace: collector.namespace,
      type: collector.type,
      replicas: collector.replicas,
      metricsEndpoint: collector.metricsEndpoint,
      logsEndpoint: collector.logsEndpoint,
      tracesEndpoint: collector.tracesEndpoint,
    })
    setIsModalOpen(true)
  }

  // Handle delete
  const handleDelete = async (id: string) => {
    deleteMutation.mutate(id)
  }

  // Handle start
  const handleStart = async (id: string) => {
    startMutation.mutate(id)
  }

  // Handle stop
  const handleStop = async (id: string) => {
    stopMutation.mutate(id)
  }

  // Handle restart
  const handleRestart = async (id: string) => {
    restartMutation.mutate(id)
  }

  // Get status tag
  const getStatusTag = (collector: OtelCollector) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      deploying: { color: 'processing', text: 'Deploying' },
      running: { color: 'success', text: 'Running' },
      stopped: { color: 'default', text: 'Stopped' },
      error: { color: 'error', text: 'Error' },
      pending: { color: 'default', text: 'Pending' },
    }
    const { color, text } = statusMap[collector.status] || { color: 'default', text: collector.status }
    return <Tag color={color}>{text}</Tag>
  }

  // Get type tag
  const getTypeTag = (type: string) => {
    const typeMap: Record<string, string> = {
      metrics: 'blue',
      logs: 'green',
      traces: 'purple',
      all: 'orange',
    }
    const color = typeMap[type] || 'default'
    return <Tag color={color}>{type.toUpperCase()}</Tag>
  }

  // Table columns
  const columns: ColumnsType<OtelCollector> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <ApiOutlined />
          <span>{name}</span>
        </Space>
      ),
    },
    {
      title: 'Cluster',
      key: 'cluster',
      render: (_: any, record: OtelCollector) => record.cluster?.name || '-',
    },
    {
      title: 'Namespace',
      dataIndex: 'namespace',
      key: 'namespace',
      render: (namespace: string) => <Tag>{namespace}</Tag>,
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => getTypeTag(type),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (_: any, record: OtelCollector) => getStatusTag(record),
    },
    {
      title: 'Replicas',
      dataIndex: 'replicas',
      key: 'replicas',
      width: 80,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 250,
      render: (_: any, record: OtelCollector) => (
        <Space size="small">
          {record.status === 'running' || record.status === 'deploying' ? (
            <Tooltip title="Stop">
              <Button
                size="small"
                icon={<StopOutlined />}
                onClick={() => handleStop(record.id)}
                loading={stopMutation.isPending}
              />
            </Tooltip>
          ) : (
            <Tooltip title="Start">
              <Button
                size="small"
                type="primary"
                icon={<PlayCircleOutlined />}
                onClick={() => handleStart(record.id)}
                loading={startMutation.isPending}
              />
            </Tooltip>
          )}
          <Tooltip title="Restart">
            <Button
              size="small"
              icon={<ReloadOutlined />}
              onClick={() => handleRestart(record.id)}
              loading={restartMutation.isPending}
            />
          </Tooltip>
          <Tooltip title="Edit">
            <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          </Tooltip>
          <Popconfirm
            title="Delete Collector"
            description="Are you sure you want to delete this collector?"
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
            <CloudServerOutlined /> OpenTelemetry Collectors
          </h2>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => refetch()} loading={isLoading}>
              Refresh
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
              Add Collector
            </Button>
          </Space>
        </div>
      </Card>

      {/* Table */}
      <Card>
        {collectors.length === 0 && !isLoading ? (
          <Empty description="No collectors configured. Add your first OpenTelemetry collector to start collecting observability data." />
        ) : (
          <Table
            columns={columns}
            dataSource={collectors}
            rowKey="id"
            loading={isLoading}
            pagination={{
              total: collectorsData?.total || 0,
              pageSize: collectorsData?.pageSize || 20,
              current: collectorsData?.page || 1,
              showSizeChanger: false,
            }}
          />
        )}
      </Card>

      {/* Add/Edit Modal */}
      <Modal
        title={editingCollector ? 'Edit Collector' : 'Add Collector'}
        open={isModalOpen}
        onOk={() => form.submit()}
        onCancel={() => {
          setIsModalOpen(false)
          setEditingCollector(null)
          form.resetFields()
        }}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={async (values) => {
          if (editingCollector) {
            updateMutation.mutate({
              id: editingCollector.id,
              data: values,
            })
          } else {
            createMutation.mutate(values as CreateCollectorRequest)
          }
        }}>
          <Form.Item
            label="Cluster"
            name="clusterId"
            rules={[{ required: true, message: 'Please select a cluster' }]}
          >
            <Select placeholder="Select a cluster">
              {/* TODO: Fetch clusters from API */}
              <Option value="">Select a cluster...</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Name"
            name="name"
            rules={[{ required: true, message: 'Please enter a name' }]}
          >
            <Input placeholder="e.g., otel-collector" />
          </Form.Item>

          <Form.Item
            label="Namespace"
            name="namespace"
            rules={[{ required: true, message: 'Please enter a namespace' }]}
          >
            <Input placeholder="e.g., observability" />
          </Form.Item>

          <Form.Item
            label="Type"
            name="type"
            rules={[{ required: true, message: 'Please select a type' }]}
          >
            <Select>
              <Option value="all">All (Metrics + Logs + Traces)</Option>
              <Option value="metrics">Metrics Only</Option>
              <Option value="logs">Logs Only</Option>
              <Option value="traces">Traces Only</Option>
            </Select>
          </Form.Item>

          <Form.Item label="Replicas" name="replicas" rules={[{ required: true }]}>
            <InputNumber min={1} max={10} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item label="Metrics Endpoint" name="metricsEndpoint">
            <Input placeholder="e.g., http://prometheus:9090/api/v1/write" />
          </Form.Item>

          <Form.Item label="Logs Endpoint" name="logsEndpoint">
            <Input placeholder="e.g., http://loki:3100/loki/api/v1/push" />
          </Form.Item>

          <Form.Item label="Traces Endpoint" name="tracesEndpoint">
            <Input placeholder="e.g., http://tempo:4318" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default OtelCollectorPage
