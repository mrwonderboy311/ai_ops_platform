import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Table, Button, Space, Tag, message, Select, Popconfirm, Card, Statistic, Row, Col } from 'antd'
import {
  PlusOutlined,
  EyeOutlined,
  DeleteOutlined,
  ReloadOutlined,
  CloudServerOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { clusterApi } from '../api/cluster'
import type { K8sCluster, ClusterStatus, ClusterType } from '../types/cluster'
import { CreateClusterModal } from '../components/CreateClusterModal'

const { Option } = Select

export const ClusterListPage: React.FC = () => {
  const navigate = useNavigate()
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [statusFilter, setStatusFilter] = useState<string | undefined>()
  const [typeFilter, setTypeFilter] = useState<string | undefined>()
  const [providerFilter, setProviderFilter] = useState<string | undefined>()

  // Fetch clusters
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['clusters', page, pageSize, statusFilter, typeFilter, providerFilter],
    queryFn: () =>
      clusterApi.listClusters({
        page,
        pageSize,
        status: statusFilter,
        type: typeFilter,
        provider: providerFilter,
      }),
  })

  // Handle delete
  const handleDelete = async (clusterId: string) => {
    try {
      await clusterApi.deleteCluster(clusterId)
      message.success('Cluster deleted successfully')
      refetch()
    } catch (error: any) {
      message.error(`Failed to delete: ${error.response?.data?.message || error.message}`)
    }
  }

  // Get status color and icon
  const getStatusConfig = (status: ClusterStatus) => {
    const configs = {
      connected: { color: 'success', icon: <CheckCircleOutlined />, label: 'Connected' },
      pending: { color: 'default', icon: <ExclamationCircleOutlined />, label: 'Pending' },
      error: { color: 'error', icon: <CloseCircleOutlined />, label: 'Error' },
      disabled: { color: 'warning', icon: <CloseCircleOutlined />, label: 'Disabled' },
    }
    return configs[status] || configs.pending
  }

  // Get type label
  const getTypeLabel = (type: ClusterType): string => {
    const labels: Record<ClusterType, string> = {
      managed: 'Managed',
      'self-hosted': 'Self-Hosted',
    }
    return labels[type] || type
  }

  // Calculate statistics
  const stats = React.useMemo(() => {
    if (!data?.clusters) return { total: 0, connected: 0, error: 0, pending: 0 }
    return {
      total: data.total,
      connected: data.clusters.filter((c) => c.status === 'connected').length,
      error: data.clusters.filter((c) => c.status === 'error').length,
      pending: data.clusters.filter((c) => c.status === 'pending').length,
    }
  }, [data])

  // Table columns
  const columns: ColumnsType<K8sCluster> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name, record) => (
        <Space>
          <CloudServerOutlined />
          <a onClick={() => navigate(`/clusters/${record.id}`)}>{name}</a>
        </Space>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: 120,
      render: (type: ClusterType) => (
        <Tag>{getTypeLabel(type)}</Tag>
      ),
    },
    {
      title: 'Provider',
      dataIndex: 'provider',
      key: 'provider',
      width: 100,
      render: (provider: string) => provider ? <Tag>{provider.toUpperCase()}</Tag> : '-',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: ClusterStatus) => {
        const config = getStatusConfig(status)
        return (
          <Tag color={config.color} icon={config.icon}>
            {config.label}
          </Tag>
        )
      },
    },
    {
      title: 'Version',
      dataIndex: 'version',
      key: 'version',
      width: 100,
    },
    {
      title: 'Nodes',
      dataIndex: 'nodeCount',
      key: 'nodeCount',
      width: 80,
      render: (count: number) => (count > 0 ? count : '-'),
    },
    {
      title: 'Region',
      dataIndex: 'region',
      key: 'region',
      width: 120,
    },
    {
      title: 'Last Connected',
      dataIndex: 'lastConnectedAt',
      key: 'lastConnectedAt',
      width: 180,
      render: (date?: string) => (date ? new Date(date).toLocaleString() : 'Never'),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/clusters/${record.id}`)}
          >
            View
          </Button>
          <Popconfirm
            title="Delete this cluster?"
            description="This will disconnect and remove the cluster from the platform."
            onConfirm={() => handleDelete(record.id)}
          >
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
            >
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0 }}>Kubernetes Clusters</h1>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalVisible(true)}>
          Add Cluster
        </Button>
      </div>

      {/* Statistics */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic title="Total Clusters" value={stats.total} prefix={<CloudServerOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Connected"
              value={stats.connected}
              valueStyle={{ color: '#52c41a' }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Error"
              value={stats.error}
              valueStyle={{ color: '#f5222d' }}
              prefix={<CloseCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Pending"
              value={stats.pending}
              valueStyle={{ color: '#faad14' }}
              prefix={<ExclamationCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* Filters */}
      <Space style={{ marginBottom: '16px' }}>
        <Select
          placeholder="Filter by status"
          allowClear
          style={{ width: 130 }}
          onChange={(value) => {
            setStatusFilter(value)
            setPage(1)
          }}
          value={statusFilter}
        >
          <Option value="connected">Connected</Option>
          <Option value="pending">Pending</Option>
          <Option value="error">Error</Option>
          <Option value="disabled">Disabled</Option>
        </Select>
        <Select
          placeholder="Filter by type"
          allowClear
          style={{ width: 130 }}
          onChange={(value) => {
            setTypeFilter(value)
            setPage(1)
          }}
          value={typeFilter}
        >
          <Option value="managed">Managed</Option>
          <Option value="self-hosted">Self-Hosted</Option>
        </Select>
        <Select
          placeholder="Filter by provider"
          allowClear
          style={{ width: 130 }}
          onChange={(value) => {
            setProviderFilter(value)
            setPage(1)
          }}
          value={providerFilter}
        >
          <Option value="aws">AWS</Option>
          <Option value="gcp">GCP</Option>
          <Option value="azure">Azure</Option>
          <Option value="alibaba">Alibaba</Option>
          <Option value="tencent">Tencent</Option>
        </Select>
        <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
          Refresh
        </Button>
      </Space>

      {/* Table */}
      <Table
        columns={columns}
        dataSource={data?.clusters || []}
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: page,
          pageSize: pageSize,
          total: data?.total || 0,
          showSizeChanger: true,
          showTotal: (total) => `Total ${total} clusters`,
          onChange: (newPage, newPageSize) => {
            setPage(newPage)
            setPageSize(newPageSize || 20)
          },
        }}
      />

      {/* Create Modal */}
      <CreateClusterModal
        visible={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onSuccess={() => {
          setCreateModalVisible(false)
          refetch()
        }}
      />
    </div>
  )
}

export default ClusterListPage
