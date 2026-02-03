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
  Switch,
  message,
  Popconfirm,
  Tag,
  Card,
  Tooltip,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SyncOutlined,
  CloudServerOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import helmApi, { type HelmRepository, type UpdateHelmRepoRequest } from '../api/helm'

const { Option } = Select
const { TextArea } = Input

export const HelmRepositoryPage: React.FC = () => {
  const queryClient = useQueryClient()

  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingRepo, setEditingRepo] = useState<HelmRepository | null>(null)
  const [form] = Form.useForm()

  // Fetch repositories
  const { data: reposData, isLoading, refetch } = useQuery({
    queryKey: ['helmRepos'],
    queryFn: () => helmApi.getRepositories({ page: 1, pageSize: 100 }),
  })

  const repositories = reposData?.data || []

  // Create mutation
  const createMutation = useMutation({
    mutationFn: helmApi.createRepository,
    onSuccess: () => {
      message.success('Repository created successfully')
      setIsModalOpen(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['helmRepos'] })
    },
    onError: (error: any) => {
      message.error(`Failed to create repository: ${error.response?.data?.message || error.message}`)
    },
  })

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateHelmRepoRequest }) =>
      helmApi.updateRepository(id, data),
    onSuccess: () => {
      message.success('Repository updated successfully')
      setIsModalOpen(false)
      setEditingRepo(null)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['helmRepos'] })
    },
    onError: (error: any) => {
      message.error(`Failed to update repository: ${error.response?.data?.message || error.message}`)
    },
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: helmApi.deleteRepository,
    onSuccess: () => {
      message.success('Repository deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['helmRepos'] })
    },
    onError: (error: any) => {
      message.error(`Failed to delete repository: ${error.response?.data?.message || error.message}`)
    },
  })

  // Sync mutation
  const syncMutation = useMutation({
    mutationFn: helmApi.syncRepository,
    onSuccess: () => {
      message.success('Repository synced successfully')
      queryClient.invalidateQueries({ queryKey: ['helmRepos'] })
    },
    onError: (error: any) => {
      message.error(`Failed to sync repository: ${error.response?.data?.message || error.message}`)
    },
  })

  // Handle add
  const handleAdd = () => {
    setEditingRepo(null)
    form.resetFields()
    setIsModalOpen(true)
  }

  // Handle edit
  const handleEdit = (repo: HelmRepository) => {
    setEditingRepo(repo)
    form.setFieldsValue({
      name: repo.name,
      description: repo.description,
      type: repo.type,
      url: repo.url,
      username: repo.username,
      insecureSkipTLS: repo.insecureSkipTLS,
    })
    setIsModalOpen(true)
  }

  // Handle delete
  const handleDelete = async (id: string) => {
    deleteMutation.mutate(id)
  }

  // Handle sync
  const handleSync = async (id: string) => {
    syncMutation.mutate(id)
  }

  // Handle submit
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()

      if (editingRepo) {
        updateMutation.mutate({
          id: editingRepo.id,
          data: values,
        })
      } else {
        createMutation.mutate(values)
      }
    } catch (error) {
      console.error('Validation failed:', error)
    }
  }

  // Get status tag
  const getStatusTag = (repo: HelmRepository) => {
    const statusMap = {
      active: { color: 'success', text: 'Active' },
      inactive: { color: 'default', text: 'Inactive' },
      error: { color: 'error', text: 'Error' },
    }
    const { color, text } = statusMap[repo.status] || { color: 'default', text: repo.status }
    return <Tag color={color}>{text}</Tag>
  }

  // Table columns
  const columns: ColumnsType<HelmRepository> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record) => (
        <Space>
          <CloudServerOutlined />
          <span>
            {name}
            {record.description && (
              <Tooltip title={record.description}>
                <span style={{ color: '#999', marginLeft: 4 }}>(...)</span>
              </Tooltip>
            )}
          </span>
        </Space>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag color="blue">{type.toUpperCase()}</Tag>,
      width: 100,
    },
    {
      title: 'URL',
      dataIndex: 'url',
      key: 'url',
      ellipsis: true,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (_: any, record: HelmRepository) => getStatusTag(record),
    },
    {
      title: 'Charts',
      dataIndex: 'chartCount',
      key: 'chartCount',
      width: 80,
    },
    {
      title: 'Last Sync',
      key: 'lastSync',
      width: 180,
      render: (_: any, record: HelmRepository) => {
        if (!record.lastSyncedAt) return '-'
        const date = new Date(record.lastSyncedAt)
        return <span style={{ fontSize: '12px' }}>{date.toLocaleString()}</span>
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (_: any, record: HelmRepository) => (
        <Space size="small">
          <Tooltip title="Sync">
            <Button
              size="small"
              icon={<SyncOutlined spin={syncMutation.isPending} />}
              onClick={() => handleSync(record.id)}
              loading={syncMutation.isPending}
            />
          </Tooltip>
          <Tooltip title="Edit">
            <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          </Tooltip>
          <Popconfirm
            title="Delete Repository"
            description="Are you sure you want to delete this repository?"
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
            <CloudServerOutlined /> Helm Repositories
          </h2>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => refetch()} loading={isLoading}>
              Refresh
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
              Add Repository
            </Button>
          </Space>
        </div>
      </Card>

      {/* Table */}
      <Card>
        <Table
          columns={columns}
          dataSource={repositories}
          rowKey="id"
          loading={isLoading}
          pagination={{
            total: reposData?.total || 0,
            pageSize: reposData?.pageSize || 20,
            current: reposData?.page || 1,
            showSizeChanger: false,
          }}
        />
      </Card>

      {/* Add/Edit Modal */}
      <Modal
        title={editingRepo ? 'Edit Repository' : 'Add Repository'}
        open={isModalOpen}
        onOk={handleSubmit}
        onCancel={() => {
          setIsModalOpen(false)
          setEditingRepo(null)
          form.resetFields()
        }}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="Name"
            name="name"
            rules={[{ required: true, message: 'Please enter a name' }]}
          >
            <Input placeholder="e.g., stable" />
          </Form.Item>

          <Form.Item label="Description" name="description">
            <TextArea rows={2} placeholder="Repository description" />
          </Form.Item>

          <Form.Item
            label="Type"
            name="type"
            rules={[{ required: true, message: 'Please select a type' }]}
            initialValue="https"
          >
            <Select>
              <Option value="http">HTTP</Option>
              <Option value="https">HTTPS</Option>
              <Option value="oci">OCI</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="URL"
            name="url"
            rules={[{ required: true, message: 'Please enter a URL' }]}
          >
            <Input placeholder="e.g., https://charts.helm.sh/stable" />
          </Form.Item>

          <Form.Item label="Username" name="username">
            <Input placeholder="Username (optional)" />
          </Form.Item>

          <Form.Item label="Password" name="password">
            <Input.Password placeholder="Password (optional)" />
          </Form.Item>

          <Form.Item
            label="Insecure Skip TLS"
            name="insecureSkipTLS"
            valuePropName="checked"
            initialValue={false}
          >
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default HelmRepositoryPage
