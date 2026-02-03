import React, { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { Button, Space, Table, Tag, Card, Modal, Form, Input, Select, Switch, message, Popconfirm, Row, Col, Tabs, Descriptions } from 'antd'
import {
  UserOutlined,
  EditOutlined,
  DeleteOutlined,
  SafetyOutlined,
  KeyOutlined,
  ReloadOutlined,
  TeamOutlined,
  LockOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { userApi } from '../api/user'
import type { User, Role, UserRole } from '../types/user'

const { Option } = Select

export const UserManagementPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState('users')
  const [userModal, setUserModal] = useState<{ visible: boolean; user?: User }>({ visible: false })
  const [roleModal, setRoleModal] = useState<{ visible: boolean; role?: Role }>({ visible: false })
  const [permissionModal, setPermissionModal] = useState<{ visible: boolean; role?: Role }>({ visible: false })

  const [userForm] = Form.useForm()
  const [roleForm] = Form.useForm()

  // Fetch users
  const { data: usersData, isLoading: usersLoading, refetch: refetchUsers } = useQuery({
    queryKey: ['users'],
    queryFn: () => userApi.getUsers(),
  })

  // Fetch roles
  const { data: rolesData, isLoading: rolesLoading, refetch: refetchRoles } = useQuery({
    queryKey: ['roles'],
    queryFn: () => userApi.getRoles(),
  })

  // Fetch permissions
  const { data: permissionsData } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => userApi.getPermissions(),
  })

  // Create user mutation
  const createUserMutation = useMutation({
    mutationFn: (user: Partial<User>) => userApi.createUser(user),
    onSuccess: () => {
      message.success('User created successfully')
      setUserModal({ visible: false })
      userForm.resetFields()
      refetchUsers()
    },
  })

  // Update user mutation
  const updateUserMutation = useMutation({
    mutationFn: ({ userId, user }: { userId: string; user: Partial<User> }) =>
      userApi.updateUser(userId, user),
    onSuccess: () => {
      message.success('User updated successfully')
      setUserModal({ visible: false })
      refetchUsers()
    },
  })

  // Delete user mutation
  const deleteUserMutation = useMutation({
    mutationFn: (userId: string) => userApi.deleteUser(userId),
    onSuccess: () => {
      message.success('User deleted successfully')
      refetchUsers()
    },
  })

  // Create role mutation
  const createRoleMutation = useMutation({
    mutationFn: (role: Partial<Role>) => userApi.createRole(role),
    onSuccess: () => {
      message.success('Role created successfully')
      setRoleModal({ visible: false })
      roleForm.resetFields()
      refetchRoles()
    },
  })

  // Update role mutation
  const updateRoleMutation = useMutation({
    mutationFn: ({ roleId, role }: { roleId: string; role: Partial<Role> }) =>
      userApi.updateRole(roleId, role),
    onSuccess: () => {
      message.success('Role updated successfully')
      setRoleModal({ visible: false })
      refetchRoles()
    },
  })

  // Delete role mutation
  const deleteRoleMutation = useMutation({
    mutationFn: (roleId: string) => userApi.deleteRole(roleId),
    onSuccess: () => {
      message.success('Role deleted successfully')
      refetchRoles()
    },
  })

  const users = usersData?.data.users || []
  const roles = rolesData?.data.roles || []
  const permissions = permissionsData?.data.permissions || []

  // User table columns
  const userColumns: ColumnsType<User> = [
    {
      title: 'User',
      dataIndex: 'username',
      key: 'username',
      width: 150,
      render: (username: string, record: User) => (
        <Space>
          <UserOutlined />
          {username}
          {record.userType === 'ldap' && <Tag color="blue">LDAP</Tag>}
        </Space>
      ),
    },
    {
      title: 'Display Name',
      dataIndex: 'displayName',
      key: 'displayName',
      width: 150,
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
      width: 200,
    },
    {
      title: 'Department',
      dataIndex: 'department',
      key: 'department',
      width: 150,
    },
    {
      title: 'Position',
      dataIndex: 'position',
      key: 'position',
      width: 150,
    },
    {
      title: 'Roles',
      dataIndex: 'roles',
      key: 'roles',
      width: 200,
      render: (roles: Role[]) => (
        <Space direction="vertical" size="small">
          {roles?.map((role) => (
            <Tag key={role.id} color={getRoleColor(role.name)}>{role.displayName}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'isActive',
      key: 'isActive',
      width: 100,
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'success' : 'default'}>{isActive ? 'Active' : 'Inactive'}</Tag>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 150,
      render: (_: any, record: User) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => setUserModal({ visible: true, user: record })}>
            Edit
          </Button>
          <Popconfirm
            title="Are you sure you want to delete this user?"
            onConfirm={() => deleteUserMutation.mutate(record.id)}
            okText="Yes"
            cancelText="No"
          >
            <Button size="small" danger icon={<DeleteOutlined />} loading={deleteUserMutation.isPending}>
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  // Role table columns
  const roleColumns: ColumnsType<Role> = [
    {
      title: 'Name',
      dataIndex: 'displayName',
      key: 'displayName',
      width: 200,
      render: (displayName: string, record: Role) => (
        <Space>
          {displayName}
          {record.isSystem && <Tag color="blue">System</Tag>}
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
      key: 'permissions',
      width: 150,
      render: (_: any, record: Role) => (
        <Tag color="blue">{record.permissions?.length || 0} permissions</Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 150,
      render: (_: any, record: Role) => (
        <Space>
          <Button
            size="small"
            icon={<KeyOutlined />}
            onClick={() => setPermissionModal({ visible: true, role: record })}
            disabled={record.isSystem}
          >
            Permissions
          </Button>
          {!record.isSystem && (
            <>
              <Button size="small" icon={<EditOutlined />} onClick={() => setRoleModal({ visible: true, role: record })}>
                Edit
              </Button>
              <Popconfirm
                title="Are you sure you want to delete this role?"
                onConfirm={() => deleteRoleMutation.mutate(record.id)}
                okText="Yes"
                cancelText="No"
              >
                <Button size="small" danger icon={<DeleteOutlined />} loading={deleteRoleMutation.isPending}>
                  Delete
                </Button>
              </Popconfirm>
            </>
          )}
        </Space>
      ),
    },
  ]

  const getRoleColor = (roleName: UserRole) => {
    switch (roleName) {
      case 'admin':
        return 'red'
      case 'operator':
        return 'orange'
      case 'viewer':
        return 'green'
      case 'auditor':
        return 'purple'
      default:
        return 'default'
    }
  }

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
            <TeamOutlined /> User & Role Management
          </span>
        </Space>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => { refetchUsers(); refetchRoles() }}>
            Refresh
          </Button>
        </Space>
      </div>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card loading={usersLoading}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <TeamOutlined style={{ fontSize: '24px', color: '#1890ff' }} />
              <div>
                <div style={{ color: '#888' }}>Total Users</div>
                <div style={{ fontSize: '24px', fontWeight: 'bold' }}>{users.length}</div>
              </div>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={rolesLoading}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <SafetyOutlined style={{ fontSize: '24px', color: '#52c41a' }} />
              <div>
                <div style={{ color: '#888' }}>Total Roles</div>
                <div style={{ fontSize: '24px', fontWeight: 'bold' }}>{roles.length}</div>
              </div>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <LockOutlined style={{ fontSize: '24px', color: '#faad14' }} />
              <div>
                <div style={{ color: '#888' }}>Total Permissions</div>
                <div style={{ fontSize: '24px', fontWeight: 'bold' }}>{permissions.length}</div>
              </div>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <UserOutlined style={{ fontSize: '24px', color: '#722ed1' }} />
              <div>
                <div style={{ color: '#888' }}>Active Users</div>
                <div style={{ fontSize: '24px', fontWeight: 'bold' }}>
                  {users.filter((u) => u.isActive).length}
                </div>
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      {/* Tabs */}
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={[
        {
          key: 'users',
          label: `Users (${users.length})`,
          children: (
            <Table
              columns={userColumns}
              dataSource={users}
              rowKey="id"
              loading={usersLoading}
              pagination={{
                total: usersData?.data.total || 0,
                pageSize: usersData?.data.pageSize || 20,
                current: usersData?.data.page || 1,
              }}
            />
          ),
        },
        {
          key: 'roles',
          label: `Roles (${roles.length})`,
          children: (
            <Table
              columns={roleColumns}
              dataSource={roles}
              rowKey="id"
              loading={rolesLoading}
              pagination={false}
            />
          ),
        },
      ]} />

      {/* Create/Edit User Modal */}
      <Modal
        title={userModal.user ? 'Edit User' : 'Create User'}
        open={userModal.visible}
        onCancel={() => setUserModal({ visible: false })}
        onOk={() => {
          userForm.validateFields().then((values) => {
            if (userModal.user) {
              updateUserMutation.mutate({ userId: userModal.user.id, user: values })
            } else {
              createUserMutation.mutate(values)
            }
          })
        }}
        width={600}
      >
        <Form form={userForm} layout="vertical" initialValues={userModal.user}>
          <Form.Item
            label="Username"
            name="username"
            rules={[{ required: true, message: 'Please enter username' }]}
          >
            <Input placeholder="Enter username" disabled={!!userModal.user} />
          </Form.Item>

          <Form.Item
            label="Email"
            name="email"
            rules={[
              { required: true, message: 'Please enter email' },
              { type: 'email', message: 'Please enter valid email' },
            ]}
          >
            <Input placeholder="Enter email" />
          </Form.Item>

          <Form.Item label="Display Name" name="displayName">
            <Input placeholder="Enter display name" />
          </Form.Item>

          <Form.Item label="Department" name="department">
            <Input placeholder="Enter department" />
          </Form.Item>

          <Form.Item label="Position" name="position">
            <Input placeholder="Enter position" />
          </Form.Item>

          <Form.Item label="User Type" name="userType" initialValue="local">
            <Select>
              <Option value="local">Local</Option>
              <Option value="ldap">LDAP</Option>
            </Select>
          </Form.Item>

          <Form.Item label="Status" name="isActive" valuePropName="checked" initialValue={true}>
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      {/* Create/Edit Role Modal */}
      <Modal
        title={roleModal.role ? 'Edit Role' : 'Create Role'}
        open={roleModal.visible}
        onCancel={() => setRoleModal({ visible: false })}
        onOk={() => {
          roleForm.validateFields().then((values) => {
            if (roleModal.role) {
              updateRoleMutation.mutate({ roleId: roleModal.role.id, role: values })
            } else {
              createRoleMutation.mutate(values)
            }
          })
        }}
        width={600}
      >
        <Form form={roleForm} layout="vertical" initialValues={roleModal.role}>
          <Form.Item
            label="Role Name"
            name="name"
            rules={[{ required: true, message: 'Please enter role name' }]}
          >
            <Input placeholder="e.g., viewer, operator" disabled={!!roleModal.role} />
          </Form.Item>

          <Form.Item
            label="Display Name"
            name="displayName"
            rules={[{ required: true, message: 'Please enter display name' }]}
          >
            <Input placeholder="e.g., Viewer, Operator" />
          </Form.Item>

          <Form.Item label="Description" name="description">
            <Input.TextArea rows={3} placeholder="Enter role description" />
          </Form.Item>
        </Form>
      </Modal>

      {/* Role Permissions Modal */}
      <Modal
        title={`Manage Permissions - ${permissionModal.role?.displayName}`}
        open={permissionModal.visible}
        onCancel={() => setPermissionModal({ visible: false })}
        footer={[
          <Button key="cancel" onClick={() => setPermissionModal({ visible: false })}>
            Close
          </Button>,
        ]}
        width={700}
      >
        {permissionModal.role && (
          <div>
            <Descriptions title="Role Info" column={1} bordered size="small" style={{ marginBottom: '16px' }}>
              <Descriptions.Item label="Name">{permissionModal.role.displayName}</Descriptions.Item>
              <Descriptions.Item label="System Role">{permissionModal.role.isSystem ? 'Yes' : 'No'}</Descriptions.Item>
            </Descriptions>

            <div style={{ marginBottom: '16px' }}>
              <strong>Assigned Permissions ({permissionModal.role.permissions?.length || 0}):</strong>
            </div>
            <Table
              columns={[
                { title: 'Resource', dataIndex: 'resource', key: 'resource', width: 150 },
                { title: 'Action', dataIndex: 'action', key: 'action', width: 150 },
                { title: 'Description', dataIndex: 'description', key: 'description', ellipsis: true },
              ]}
              dataSource={permissionModal.role.permissions}
              rowKey="id"
              pagination={false}
              size="small"
            />
          </div>
        )}
      </Modal>
    </div>
  )
}

export default UserManagementPage
