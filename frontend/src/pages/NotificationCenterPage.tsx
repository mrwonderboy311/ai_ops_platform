import React, { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, Space, List, Tag, Card, Badge, Drawer, Switch, Form, Select, Empty, Popconfirm } from 'antd'
import {
  BellOutlined,
  DeleteOutlined,
  CheckOutlined,
  EyeOutlined,
  SettingOutlined,
  ReloadOutlined,
  AlertOutlined,
  InfoCircleOutlined,
  SecurityScanOutlined,
  SyncOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import { notificationApi } from '../api/notification'
import type { NotificationType, NotificationPriority } from '../types/notification'

dayjs.extend(relativeTime)

const { Option } = Select

export const NotificationCenterPage: React.FC = () => {
  const queryClient = useQueryClient()
  const [settingsVisible, setSettingsVisible] = useState(false)
  const [form] = Form.useForm()

  // Fetch notifications
  const { data: notificationsData, isLoading } = useQuery({
    queryKey: ['notifications'],
    queryFn: () => notificationApi.getNotifications({ limit: 50 }),
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  // Fetch unread count
  const { data: unreadData } = useQuery({
    queryKey: ['unreadCount'],
    queryFn: () => notificationApi.getUnreadCount(),
    refetchInterval: 15000, // Refresh every 15 seconds
  })

  // Fetch notification stats
  const { data: statsData } = useQuery({
    queryKey: ['notificationStats'],
    queryFn: () => notificationApi.getNotificationStats(),
  })

  // Fetch preferences
  const { data: prefData } = useQuery({
    queryKey: ['notificationPreference'],
    queryFn: () => notificationApi.getNotificationPreference(),
  })

  // Mark as read mutation
  const markAsReadMutation = useMutation({
    mutationFn: (id: string) => notificationApi.markAsRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['unreadCount'] })
      queryClient.invalidateQueries({ queryKey: ['notificationStats'] })
    },
  })

  // Mark all as read mutation
  const markAllAsReadMutation = useMutation({
    mutationFn: () => notificationApi.markAllAsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['unreadCount'] })
      queryClient.invalidateQueries({ queryKey: ['notificationStats'] })
    },
  })

  // Delete notification mutation
  const deleteNotificationMutation = useMutation({
    mutationFn: (id: string) => notificationApi.deleteNotification(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['unreadCount'] })
      queryClient.invalidateQueries({ queryKey: ['notificationStats'] })
    },
  })

  const notifications = notificationsData?.data.notifications || []
  const unreadCount = unreadData?.data.unreadCount || 0
  const stats = statsData?.data
  const preference = prefData?.data

  // Get icon and color for notification type
  const getTypeConfig = (type: NotificationType) => {
    switch (type) {
      case 'alert':
        return { icon: <AlertOutlined />, color: '#ff4d4f', bgColor: '#fff1f0' }
      case 'system':
        return { icon: <InfoCircleOutlined />, color: '#1890ff', bgColor: '#e6f7ff' }
      case 'task':
        return { icon: <SyncOutlined />, color: '#52c41a', bgColor: '#f6ffed' }
      case 'security':
        return { icon: <SecurityScanOutlined />, color: '#faad14', bgColor: '#fffbe6' }
      case 'info':
      default:
        return { icon: <InfoCircleOutlined />, color: '#8c8c8c', bgColor: '#fafafa' }
    }
  }

  // Get tag for priority
  const getPriorityTag = (priority: NotificationPriority) => {
    switch (priority) {
      case 'critical':
        return <Tag color="red">Critical</Tag>
      case 'high':
        return <Tag color="orange">High</Tag>
      case 'medium':
        return <Tag color="blue">Medium</Tag>
      case 'low':
        return <Tag color="default">Low</Tag>
      default:
        return <Tag>{priority}</Tag>
    }
  }

  // Handle refresh
  const handleRefresh = () => {
    queryClient.invalidateQueries({ queryKey: ['notifications'] })
    queryClient.invalidateQueries({ queryKey: ['unreadCount'] })
    queryClient.invalidateQueries({ queryKey: ['notificationStats'] })
  }

  // Handle mark all as read
  const handleMarkAllAsRead = () => {
    markAllAsReadMutation.mutate()
  }

  // Handle delete
  const handleDelete = (id: string) => {
    deleteNotificationMutation.mutate(id)
  }

  // Handle update preferences
  const handleUpdatePreferences = async (values: any) => {
    try {
      await notificationApi.updateNotificationPreference(values)
      setSettingsVisible(false)
      queryClient.invalidateQueries({ queryKey: ['notificationPreference'] })
    } catch (error) {
      console.error('Failed to update preferences:', error)
    }
  }

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Badge count={unreadCount} offset={[10, 0]}>
            <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
              <BellOutlined /> Notification Center
            </span>
          </Badge>
          {unreadCount > 0 && (
            <Button type="link" onClick={handleMarkAllAsRead} loading={markAllAsReadMutation.isPending}>
              Mark all as read
            </Button>
          )}
        </Space>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            Refresh
          </Button>
          <Button icon={<SettingOutlined />} onClick={() => setSettingsVisible(true)}>
            Settings
          </Button>
        </Space>
      </div>

      {/* Statistics Cards */}
      {stats && (
        <div style={{ display: 'flex', gap: '16px', marginBottom: '24px' }}>
          <Card size="small" style={{ flex: 1 }}>
            <Space>
              <BellOutlined />
              <span>Total: {stats.total}</span>
            </Space>
          </Card>
          <Card size="small" style={{ flex: 1 }}>
            <Space>
              <EyeOutlined />
              <span>Unread: {stats.unread}</span>
            </Space>
          </Card>
          <Card size="small" style={{ flex: 1 }}>
            <Space>
              <AlertOutlined />
              <span>Alerts: {stats.alert}</span>
            </Space>
          </Card>
          <Card size="small" style={{ flex: 1 }}>
            <Space>
              <SyncOutlined />
              <span>Tasks: {stats.task}</span>
            </Space>
          </Card>
        </div>
      )}

      {/* Notifications List */}
      <Card>
        {isLoading ? (
          <div style={{ textAlign: 'center', padding: '40px' }}>Loading...</div>
        ) : notifications.length === 0 ? (
          <Empty description="No notifications" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <List
            dataSource={notifications}
            renderItem={(item) => {
              const typeConfig = getTypeConfig(item.type)
              return (
                <List.Item
                  key={item.id}
                  style={{
                    backgroundColor: item.read ? 'transparent' : '#f5f5f5',
                    padding: '12px',
                    borderRadius: '4px',
                    marginBottom: '8px',
                  }}
                  actions={[
                    !item.read && (
                      <Button
                        size="small"
                        type="text"
                        icon={<CheckOutlined />}
                        onClick={() => markAsReadMutation.mutate(item.id)}
                        loading={markAsReadMutation.isPending}
                      >
                        Mark Read
                      </Button>
                    ),
                    <Popconfirm
                      title="Delete this notification?"
                      onConfirm={() => handleDelete(item.id)}
                      okText="Yes"
                      cancelText="No"
                    >
                      <Button
                        size="small"
                        type="text"
                        danger
                        icon={<DeleteOutlined />}
                        loading={deleteNotificationMutation.isPending}
                      >
                        Delete
                      </Button>
                    </Popconfirm>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={
                      <div
                        style={{
                          width: '40px',
                          height: '40px',
                          borderRadius: '50%',
                          backgroundColor: typeConfig.bgColor,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          color: typeConfig.color,
                        }}
                      >
                        {typeConfig.icon}
                      </div>
                    }
                    title={
                      <Space>
                        <span style={{ fontWeight: item.read ? 'normal' : 'bold' }}>{item.title}</span>
                        {getPriorityTag(item.priority)}
                      </Space>
                    }
                    description={
                      <div>
                        <div>{item.message}</div>
                        <div style={{ color: '#8c8c8c', fontSize: '12px', marginTop: '4px' }}>
                          {dayjs(item.createdAt).fromNow()}
                        </div>
                      </div>
                    }
                  />
                </List.Item>
              )
            }}
          />
        )}
      </Card>

      {/* Settings Drawer */}
      <Drawer
        title="Notification Settings"
        open={settingsVisible}
        onClose={() => setSettingsVisible(false)}
        width={400}
      >
        {preference && (
          <Form
            form={form}
            layout="vertical"
            initialValues={preference}
            onFinish={handleUpdatePreferences}
          >
            <Form.Item label="Email Notifications" name="emailEnabled" valuePropName="checked">
              <Switch />
            </Form.Item>

            <Form.Item label="Web Notifications" name="webEnabled" valuePropName="checked">
              <Switch />
            </Form.Item>

            <Form.Item label="Push Notifications" name="pushEnabled" valuePropName="checked">
              <Switch />
            </Form.Item>

            <Form.Item label="Minimum Priority" name="minPriority">
              <Select>
                <Option value="low">Low</Option>
                <Option value="medium">Medium</Option>
                <Option value="high">High</Option>
                <Option value="critical">Critical</Option>
              </Select>
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" block>
                Save Settings
              </Button>
            </Form.Item>
          </Form>
        )}
      </Drawer>
    </div>
  )
}

export default NotificationCenterPage
