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
} from 'antd'
import {
  PlusOutlined,
  DeleteOutlined,
  RollbackOutlined,
  ArrowUpOutlined,
  RocketOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import helmApi from '../api/helm'

const { Option } = Select
const { TextArea } = Input

export const HelmApplicationPage: React.FC = () => {
  const queryClient = useQueryClient()

  const [isInstallModalOpen, setIsInstallModalOpen] = useState(false)
  const [isHistoryModalOpen, setIsHistoryModalOpen] = useState(false)
  const [selectedReleaseHistory, setSelectedReleaseHistory] = useState<any[]>([])
  const [form] = Form.useForm()

  // Fetch releases
  const { data: releasesData, isLoading, refetch } = useQuery({
    queryKey: ['helmReleases'],
    queryFn: () => helmApi.getReleases({ page: 1, pageSize: 100 }),
  })

  const releases = releasesData?.data || []

  // Install mutation
  const installMutation = useMutation({
    mutationFn: helmApi.installRelease,
    onSuccess: () => {
      message.success('Release installation initiated')
      setIsInstallModalOpen(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['helmReleases'] })
    },
    onError: (error: any) => {
      message.error(`Failed to install release: ${error.response?.data?.message || error.message}`)
    },
  })

  // Upgrade mutation
  const upgradeMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      helmApi.upgradeRelease(id, data),
    onSuccess: () => {
      message.success('Release upgraded successfully')
      queryClient.invalidateQueries({ queryKey: ['helmReleases'] })
    },
    onError: (error: any) => {
      message.error(`Failed to upgrade release: ${error.response?.data?.message || error.message}`)
    },
  })

  // Rollback mutation
  const rollbackMutation = useMutation({
    mutationFn: ({ id, revision }: { id: string; revision: number }) =>
      helmApi.rollbackRelease(id, { revision }),
    onSuccess: () => {
      message.success('Release rolled back successfully')
      queryClient.invalidateQueries({ queryKey: ['helmReleases'] })
    },
    onError: (error: any) => {
      message.error(`Failed to rollback release: ${error.response?.data?.message || error.message}`)
    },
  })

  // Uninstall mutation
  const uninstallMutation = useMutation({
    mutationFn: helmApi.uninstallRelease,
    onSuccess: () => {
      message.success('Release uninstalled successfully')
      queryClient.invalidateQueries({ queryKey: ['helmReleases'] })
    },
    onError: (error: any) => {
      message.error(`Failed to uninstall release: ${error.response?.data?.message || error.message}`)
    },
  })

  // Handle install
  const handleInstall = () => {
    form.resetFields()
    setIsInstallModalOpen(true)
  }

  // Handle submit install
  const handleSubmitInstall = async () => {
    try {
      const values = await form.validateFields()
      installMutation.mutate(values)
    } catch (error) {
      console.error('Validation failed:', error)
    }
  }

  // Handle upgrade
  const handleUpgrade = async (release: any) => {
    Modal.confirm({
      title: 'Upgrade Release',
      content: `Upgrade release "${release.name}" to the latest version?`,
      onOk: async () => {
        upgradeMutation.mutate({
          id: release.id,
          data: {},
        })
      },
    })
  }

  // Handle rollback
  const handleRollback = async (release: any) => {
    // Show history modal
    try {
      const history = await helmApi.getReleaseHistory(release.id)
      setSelectedReleaseHistory(history.history)
      setIsHistoryModalOpen(true)
    } catch (error: any) {
      message.error(`Failed to fetch history: ${error.response?.data?.message || error.message}`)
    }
  }

  // Handle rollback to specific revision
  const handleRollbackToRevision = async (release: any, revision: number) => {
    setIsHistoryModalOpen(false)
    rollbackMutation.mutate({
      id: release.id,
      revision: revision,
    })
  }

  // Handle uninstall
  const handleUninstall = async (release: any) => {
    uninstallMutation.mutate(release.id)
  }

  // Get status tag
  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      deployed: { color: 'success', text: 'Deployed' },
      pending: { color: 'processing', text: 'Pending' },
      'pending-upgrade': { color: 'processing', text: 'Pending Upgrade' },
      'pending-rollback': { color: 'processing', text: 'Pending Rollback' },
      superseded: { color: 'default', text: 'Superseded' },
      failed: { color: 'error', text: 'Failed' },
      unknown: { color: 'default', text: 'Unknown' },
      uninstalling: { color: 'warning', text: 'Uninstalling' },
    }
    const { color, text } = statusMap[status] || { color: 'default', text: status }
    return <Tag color={color}>{text}</Tag>
  }

  // Table columns
  const columns: ColumnsType<any> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <RocketOutlined />
          <span style={{ fontWeight: 'bold' }}>{name}</span>
        </Space>
      ),
    },
    {
      title: 'Namespace',
      dataIndex: 'namespace',
      key: 'namespace',
      render: (namespace: string) => <Tag>{namespace}</Tag>,
    },
    {
      title: 'Cluster',
      key: 'cluster',
      render: (_: any, record: any) => record.cluster?.name || '-',
    },
    {
      title: 'Chart',
      dataIndex: 'chart',
      key: 'chart',
      render: (chart: string, record: any) => (
        <span>
          {chart}
          {record.chartVersion && <Tag style={{ marginLeft: 4 }}>{record.chartVersion}</Tag>}
        </span>
      ),
    },
    {
      title: 'Revision',
      dataIndex: 'revision',
      key: 'revision',
      width: 80,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => getStatusTag(status),
    },
    {
      title: 'Updated',
      dataIndex: 'updated',
      key: 'updated',
      width: 150,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (_: any, record: any) => (
        <Space size="small">
          <Tooltip title="Upgrade">
            <Button
              size="small"
              icon={<ArrowUpOutlined />}
              onClick={() => handleUpgrade(record)}
              loading={upgradeMutation.isPending}
            />
          </Tooltip>
          <Tooltip title="Rollback">
            <Button
              size="small"
              icon={<RollbackOutlined />}
              onClick={() => handleRollback(record)}
            />
          </Tooltip>
          <Popconfirm
            title="Uninstall Release"
            description="Are you sure you want to uninstall this release?"
            onConfirm={() => handleUninstall(record)}
            okText="Yes"
            cancelText="No"
          >
            <Tooltip title="Uninstall">
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
            <RocketOutlined /> Helm Applications
          </h2>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => refetch()} loading={isLoading}>
              Refresh
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleInstall}>
              Install Chart
            </Button>
          </Space>
        </div>
      </Card>

      {/* Table */}
      <Card>
        <Table
          columns={columns}
          dataSource={releases}
          rowKey="id"
          loading={isLoading}
          pagination={{
            total: releasesData?.total || 0,
            pageSize: releasesData?.pageSize || 20,
            current: releasesData?.page || 1,
            showSizeChanger: false,
          }}
        />
      </Card>

      {/* Install Modal */}
      <Modal
        title="Install Helm Chart"
        open={isInstallModalOpen}
        onOk={handleSubmitInstall}
        onCancel={() => {
          setIsInstallModalOpen(false)
          form.resetFields()
        }}
        confirmLoading={installMutation.isPending}
        width={600}
      >
        <Form form={form} layout="vertical">
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
            label="Namespace"
            name="namespace"
            rules={[{ required: true, message: 'Please enter a namespace' }]}
            initialValue="default"
          >
            <Input placeholder="e.g., default" />
          </Form.Item>

          <Form.Item
            label="Release Name"
            name="name"
            rules={[{ required: true, message: 'Please enter a release name' }]}
          >
            <Input placeholder="e.g., my-app" />
          </Form.Item>

          <Form.Item
            label="Chart"
            name="chart"
            rules={[{ required: true, message: 'Please enter a chart name' }]}
          >
            <Input placeholder="e.g., stable/nginx" />
          </Form.Item>

          <Form.Item label="Chart Version" name="chartVersion">
            <Input placeholder="Leave empty for latest version" />
          </Form.Item>

          <Form.Item label="Values (YAML)" name="values">
            <TextArea
              rows={8}
              placeholder="Enter custom values in YAML format"
              style={{ fontFamily: 'monospace' }}
            />
          </Form.Item>

          <Form.Item label="Description" name="description">
            <TextArea rows={2} placeholder="Release description" />
          </Form.Item>
        </Form>
      </Modal>

      {/* History Modal */}
      <Modal
        title="Release History"
        open={isHistoryModalOpen}
        onCancel={() => setIsHistoryModalOpen(false)}
        footer={null}
        width={800}
      >
        <Table
          columns={[
            { title: 'Revision', dataIndex: 'revision', key: 'revision', width: 80 },
            { title: 'Updated', dataIndex: 'updated', key: 'updated', width: 150 },
            { title: 'Status', dataIndex: 'status', key: 'status', render: (s: string) => getStatusTag(s) },
            { title: 'Chart', dataIndex: 'chart', key: 'chart' },
            { title: 'Version', dataIndex: 'chartVersion', key: 'chartVersion', width: 100 },
            {
              title: 'Actions',
              key: 'actions',
              width: 100,
              render: (_: any, record: any) => (
                <Button
                  size="small"
                  type="primary"
                  onClick={() => handleRollbackToRevision(selectedReleaseHistory[0], record.revision)}
                >
                  Rollback
                </Button>
              ),
            },
          ]}
          dataSource={selectedReleaseHistory}
          rowKey="revision"
          pagination={false}
          size="small"
        />
      </Modal>
    </div>
  )
}

export default HelmApplicationPage
