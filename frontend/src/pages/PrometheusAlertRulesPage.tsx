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
  AlertOutlined,
  CheckCircleOutlined,
  FireOutlined,
  InfoCircleOutlined,
  WarningOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import prometheusApi, { type PrometheusAlertRule, type CreatePrometheusAlertRuleRequest } from '../api/prometheus'

const { Option } = Select
const { TextArea } = Input

export const PrometheusAlertRulesPage: React.FC = () => {
  const queryClient = useQueryClient()

  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<PrometheusAlertRule | null>(null)
  const [form] = Form.useForm()

  // Fetch alert rules
  const { data: alertRulesData, isLoading, refetch } = useQuery({
    queryKey: ['prometheusAlertRules'],
    queryFn: () => prometheusApi.getAlertRules({ page: 1, pageSize: 100 }),
  })

  // Fetch data sources for the dropdown
  const { data: dataSourcesData } = useQuery({
    queryKey: ['prometheusDataSources'],
    queryFn: () => prometheusApi.getDataSources({ page: 1, pageSize: 100 }),
  })

  const alertRules = alertRulesData?.data || []
  const dataSources = dataSourcesData?.data || []

  // Create mutation
  const createMutation = useMutation({
    mutationFn: prometheusApi.createAlertRule,
    onSuccess: () => {
      message.success('Alert rule created successfully')
      setIsModalOpen(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['prometheusAlertRules'] })
    },
    onError: (error: any) => {
      message.error(`Failed to create alert rule: ${error.response?.data?.message || error.message}`)
    },
  })

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      prometheusApi.updateAlertRule(id, data),
    onSuccess: () => {
      message.success('Alert rule updated successfully')
      setIsModalOpen(false)
      setEditingRule(null)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['prometheusAlertRules'] })
    },
    onError: (error: any) => {
      message.error(`Failed to update alert rule: ${error.response?.data?.message || error.message}`)
    },
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: prometheusApi.deleteAlertRule,
    onSuccess: () => {
      message.success('Alert rule deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['prometheusAlertRules'] })
    },
    onError: (error: any) => {
      message.error(`Failed to delete alert rule: ${error.response?.data?.message || error.message}`)
    },
  })

  // Handle add
  const handleAdd = () => {
    setEditingRule(null)
    form.resetFields()
    form.setFieldsValue({
      duration: 300,
      severity: 'warning',
      enabled: true,
    })
    setIsModalOpen(true)
  }

  // Handle edit
  const handleEdit = (rule: PrometheusAlertRule) => {
    setEditingRule(rule)
    form.setFieldsValue({
      dataSourceId: rule.dataSourceId,
      clusterId: rule.clusterId,
      name: rule.name,
      expression: rule.expression,
      duration: rule.duration,
      severity: rule.severity,
      summary: rule.summary,
      description: rule.description,
      enabled: rule.enabled,
    })
    setIsModalOpen(true)
  }

  // Handle delete
  const handleDelete = async (id: string) => {
    deleteMutation.mutate(id)
  }

  // Get severity tag
  const getSeverityTag = (severity: string) => {
    const severityMap: Record<string, { color: string; text: string; icon: React.ReactNode }> = {
      critical: { color: 'error', text: 'Critical', icon: <FireOutlined /> },
      warning: { color: 'warning', text: 'Warning', icon: <WarningOutlined /> },
      info: { color: 'processing', text: 'Info', icon: <InfoCircleOutlined /> },
    }
    const { color, text, icon } = severityMap[severity] || { color: 'default', text: severity, icon: null }
    return (
      <Tag color={color} icon={icon}>
        {text}
      </Tag>
    )
  }

  // Get enabled status
  const getEnabledStatus = (enabled: boolean, synced: boolean) => {
    if (!enabled) {
      return <Tag color="default">Disabled</Tag>
    }
    if (synced) {
      return <Tag color="success" icon={<CheckCircleOutlined />}>Active</Tag>
    }
    return <Tag color="warning">Pending Sync</Tag>
  }

  // Calculate statistics
  const criticalCount = alertRules.filter(r => r.severity === 'critical' && r.enabled).length
  const warningCount = alertRules.filter(r => r.severity === 'warning' && r.enabled).length
  const infoCount = alertRules.filter(r => r.severity === 'info' && r.enabled).length
  const totalTriggers = alertRules.reduce((sum, r) => sum + r.triggerCount, 0)

  // Table columns
  const columns: ColumnsType<PrometheusAlertRule> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: PrometheusAlertRule) => (
        <Space>
          <AlertOutlined />
          <span style={{ fontWeight: record.enabled ? 'normal' : 'normal', opacity: record.enabled ? 1 : 0.6 }}>
            {name}
          </span>
        </Space>
      ),
    },
    {
      title: 'Data Source',
      key: 'dataSource',
      render: (_: any, record: PrometheusAlertRule) => record.dataSource?.name || '-',
    },
    {
      title: 'Severity',
      dataIndex: 'severity',
      key: 'severity',
      width: 100,
      render: (severity: string) => getSeverityTag(severity),
    },
    {
      title: 'Expression',
      dataIndex: 'expression',
      key: 'expression',
      ellipsis: true,
      render: (expr: string) => (
        <code style={{ fontSize: '12px', background: '#f5f5f5', padding: '2px 6px', borderRadius: '3px' }}>
          {expr}
        </code>
      ),
    },
    {
      title: 'Duration',
      dataIndex: 'duration',
      key: 'duration',
      width: 100,
      render: (duration: number) => {
        const minutes = Math.floor(duration / 60)
        const seconds = duration % 60
        return minutes > 0 ? `${minutes}m ${seconds}s` : `${seconds}s`
      },
    },
    {
      title: 'Status',
      key: 'status',
      width: 120,
      render: (_: any, record: PrometheusAlertRule) => getEnabledStatus(record.enabled, record.synced),
    },
    {
      title: 'Triggers',
      dataIndex: 'triggerCount',
      key: 'triggerCount',
      width: 80,
      render: (count: number) => (
        <Tag color={count > 0 ? 'orange' : 'default'}>{count}</Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 150,
      render: (_: any, record: PrometheusAlertRule) => (
        <Space size="small">
          <Tooltip title="Edit">
            <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          </Tooltip>
          <Popconfirm
            title="Delete Alert Rule"
            description="Are you sure you want to delete this alert rule?"
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
            <AlertOutlined /> Prometheus Alert Rules
          </h2>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => refetch()} loading={isLoading}>
              Refresh
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
              Add Alert Rule
            </Button>
          </Space>
        </div>
      </Card>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '16px' }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="Critical"
              value={criticalCount}
              valueStyle={{ color: '#cf1322' }}
              prefix={<FireOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Warning"
              value={warningCount}
              valueStyle={{ color: '#faad14' }}
              prefix={<WarningOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Info"
              value={infoCount}
              valueStyle={{ color: '#1677ff' }}
              prefix={<InfoCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Total Triggers"
              value={totalTriggers}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Table */}
      <Card>
        {alertRules.length === 0 && !isLoading ? (
          <Empty description="No alert rules configured. Add your first alert rule to start monitoring metrics." />
        ) : (
          <Table
            columns={columns}
            dataSource={alertRules}
            rowKey="id"
            loading={isLoading}
            pagination={{
              total: alertRulesData?.total || 0,
              pageSize: alertRulesData?.pageSize || 20,
              current: alertRulesData?.page || 1,
              showSizeChanger: false,
            }}
          />
        )}
      </Card>

      {/* Add/Edit Modal */}
      <Modal
        title={editingRule ? 'Edit Alert Rule' : 'Add Alert Rule'}
        open={isModalOpen}
        onOk={() => form.submit()}
        onCancel={() => {
          setIsModalOpen(false)
          setEditingRule(null)
          form.resetFields()
        }}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        width={700}
      >
        <Form form={form} layout="vertical" onFinish={async (values) => {
          if (editingRule) {
            updateMutation.mutate({
              id: editingRule.id,
              data: values,
            })
          } else {
            createMutation.mutate(values as CreatePrometheusAlertRuleRequest)
          }
        }}>
          <Form.Item
            label="Data Source"
            name="dataSourceId"
            rules={[{ required: true, message: 'Please select a data source' }]}
          >
            <Select placeholder="Select a Prometheus data source">
              {dataSources.map(ds => (
                <Option key={ds.id} value={ds.id}>{ds.name}</Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            label="Rule Name"
            name="name"
            rules={[{ required: true, message: 'Please enter a rule name' }]}
          >
            <Input placeholder="e.g., HighCPUUsage" />
          </Form.Item>

          <Form.Item
            label="PromQL Expression"
            name="expression"
            rules={[{ required: true, message: 'Please enter a PromQL expression' }]}
            tooltip="The PromQL expression to evaluate"
          >
            <TextArea
              rows={3}
              placeholder="e.g., rate(cpu_usage_total[5m]) > 0.8"
              style={{ fontFamily: 'monospace' }}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Duration"
                name="duration"
                rules={[{ required: true, message: 'Please enter duration' }]}
                tooltip="How long the condition must be true before triggering"
              >
                <InputNumber
                  min={1}
                  max={86400}
                  style={{ width: '100%' }}
                  addonAfter="seconds"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Severity"
                name="severity"
                rules={[{ required: true, message: 'Please select severity' }]}
              >
                <Select>
                  <Option value="critical">Critical</Option>
                  <Option value="warning">Warning</Option>
                  <Option value="info">Info</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item label="Summary" name="summary">
            <Input placeholder="Brief summary of the alert" />
          </Form.Item>

          <Form.Item label="Description" name="description">
            <TextArea rows={3} placeholder="Detailed description of the alert" />
          </Form.Item>

          <Form.Item
            label="Enabled"
            name="enabled"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default PrometheusAlertRulesPage
