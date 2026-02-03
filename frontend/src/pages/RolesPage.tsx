import React, { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Button,
  Table,
  Space,
  Modal,
  Form,
  Input,
  message,
  Popconfirm,
  Tag,
  Card,
  Tooltip,
  Switch,
  Empty,
  Row,
  Col,
  Statistic,
  Transfer,
  TransferProps,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  SafetyOutlined,
  KeyOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import rbacApi, { type Role, type Permission } from '../api/rbac'

const { TextArea } = Input

export const RolesPage: React.FC = () => {
  const queryClient = useQueryClient()

  const [isRoleModalOpen, setIsRoleModalOpen] = useState(false)
  const [isPermissionModalOpen, setIsPermissionModalOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [selectedRole, setSelectedRole] = useState<Role | null>(null)
  const [form] = Form.useForm()
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>([])

  // Fetch roles
  const { data: rolesData, isLoading, refetch } = useQuery({
    queryKey: ['roles'],
    queryFn: () => rbacApi.getRoles({ page: 1, pageSize: 100 }),
  })

  // Fetch permissions for assignment
  const { data: permissionsData } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => rbacApi.getPermissions({ page: 1, pageSize: 1000 }),
  })

  const roles = rolesData?.data.data || []
  const allPermissions = permissionsData?.data.data || []

  // Create mutation
  const createMutation = useMutation({
    mutationFn: rbacApi.createRole,
    onSuccess: () => {
      message.success('Role created successfully')
      setIsRoleModalOpen(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['roles'] })
    },
    onError: (error: any) => {
      message.error(`Failed to create role: ${error.response?.data?.error || error.message}`)
    },
  })

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      rbacApi.updateRole(id, data),
    onSuccess: () => {
      message.success('Role updated successfully')
      setIsRoleModalOpen(false)
      setEditingRole(null)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['roles'] })
    },
    onError: (error: any) => {
      message.error(`Failed to update role: ${error.response?.data?.error || error.message}`)
    },
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: rbacApi.deleteRole,
    onSuccess: () => {
      message.success('Role deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['roles'] })
    },
    onError: (error: any) => {
      message.error(`Failed to delete role: ${error.response?.data?.error || error.message}`)
    },
  })

  // Assign permissions mutation
  const assignPermissionsMutation = useMutation({
    mutationFn: ({ roleId, permissionIds }: { roleId: string; permissionIds: string[] }) =>
      rbacApi.assignRolePermissions(roleId, { permissionIds, override: true }),
    onSuccess: () => {
      message.success('Permissions assigned successfully')
      setIsPermissionModalOpen(false)
      setSelectedPermissions([])
      queryClient.invalidateQueries({ queryKey: ['roles'] })
    },
    onError: (error: any) => {
      message.error(`Failed to assign permissions: ${error.response?.data?.error || error.message}`)
    },
  })

  // Seed default roles
  const seedMutation = useMutation({
    mutationFn: rbacApi.seedDefaultRoles,
    onSuccess: () => {
      message.success('Default roles and permissions seeded successfully')
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      queryClient.invalidateQueries({ queryKey: ['permissions'] })
    },
    onError: (error: any) => {
      message.error(`Failed to seed defaults: ${error.response?.data?.error || error.message}`)
    },
  })

  // Handle add
  const handleAdd = () => {
    setEditingRole(null)
    form.resetFields()
    form.setFieldsValue({
      isDefault: false,
    })
    setIsRoleModalOpen(true)
  }

  // Handle edit
  const handleEdit = (role: Role) => {
    setEditingRole(role)
    form.setFieldsValue({
      name: role.name,
      displayName: role.displayName,
      description: role.description,
      isDefault: role.isDefault,
    })
    setIsRoleModalOpen(true)
  }

  // Handle delete
  const handleDelete = async (id: string) => {
    deleteMutation.mutate(id)
  }

  // Handle manage permissions
  const handleManagePermissions = async (role: Role) => {
    setSelectedRole(role)
    // Get role permissions
    const result = await rbacApi.getRolePermissions(role.id)
    setSelectedPermissions(result.data.data.map((p: Permission) => p.id))
    setIsPermissionModalOpen(true)
  }

  // Calculate statistics
  const totalRoles = roles.length
  const systemRoles = roles.filter(r => r.isSystem).length
  const customRoles = totalRoles - systemRoles
  const totalPermissions = allPermissions.length

  // Table columns
  const columns: ColumnsType<Role> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (_name: string, record: Role) => (
        <Space>
          <SafetyOutlined style={{ color: record.isSystem ? '#faad14' : '#1890ff' }} />
          <span style={{ fontWeight: record.isSystem ? 'bold' : 'normal' }}>
            {record.displayName}
          </span>
          {record.isDefault && <Tag color="blue">Default</Tag>}
        </Space>
      ),
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: 'Permissions',
      dataIndex: 'permissions',
      key: 'permissions',
      width: 150,
      render: (permissions: Permission[] | undefined) => (
        <Tag color={permissions && permissions.length > 0 ? 'green' : 'default'}>
          {permissions?.length || 0}
        </Tag>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'isSystem',
      key: 'isSystem',
      width: 100,
      render: (isSystem: boolean) => (
        <Tag color={isSystem ? 'orange' : 'blue'}>
          {isSystem ? 'System' : 'Custom'}
        </Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (_: any, record: Role) => (
        <Space size="small">
          <Tooltip title="Manage Permissions">
            <Button
              size="small"
              type="primary"
              icon={<KeyOutlined />}
              onClick={() => handleManagePermissions(record)}
            />
          </Tooltip>
          <Tooltip title="Edit">
            <Button
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
              disabled={record.isSystem}
            />
          </Tooltip>
          <Popconfirm
            title="Delete Role"
            description="Are you sure you want to delete this role?"
            onConfirm={() => handleDelete(record.id)}
            okText="Yes"
            cancelText="No"
            disabled={record.isSystem}
          >
            <Tooltip title="Delete">
              <Button size="small" danger icon={<DeleteOutlined />} disabled={record.isSystem} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  // Permission transfer data source
  const transferDataSource: TransferProps['dataSource'] = allPermissions.map((perm) => ({
    key: perm.id,
    title: (
      <div>
        <Tag color="blue" style={{ marginRight: 8 }}>{perm.resource}</Tag>
        <span>{perm.displayName}</span>
        <Tag style={{ marginLeft: 8 }}>{perm.action}</Tag>
      </div>
    ),
    description: perm.name,
    category: perm.category,
  }))

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <Card style={{ marginBottom: '16px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2 style={{ margin: 0 }}>
            <SafetyOutlined /> Role-Based Access Control
          </h2>
          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => refetch()}
              loading={isLoading}
            >
              Refresh
            </Button>
            <Button
              onClick={() => seedMutation.mutate()}
              loading={seedMutation.isPending}
            >
              Seed Defaults
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
              Add Role
            </Button>
          </Space>
        </div>
      </Card>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '16px' }}>
        <Col span={6}>
          <Card>
            <Statistic title="Total Roles" value={totalRoles} prefix={<SafetyOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="System Roles"
              value={systemRoles}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Custom Roles"
              value={customRoles}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="Total Permissions" value={totalPermissions} prefix={<KeyOutlined />} />
          </Card>
        </Col>
      </Row>

      {/* Roles Table */}
      <Card title="Roles">
        {roles.length === 0 && !isLoading ? (
          <Empty
            description="No roles configured. Add your first role or seed default roles."
          />
        ) : (
          <Table
            columns={columns}
            dataSource={roles}
            rowKey="id"
            loading={isLoading}
            pagination={{
              total: rolesData?.data.total || 0,
              pageSize: rolesData?.data.pageSize || 20,
              current: rolesData?.data.page || 1,
              showSizeChanger: false,
            }}
          />
        )}
      </Card>

      {/* Add/Edit Role Modal */}
      <Modal
        title={editingRole ? 'Edit Role' : 'Add Role'}
        open={isRoleModalOpen}
        onOk={() => form.submit()}
        onCancel={() => {
          setIsRoleModalOpen(false)
          setEditingRole(null)
          form.resetFields()
        }}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={async (values) => {
          if (editingRole) {
            updateMutation.mutate({
              id: editingRole.id,
              data: values,
            })
          } else {
            createMutation.mutate(values)
          }
        }}>
          <Form.Item
            label="Role Name"
            name="name"
            rules={[{ required: true, message: 'Please enter a role name' }]}
          >
            <Input placeholder="e.g., operator" disabled={editingRole?.isSystem} />
          </Form.Item>

          <Form.Item
            label="Display Name"
            name="displayName"
            rules={[{ required: true, message: 'Please enter a display name' }]}
          >
            <Input placeholder="e.g., Operator" />
          </Form.Item>

          <Form.Item label="Description" name="description">
            <TextArea rows={3} placeholder="Describe this role's purpose" />
          </Form.Item>

          <Form.Item
            label="Default Role"
            name="isDefault"
            valuePropName="checked"
            tooltip="New users will be assigned this role by default"
          >
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      {/* Manage Permissions Modal */}
      <Modal
        title={`Manage Permissions - ${selectedRole?.displayName || 'Role'}`}
        open={isPermissionModalOpen}
        onOk={() => {
          if (selectedRole) {
            assignPermissionsMutation.mutate({
              roleId: selectedRole.id,
              permissionIds: selectedPermissions,
            })
          }
        }}
        onCancel={() => {
          setIsPermissionModalOpen(false)
          setSelectedRole(null)
          setSelectedPermissions([])
        }}
        confirmLoading={assignPermissionsMutation.isPending}
        width={800}
      >
        <div style={{ marginBottom: 16 }}>
          <p>Assign permissions to this role. System roles cannot be modified.</p>
        </div>
        <Transfer
          dataSource={transferDataSource}
          titles={['Available', 'Assigned']}
          targetKeys={selectedPermissions}
          onChange={(keys) => setSelectedPermissions(keys as string[])}
          render={item => item.title as any}
          listStyle={{
            width: 350,
            height: 400,
          }}
          showSearch
          filterOption={(inputValue, item: any) =>
            item.title?.props?.children?.[1]?.toLowerCase().includes(inputValue.toLowerCase()) ||
            item.description?.toLowerCase().includes(inputValue.toLowerCase())
          }
        />
      </Modal>
    </div>
  )
}

export default RolesPage
