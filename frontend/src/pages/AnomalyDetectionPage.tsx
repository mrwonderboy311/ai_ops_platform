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
  ThunderboltOutlined,
  PlayCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  LineChartOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import aiAnalysisApi, { type AnomalyDetectionRule, type CreateAnomalyDetectionRuleRequest } from '../api/aiAnalysis'

const { Option } = Select
const { TextArea } = Input

export const AnomalyDetectionPage: React.FC = () => {
  const queryClient = useQueryClient()

  const [isRuleModalOpen, setIsRuleModalOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<AnomalyDetectionRule | null>(null)
  const [executing, setExecuting] = useState<Record<string, boolean>>({})
  const [form] = Form.useForm()

  // Fetch rules
  const { data: rulesData, isLoading, refetch } = useQuery({
    queryKey: ['anomalyRules'],
    queryFn: () => aiAnalysisApi.getAnomalyRules({ page: 1, pageSize: 100 }),
  })

  // Fetch events
  const { data: eventsData } = useQuery({
    queryKey: ['anomalyEvents'],
    queryFn: () => aiAnalysisApi.getAnomalyEvents({ page: 1, pageSize: 20 }),
  })

  const rules = rulesData?.data || []
  const events = eventsData?.data || []

  // Create mutation
  const createMutation = useMutation({
    mutationFn: aiAnalysisApi.createAnomalyRule,
    onSuccess: () => {
      message.success('Anomaly detection rule created successfully')
      setIsRuleModalOpen(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['anomalyRules'] })
    },
    onError: (error: any) => {
      message.error(`Failed to create rule: ${error.response?.data?.message || error.message}`)
    },
  })

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      aiAnalysisApi.updateAnomalyRule(id, data),
    onSuccess: () => {
      message.success('Anomaly detection rule updated successfully')
      setIsRuleModalOpen(false)
      setEditingRule(null)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['anomalyRules'] })
    },
    onError: (error: any) => {
      message.error(`Failed to update rule: ${error.response?.data?.message || error.message}`)
    },
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: aiAnalysisApi.deleteAnomalyRule,
    onSuccess: () => {
      message.success('Anomaly detection rule deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['anomalyRules'] })
    },
    onError: (error: any) => {
      message.error(`Failed to delete rule: ${error.response?.data?.message || error.message}`)
    },
  })

  // Handle add
  const handleAdd = () => {
    setEditingRule(null)
    form.resetFields()
    form.setFieldsValue({
      algorithm: 'stl',
      sensitivity: 0.95,
      windowSize: 100,
      enabled: true,
      evalInterval: 300,
      alertThreshold: 0.8,
      alertOnRecovery: false,
    })
    setIsRuleModalOpen(true)
  }

  // Handle edit
  const handleEdit = (rule: AnomalyDetectionRule) => {
    setEditingRule(rule)
    form.setFieldsValue({
      clusterId: rule.clusterId,
      dataSourceId: rule.dataSourceId,
      name: rule.name,
      description: rule.description,
      metricQuery: rule.metricQuery,
      algorithm: rule.algorithm,
      sensitivity: rule.sensitivity,
      windowSize: rule.windowSize,
      minValue: rule.minValue,
      maxValue: rule.maxValue,
      enabled: rule.enabled,
      evalInterval: rule.evalInterval,
      alertThreshold: rule.alertThreshold,
      alertOnRecovery: rule.alertOnRecovery,
    })
    setIsRuleModalOpen(true)
  }

  // Handle delete
  const handleDelete = async (id: string) => {
    deleteMutation.mutate(id)
  }

  // Handle execute
  const handleExecute = async (rule: AnomalyDetectionRule) => {
    try {
      setExecuting({ ...executing, [rule.id]: true })
      const result = await aiAnalysisApi.executeAnomalyDetection({ ruleId: rule.id })
      message.success(`Detection completed. Found ${result.anomalyCount} anomalies`)
      queryClient.invalidateQueries({ queryKey: ['anomalyRules'] })
      queryClient.invalidateQueries({ queryKey: ['anomalyEvents'] })
    } catch (error: any) {
      message.error(`Detection failed: ${error.response?.data?.message || error.message}`)
    } finally {
      setExecuting({ ...executing, [rule.id]: false })
    }
  }

  // Get algorithm tag
  const getAlgorithmTag = (algorithm: string) => {
    const algoMap: Record<string, { color: string; text: string; icon: React.ReactNode }> = {
      stl: { color: 'blue', text: 'STL', icon: <LineChartOutlined /> },
      isolation_forest: { color: 'purple', text: 'Isolation Forest', icon: <ThunderboltOutlined /> },
      lstm: { color: 'green', text: 'LSTM', icon: <LineChartOutlined /> },
      baseline: { color: 'orange', text: 'Baseline', icon: <LineChartOutlined /> },
    }
    const { color, text, icon } = algoMap[algorithm] || { color: 'default', text: algorithm, icon: null }
    return (
      <Tag color={color} icon={icon}>
        {text}
      </Tag>
    )
  }

  // Calculate statistics
  const totalRules = rules.length
  const activeRules = rules.filter(r => r.enabled).length
  const totalAnomalies = rules.reduce((sum, r) => sum + r.anomaliesDetected, 0)
  const totalEvaluations = rules.reduce((sum, r) => sum + r.totalEvaluations, 0)

  // Table columns
  const columns: ColumnsType<AnomalyDetectionRule> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: AnomalyDetectionRule) => (
        <Space>
          <ThunderboltOutlined style={{ color: record.enabled ? '#faad14' : '#999' }} />
          <span style={{ fontWeight: record.enabled ? 'bold' : 'normal', opacity: record.enabled ? 1 : 0.6 }}>
            {name}
          </span>
        </Space>
      ),
    },
    {
      title: 'Algorithm',
      dataIndex: 'algorithm',
      key: 'algorithm',
      width: 150,
      render: (algorithm: string) => getAlgorithmTag(algorithm),
    },
    {
      title: 'Metric Query',
      dataIndex: 'metricQuery',
      key: 'metricQuery',
      ellipsis: true,
      render: (query: string) => (
        <code style={{ fontSize: '12px', background: '#f5f5f5', padding: '2px 6px', borderRadius: '3px' }}>
          {query}
        </code>
      ),
    },
    {
      title: 'Sensitivity',
      dataIndex: 'sensitivity',
      key: 'sensitivity',
      width: 100,
      render: (sensitivity: number) => (
        <Tag color={sensitivity > 0.9 ? 'error' : sensitivity > 0.7 ? 'warning' : 'success'}>
          {(sensitivity * 100).toFixed(0)}%
        </Tag>
      ),
    },
    {
      title: 'Anomalies',
      dataIndex: 'anomaliesDetected',
      key: 'anomaliesDetected',
      width: 100,
      render: (count: number) => (
        <Tag color={count > 0 ? 'orange' : 'default'}>{count}</Tag>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 100,
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'success' : 'default'} icon={enabled ? <CheckCircleOutlined /> : <CloseCircleOutlined />}>
          {enabled ? 'Active' : 'Inactive'}
        </Tag>
      ),
    },
    {
      title: 'Last Eval',
      dataIndex: 'lastEvalAt',
      key: 'lastEvalAt',
      width: 150,
      render: (date: string) => (date ? new Date(date).toLocaleString() : '-'),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (_: any, record: AnomalyDetectionRule) => (
        <Space size="small">
          <Tooltip title="Execute Now">
            <Button
              size="small"
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={() => handleExecute(record)}
              loading={executing[record.id]}
              disabled={!record.enabled}
            />
          </Tooltip>
          <Tooltip title="Edit">
            <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          </Tooltip>
          <Popconfirm
            title="Delete Rule"
            description="Are you sure you want to delete this anomaly detection rule?"
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
            <ThunderboltOutlined /> AI Anomaly Detection
          </h2>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => refetch()} loading={isLoading}>
              Refresh
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
              Add Detection Rule
            </Button>
          </Space>
        </div>
      </Card>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '16px' }}>
        <Col span={6}>
          <Card>
            <Statistic title="Total Rules" value={totalRules} prefix={<ExclamationCircleOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Active Rules"
              value={activeRules}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Anomalies Detected"
              value={totalAnomalies}
              valueStyle={{ color: totalAnomalies > 0 ? '#cf1322' : undefined }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="Total Evaluations" value={totalEvaluations} />
          </Card>
        </Col>
      </Row>

      {/* Rules Table */}
      <Card title="Detection Rules" style={{ marginBottom: '16px' }}>
        {rules.length === 0 && !isLoading ? (
          <Empty description="No anomaly detection rules configured. Add your first rule to start detecting anomalies." />
        ) : (
          <Table
            columns={columns}
            dataSource={rules}
            rowKey="id"
            loading={isLoading}
            pagination={{
              total: rulesData?.total || 0,
              pageSize: rulesData?.pageSize || 20,
              current: rulesData?.page || 1,
              showSizeChanger: false,
            }}
          />
        )}
      </Card>

      {/* Recent Anomalies */}
      <Card title="Recent Anomalies" style={{ marginBottom: '16px' }}>
        {events.length === 0 ? (
          <Empty description="No anomalies detected yet." />
        ) : (
          <Table
            columns={[
              { title: 'Severity', dataIndex: 'severity', key: 'severity', width: 100, render: (s: string) => {
                const severityMap: Record<string, { color: string; text: string }> = {
                  critical: { color: 'error', text: 'Critical' },
                  warning: { color: 'warning', text: 'Warning' },
                  info: { color: 'processing', text: 'Info' },
                }
                const { color, text } = severityMap[s] || { color: 'default', text: s }
                return <Tag color={color}>{text}</Tag>
              }},
              { title: 'Metric', dataIndex: 'metricName', key: 'metricName' },
              { title: 'Current', dataIndex: 'currentValue', key: 'currentValue', render: (v: number) => v.toFixed(2) },
              { title: 'Expected', dataIndex: 'expectedValue', key: 'expectedValue', render: (v: number) => v.toFixed(2) },
              { title: 'Deviation', dataIndex: 'deviation', key: 'deviation', render: (v: number) => `${v.toFixed(2)} sigma` },
              { title: 'Time', dataIndex: 'createdAt', key: 'createdAt', render: (t: string) => new Date(t).toLocaleString() },
            ] as ColumnsType<any>}
            dataSource={events}
            rowKey="id"
            pagination={false}
            size="small"
          />
        )}
      </Card>

      {/* Add/Edit Rule Modal */}
      <Modal
        title={editingRule ? 'Edit Detection Rule' : 'Add Detection Rule'}
        open={isRuleModalOpen}
        onOk={() => form.submit()}
        onCancel={() => {
          setIsRuleModalOpen(false)
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
            createMutation.mutate(values as CreateAnomalyDetectionRuleRequest)
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
            label="Data Source (Optional)"
            name="dataSourceId"
          >
            <Select placeholder="Select a Prometheus data source" allowClear>
              {/* TODO: Fetch data sources from API */}
              <Option value="">Select a data source...</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Rule Name"
            name="name"
            rules={[{ required: true, message: 'Please enter a rule name' }]}
          >
            <Input placeholder="e.g., High CPU Anomaly Detection" />
          </Form.Item>

          <Form.Item label="Description" name="description">
            <TextArea rows={2} placeholder="Describe what this rule detects" />
          </Form.Item>

          <Form.Item
            label="PromQL Query"
            name="metricQuery"
            rules={[{ required: true, message: 'Please enter a PromQL query' }]}
            tooltip="The metric query to evaluate for anomalies"
          >
            <TextArea
              rows={2}
              placeholder="e.g., rate(cpu_usage_total[5m])"
              style={{ fontFamily: 'monospace' }}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Algorithm"
                name="algorithm"
                rules={[{ required: true, message: 'Please select an algorithm' }]}
              >
                <Select>
                  <Option value="stl">STL (Seasonal Decomposition)</Option>
                  <Option value="isolation_forest">Isolation Forest</Option>
                  <Option value="lstm">LSTM (Deep Learning)</Option>
                  <Option value="baseline">Baseline Learning</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Sensitivity"
                name="sensitivity"
                rules={[{ required: true, message: 'Please enter sensitivity' }]}
                tooltip="Higher sensitivity = more anomalies detected (0-1)"
              >
                <InputNumber min={0} max={1} step={0.05} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item label="Window Size" name="windowSize">
                <InputNumber min={10} max={1000} style={{ width: '100%' }} addonAfter="points" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="Min Value" name="minValue">
                <InputNumber style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="Max Value" name="maxValue">
                <InputNumber style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Evaluation Interval"
                name="evalInterval"
              >
                <InputNumber min={60} max={86400} style={{ width: '100%' }} addonAfter="seconds" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Alert Threshold"
                name="alertThreshold"
              >
                <InputNumber min={0} max={1} step={0.1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="Enabled"
            name="enabled"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          <Form.Item
            label="Alert on Recovery"
            name="alertOnRecovery"
            valuePropName="checked"
            tooltip="Send an alert when the anomaly is resolved"
          >
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default AnomalyDetectionPage
