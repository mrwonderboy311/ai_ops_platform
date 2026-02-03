import React, { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Card,
  Row,
  Col,
  Descriptions,
  Tag,
  Button,
  Space,
  message,
  Modal,
  Input,
  Spin,
  Alert,
  Statistic,
  Popconfirm,
} from 'antd'
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  CheckOutlined,
  CloseOutlined,
  TagOutlined,
  CodeOutlined,
  FolderOutlined,
} from '@ant-design/icons'
import { hostApi } from '../api/host'
import type { UpdateHostRequest } from '../types/host'

const HostDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  // State
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [tagModalVisible, setTagModalVisible] = useState(false)
  const [newTag, setNewTag] = useState('')
  const [editForm, setEditForm] = useState<UpdateHostRequest>({})

  // Fetch host details
  const { data: host, isLoading, error, refetch } = useQuery({
    queryKey: ['host', id],
    queryFn: () => hostApi.getHost(id!),
    enabled: !!id,
    refetchInterval: 30000, // Auto-refresh every 30 seconds
  })

  // Update host mutation
  const updateMutation = useMutation({
    mutationFn: (data: UpdateHostRequest) => hostApi.updateHost(id!, data),
    onSuccess: () => {
      message.success('Host updated successfully')
      queryClient.invalidateQueries({ queryKey: ['host', id] })
      setEditModalVisible(false)
    },
    onError: (error: Error) => {
      message.error(`Failed to update host: ${error.message}`)
    },
  })

  // Approve host mutation
  const approveMutation = useMutation({
    mutationFn: () => hostApi.approveHost(id!),
    onSuccess: () => {
      message.success('Host approved successfully')
      queryClient.invalidateQueries({ queryKey: ['host', id] })
    },
    onError: (error: Error) => {
      message.error(`Failed to approve host: ${error.message}`)
    },
  })

  // Reject host mutation
  const rejectMutation = useMutation({
    mutationFn: (reason?: string) => hostApi.rejectHost(id!, { reason }),
    onSuccess: () => {
      message.success('Host rejected successfully')
      queryClient.invalidateQueries({ queryKey: ['host', id] })
      setRejectModalVisible(false)
      setRejectReason('')
    },
    onError: (error: Error) => {
      message.error(`Failed to reject host: ${error.message}`)
    },
  })

  // Delete host mutation
  const deleteMutation = useMutation({
    mutationFn: () => hostApi.deleteHost(id!),
    onSuccess: () => {
      message.success('Host deleted successfully')
      navigate('/hosts')
    },
    onError: (error: Error) => {
      message.error(`Failed to delete host: ${error.message}`)
    },
  })

  // Reject modal state
  const [rejectModalVisible, setRejectModalVisible] = useState(false)
  const [rejectReason, setRejectReason] = useState('')

  // Add tag mutation (simulated - would need API)
  const handleAddTag = () => {
    if (newTag && host) {
      const updatedTags = [...(host.tags || []), newTag]
      updateMutation.mutate({ tags: updatedTags })
      setNewTag('')
      setTagModalVisible(false)
    }
  }

  // Remove tag
  const handleRemoveTag = (tagToRemove: string) => {
    if (host) {
      const updatedTags = (host.tags || []).filter(tag => tag !== tagToRemove)
      updateMutation.mutate({ tags: updatedTags })
    }
  }

  // Handle edit
  const handleEdit = () => {
    if (host) {
      setEditForm({
        hostname: host.hostname,
        port: host.port,
        osType: host.osType,
        osVersion: host.osVersion,
        cpuCores: host.cpuCores ?? undefined,
        memoryGB: host.memoryGB ?? undefined,
        diskGB: host.diskGB ?? undefined,
      })
      setEditModalVisible(true)
    }
  }

  const handleEditSubmit = () => {
    updateMutation.mutate(editForm)
  }

  // Loading state
  if (isLoading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    )
  }

  // Error state
  if (error || !host) {
    return (
      <div style={{ padding: '24px' }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/hosts')}>
          Back to Hosts
        </Button>
        <Alert
          style={{ marginTop: '16px' }}
          message="Failed to load host"
          description={(error as Error)?.message || 'Host not found'}
          type="error"
          showIcon
          action={
            <Button onClick={() => refetch()}>Retry</Button>
          }
        />
      </div>
    )
  }

  // Status badge
  const statusConfig = {
    pending: { color: 'orange', text: 'Pending' },
    approved: { color: 'blue', text: 'Approved' },
    rejected: { color: 'red', text: 'Rejected' },
    offline: { color: 'default', text: 'Offline' },
    online: { color: 'green', text: 'Online' },
  }
  const statusInfo = statusConfig[host.status] || statusConfig.pending

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/hosts')}>
            Back
          </Button>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
            {host.hostname || host.ipAddress}
          </span>
          <Tag color={statusInfo.color}>{statusInfo.text}</Tag>
        </Space>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
            Refresh
          </Button>
          <Button icon={<EditOutlined />} onClick={handleEdit}>
            Edit
          </Button>

          {(host.status === 'approved' || host.status === 'online') && (
            <>
              <Button
                type="primary"
                icon={<CodeOutlined />}
                onClick={() => navigate(`/hosts/ssh/${id}`)}
              >
                Terminal
              </Button>
              <Button
                icon={<FolderOutlined />}
                onClick={() => navigate(`/hosts/files/${id}`)}
              >
                File Manager
              </Button>
            </>
          )}

          {host.status === 'pending' && (
            <>
              <Popconfirm
                title="Approve this host?"
                onConfirm={() => approveMutation.mutate()}
              >
                <Button type="primary" icon={<CheckOutlined />}>
                  Approve
                </Button>
              </Popconfirm>

              <Button
                danger
                icon={<CloseOutlined />}
                onClick={() => setRejectModalVisible(true)}
              >
                Reject
              </Button>
            </>
          )}

          <Popconfirm
            title="Delete this host?"
            description="This action cannot be undone"
            onConfirm={() => deleteMutation.mutate()}
            okText="Yes"
            cancelText="No"
          >
            <Button danger icon={<DeleteOutlined />}>
              Delete
            </Button>
          </Popconfirm>
        </Space>
      </div>

      <Row gutter={[16, 16]}>
        {/* Basic Information */}
        <Col span={24}>
          <Card title="Basic Information" extra={
            <Button type="link" onClick={handleEdit}>Edit</Button>
          }>
            <Descriptions column={2} bordered>
              <Descriptions.Item label="Hostname">{host.hostname || '-'}</Descriptions.Item>
              <Descriptions.Item label="IP Address">{host.ipAddress}:{host.port}</Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={statusInfo.color}>{statusInfo.text}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="OS Type">{host.osType || '-'}</Descriptions.Item>
              <Descriptions.Item label="OS Version">{host.osVersion || '-'}</Descriptions.Item>
              <Descriptions.Item label="Architecture">{host.labels?.arch || '-'}</Descriptions.Item>
              <Descriptions.Item label="Kernel Version">{host.labels?.kernel_version || '-'}</Descriptions.Item>
              <Descriptions.Item label="CPU Model">{host.labels?.cpu_model || '-'}</Descriptions.Item>
              <Descriptions.Item label="Registered By">{host.registeredByUser?.username || '-'}</Descriptions.Item>
              <Descriptions.Item label="Approved By">
                {host.approvedByUser?.username || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Approved At">
                {host.approvedAt ? new Date(host.approvedAt).toLocaleString() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Last Seen">
                {host.lastSeenAt ? new Date(host.lastSeenAt).toLocaleString() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Created At">
                {new Date(host.createdAt).toLocaleString()}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        {/* Hardware Information */}
        <Col span={12}>
          <Card title="Hardware">
            <Row gutter={16}>
              <Col span={12}>
                <Statistic title="CPU Cores" value={host.cpuCores ?? '-'} />
              </Col>
              <Col span={12}>
                <Statistic
                  title="Memory"
                  value={host.memoryGB ?? '-'}
                  suffix="GB"
                />
              </Col>
            </Row>
            {host.diskGB !== null && (
              <div style={{ marginTop: '16px' }}>
                <Statistic
                  title="Disk Space"
                  value={host.diskGB}
                  suffix="GB"
                />
              </div>
            )}
          </Card>
        </Col>

        {/* Network Information */}
        <Col span={12}>
          <Card title="Network">
            <Descriptions column={1} bordered>
              <Descriptions.Item label="IP Address">{host.ipAddress}</Descriptions.Item>
              <Descriptions.Item label="Port">{host.port}</Descriptions.Item>
              <Descriptions.Item label="MAC Address">{host.labels?.mac_address || '-'}</Descriptions.Item>
              <Descriptions.Item label="Gateway">{host.labels?.gateway || '-'}</Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        {/* Labels Management */}
        <Col span={12}>
          <Card
            title="Labels"
            extra={<Button type="link" onClick={() => setTagModalVisible(true)}>Add Label</Button>}
          >
            <div style={{ marginBottom: '8px' }}>
              {Object.entries(host.labels || {}).map(([key, value]) => (
                <Tag key={key} style={{ marginBottom: '4px' }}>
                  {key}: {value}
                </Tag>
              ))}
            </div>
          </Card>
        </Col>

        {/* Tags */}
        <Col span={12}>
          <Card
            title="Tags"
            extra={
              <Button
                type="link"
                icon={<TagOutlined />}
                onClick={() => setTagModalVisible(true)}
              >
                Add Tag
              </Button>
            }
          >
            <div>
              {(host.tags || []).map(tag => (
                <Tag
                  key={tag}
                  closable
                  onClose={() => handleRemoveTag(tag)}
                  style={{ marginBottom: '4px' }}
                >
                  {tag}
                </Tag>
              ))}
              {(host.tags || []).length === 0 && <span style={{ color: '#999' }}>No tags</span>}
            </div>
          </Card>
        </Col>

        {/* Monitoring Charts (Placeholder for Epic 7) */}
        <Col span={24}>
          <Card title="Monitoring">
            <Alert
              message="Monitoring charts will be available in Epic 7"
              description="CPU, memory, disk, and network usage trends will be displayed here"
              type="info"
              showIcon
            />
          </Card>
        </Col>

        {/* Operation History (Placeholder) */}
        <Col span={24}>
          <Card title="Operation History">
            <Alert
              message="Operation history tracking will be implemented in a future story"
              description="Audit logs for host operations will be displayed here"
              type="info"
              showIcon
            />
          </Card>
        </Col>
      </Row>

      {/* Edit Modal */}
      <Modal
        title="Edit Host"
        open={editModalVisible}
        onOk={handleEditSubmit}
        onCancel={() => setEditModalVisible(false)}
        okText="Save"
      >
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <div>
            <label>Hostname:</label>
            <Input
              value={editForm.hostname}
              onChange={(e) => setEditForm({ ...editForm, hostname: e.target.value })}
            />
          </div>
          <div>
            <label>Port:</label>
            <Input
              type="number"
              value={editForm.port}
              onChange={(e) => setEditForm({ ...editForm, port: parseInt(e.target.value) || 22 })}
            />
          </div>
          <div>
            <label>OS Type:</label>
            <Input
              value={editForm.osType}
              onChange={(e) => setEditForm({ ...editForm, osType: e.target.value })}
            />
          </div>
          <div>
            <label>OS Version:</label>
            <Input
              value={editForm.osVersion}
              onChange={(e) => setEditForm({ ...editForm, osVersion: e.target.value })}
            />
          </div>
        </Space>
      </Modal>

      {/* Tag Modal */}
      <Modal
        title="Add Tag"
        open={tagModalVisible}
        onOk={handleAddTag}
        onCancel={() => {
          setTagModalVisible(false)
          setNewTag('')
        }}
      >
        <Input
          placeholder="Enter tag name"
          value={newTag}
          onChange={(e) => setNewTag(e.target.value)}
          onPressEnter={handleAddTag}
        />
      </Modal>

      {/* Reject Modal */}
      <Modal
        title="Reject Host"
        open={rejectModalVisible}
        onOk={() => rejectMutation.mutate(rejectReason || undefined)}
        onCancel={() => {
          setRejectModalVisible(false)
          setRejectReason('')
        }}
        okText="Reject"
        okButtonProps={{ danger: true }}
      >
        <p>Are you sure you want to reject this host? The agent will no longer be able to report data.</p>
        <Input.TextArea
          rows={4}
          placeholder="Reason for rejection (optional)"
          value={rejectReason}
          onChange={(e) => setRejectReason(e.target.value)}
          maxLength={500}
          showCount
        />
      </Modal>
    </div>
  )
}

export default HostDetailPage
