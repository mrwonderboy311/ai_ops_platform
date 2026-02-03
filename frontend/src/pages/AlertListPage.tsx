import React, { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button, Space, message, Table, Tag, Tabs, Row, Col, Statistic, Card, Modal, Form, Input, Select, InputNumber } from 'antd'
import {
  BellOutlined,
  PlusOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  WarningOutlined,
  InfoCircleOutlined,
  DeleteOutlined,
  EditOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { alertApi } from '../api/alert'
import type { Alert, AlertRule, AlertSeverity, AlertStatus } from '../types/alert'

const { Option } = Select
const { TextArea } = Input

export const AlertListPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState('alerts')
  const [alertStatus] = useState<string>('')
  const [ruleModal, setRuleModal] = useState<{ visible: boolean; rule?: AlertRule }>({
    visible: false,
  })

  const [form] = Form.useForm()

  // Fetch alert statistics
  const { data: stats, isLoading: statsLoading, refetch: refetchStats } = useQuery({
    queryKey: ['alertStats'],
    queryFn: () => alertApi.getAlertStatistics(),
  })

  // Fetch alerts
  const { data: alertsData, isLoading: alertsLoading, refetch: refetchAlerts } = useQuery({
    queryKey: ['alerts', alertStatus],
    queryFn: () => alertApi.getAlerts({ status: alertStatus }),
  })

  // Fetch alert rules
  const { data: rulesData, isLoading: rulesLoading, refetch: refetchRules } = useQuery({
    queryKey: ['alertRules'],
    queryFn: () => alertApi.getAlertRules(),
  })

  // Handle refresh
  const handleRefresh = () => {
    refetchStats()
    refetchAlerts()
    refetchRules()
  }

  // Handle silence alert
  const handleSilenceAlert = async (alertId: string) => {
    Modal.confirm({
      title: 'Silence Alert',
      content: (
        <div>
          <p>How long do you want to silence this alert?</p>
          <Select defaultValue="1h" style={{ width: 200 }}>
            <Option value="1h">1 hour</Option>
            <Option value="6h">6 hours</Option>
            <Option value="24h">24 hours</Option>
            <Option value="7d">7 days</Option>
          </Select>
        </div>
      ),
      onOk: async () => {
        try {
          await alertApi.silenceAlert(alertId, '24h')
          message.success('Alert silenced successfully')
          refetchAlerts()
        } catch (error: any) {
          message.error(`Failed to silence alert: ${error.response?.data?.message || error.message}`)
        }
      },
    })
  }

  // Get severity icon and color
  const getSeverityConfig = (severity: AlertSeverity) => {
    switch (severity) {
      case 'critical':
        return { color: 'error', icon: <CloseCircleOutlined />, label: 'Critical' }
      case 'warning':
        return { color: 'warning', icon: <WarningOutlined />, label: 'Warning' }
      case 'info':
      default:
        return { color: 'default', icon: <InfoCircleOutlined />, label: 'Info' }
    }
  }

  // Get status icon and color
  const getStatusConfig = (status: AlertStatus) => {
    switch (status) {
      case 'firing':
        return { color: 'error', icon: <BellOutlined />, label: 'Firing' }
      case 'resolved':
        return { color: 'success', icon: <CheckCircleOutlined />, label: 'Resolved' }
      case 'silenced':
        return { color: 'default', icon: <BellOutlined />, label: 'Silenced' }
      default:
        return { color: 'default', icon: <BellOutlined />, label: status }
    }
  }

  // Alert table columns
  const alertColumns: ColumnsType<Alert> = [
    {
      title: 'Severity',
      dataIndex: 'severity',
      key: 'severity',
      width: 100,
      render: (severity: AlertSeverity) => {
        const config = getSeverityConfig(severity)
        return <Tag color={config.color} icon={config.icon}>{config.label}</Tag>
      },
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: AlertStatus) => {
        const config = getStatusConfig(status)
        return <Tag color={config.color} icon={config.icon}>{config.label}</Tag>
      },
    },
    {
      title: 'Title',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
    },
    {
      title: 'Value',
      key: 'value',
      width: 120,
      render: (_: any, record: Alert) => (
        <span style={{ color: record.value > record.threshold ? '#cf1322' : '#3f8600' }}>
          {record.value.toFixed(2)} / {record.threshold.toFixed(2)}
        </span>
      ),
    },
    {
      title: 'Started',
      dataIndex: 'startedAt',
      key: 'startedAt',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      render: (_: any, record: Alert) => (
        <Space>
          {record.status === 'firing' && (
            <Button size="small" onClick={() => handleSilenceAlert(record.id)}>
              Silence
            </Button>
          )}
        </Space>
      ),
    },
  ]

  // Alert rule table columns
  const ruleColumns: ColumnsType<AlertRule> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: AlertRule) => (
        <Space>
          {name}
          {!record.enabled && <Tag color="default">Disabled</Tag>}
        </Space>
      ),
    },
    {
      title: 'Target',
      key: 'target',
      width: 150,
      render: (_: any, record: AlertRule) => (
        <span>
          {record.targetType === 'host' ? 'Host' : record.targetType === 'cluster' ? 'Cluster' : record.targetType}
          {record.targetId && `: ${record.targetId.substring(0, 8)}`}
        </span>
      ),
    },
    {
      title: 'Condition',
      key: 'condition',
      width: 150,
      render: (_: any, record: AlertRule) => (
        <span>
          {record.metricType} {record.operator} {record.threshold}
        </span>
      ),
    },
    {
      title: 'Severity',
      dataIndex: 'severity',
      key: 'severity',
      width: 100,
      render: (severity: AlertSeverity) => {
        const config = getSeverityConfig(severity)
        return <Tag color={config.color}>{config.label}</Tag>
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      render: () => (
        <Space>
          <Button size="small" icon={<EditOutlined />} />
          <Button size="small" danger icon={<DeleteOutlined />} />
        </Space>
      ),
    },
  ]

  const alerts = alertsData?.data.alerts || []
  const rules = rulesData?.data.rules || []

  const tabItems = [
    {
      key: 'alerts',
      label: `Alerts (${alerts.length})`,
      children: (
        <Table
          columns={alertColumns}
          dataSource={alerts}
          rowKey="id"
          loading={alertsLoading}
          pagination={{
            total: alertsData?.data.total || 0,
            pageSize: alertsData?.data.pageSize || 20,
            current: alertsData?.data.page || 1,
          }}
          size="small"
        />
      ),
    },
    {
      key: 'rules',
      label: `Alert Rules (${rules.length})`,
      children: (
        <Table
          columns={ruleColumns}
          dataSource={rules}
          rowKey="id"
          loading={rulesLoading}
          pagination={{
            total: rulesData?.data.total || 0,
            pageSize: rulesData?.data.pageSize || 20,
            current: rulesData?.data.page || 1,
          }}
          size="small"
        />
      ),
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
            <BellOutlined /> Alerts & Events
          </span>
        </Space>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            Refresh
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setRuleModal({ visible: true })}>
            Create Alert Rule
          </Button>
        </Space>
      </div>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card loading={statsLoading}>
            <Statistic
              title="Total Alerts"
              value={stats?.totalAlerts || 0}
              prefix={<BellOutlined />}
              valueStyle={{ color: '#888' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={statsLoading}>
            <Statistic
              title="Firing"
              value={stats?.firingAlerts || 0}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={statsLoading}>
            <Statistic
              title="Critical"
              value={stats?.criticalAlerts || 0}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={statsLoading}>
            <Statistic
              title="Resolved"
              value={stats?.resolvedAlerts || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Tabs */}
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />

      {/* Create Alert Rule Modal */}
      <Modal
        title="Create Alert Rule"
        open={ruleModal.visible}
        onCancel={() => {
          setRuleModal({ visible: false })
          form.resetFields()
        }}
        onOk={() => {
          form.validateFields().then(async (values) => {
            try {
              await alertApi.createAlertRule(values)
              message.success('Alert rule created successfully')
              setRuleModal({ visible: false })
              form.resetFields()
              refetchRules()
            } catch (error: any) {
              message.error(`Failed to create alert rule: ${error.response?.data?.message || error.message}`)
            }
          })
        }}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="Rule Name"
            name="name"
            rules={[{ required: true, message: 'Please enter rule name' }]}
          >
            <Input placeholder="e.g., High CPU Usage Alert" />
          </Form.Item>

          <Form.Item
            label="Description"
            name="description"
          >
            <TextArea rows={2} placeholder="Optional description" />
          </Form.Item>

          <Form.Item
            label="Target Type"
            name="targetType"
            rules={[{ required: true, message: 'Please select target type' }]}
          >
            <Select>
              <Option value="host">Host</Option>
              <Option value="cluster">Cluster</Option>
              <Option value="node">Node</Option>
              <Option value="pod">Pod</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Metric Type"
            name="metricType"
            rules={[{ required: true, message: 'Please select metric type' }]}
          >
            <Select>
              <Option value="cpu_usage">CPU Usage</Option>
              <Option value="memory_usage">Memory Usage</Option>
              <Option value="disk_usage">Disk Usage</Option>
              <Option value="cluster_cpu_usage">Cluster CPU Usage</Option>
              <Option value="cluster_memory_usage">Cluster Memory Usage</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Operator"
            name="operator"
            rules={[{ required: true, message: 'Please select operator' }]}
          >
            <Select>
              <Option value="greater_than">{`>`}</Option>
              <Option value="less_than">{'<'}</Option>
              <Option value="greater_or_equal">{`>=`}</Option>
              <Option value="less_or_equal">{`<=`}</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Threshold (%)"
            name="threshold"
            rules={[{ required: true, message: 'Please enter threshold' }]}
          >
            <InputNumber min={0} max={100} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            label="Duration (seconds)"
            name="duration"
            rules={[{ required: true, message: 'Please enter duration' }]}
            initialValue={300}
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            label="Severity"
            name="severity"
            rules={[{ required: true, message: 'Please select severity' }]}
            initialValue="warning"
          >
            <Select>
              <Option value="info">Info</Option>
              <Option value="warning">Warning</Option>
              <Option value="critical">Critical</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="Enable Email Notifications"
            name="notifyEmail"
            valuePropName="checked"
            initialValue={false}
          >
            <input type="checkbox" />
          </Form.Item>

          <Form.Item
            label="Enable Webhook Notifications"
            name="notifyWebhook"
            valuePropName="checked"
            initialValue={false}
          >
            <input type="checkbox" />
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prevValues, currentValues) => {
            return prevValues?.notifyWebhook !== currentValues?.notifyWebhook
          }}>
            {({ getFieldValue }) =>
              getFieldValue('notifyWebhook') ? (
                <Form.Item
                  label="Webhook URL"
                  name="webhookUrl"
                >
                  <Input placeholder="https://example.com/webhook" />
                </Form.Item>
              ) : null
            }
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default AlertListPage
