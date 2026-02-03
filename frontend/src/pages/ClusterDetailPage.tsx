import React, { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Button, Space, Tag, message, Card, Descriptions, Table, Tabs, Alert, Spin, Row, Col, Statistic } from 'antd'
import {
  ArrowLeftOutlined,
  ReloadOutlined,
  DeleteOutlined,
  CloudServerOutlined,
  ApiOutlined,
  ContainerOutlined,
  NodeIndexOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { clusterApi } from '../api/cluster'
import type { ClusterNode, ClusterStatus } from '../types/cluster'

export const ClusterDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [autoRefresh, setAutoRefresh] = useState(false)

  // Fetch cluster details
  const { data: cluster, isLoading, error, refetch } = useQuery({
    queryKey: ['cluster', id],
    queryFn: () => clusterApi.getCluster(id!),
    enabled: !!id,
    refetchInterval: autoRefresh ? 10000 : false,
  })

  // Fetch cluster info
  const { data: clusterInfo } = useQuery({
    queryKey: ['clusterInfo', id],
    queryFn: () => clusterApi.getClusterInfo(id!),
    enabled: !!id,
    refetchInterval: autoRefresh ? 10000 : false,
  })

  // Fetch cluster nodes
  const { data: nodes, refetch: refetchNodes } = useQuery({
    queryKey: ['clusterNodes', id],
    queryFn: () => clusterApi.getClusterNodes(id!),
    enabled: !!id,
    refetchInterval: autoRefresh ? 10000 : false,
  })

  // Handle delete
  const handleDelete = async () => {
    try {
      await clusterApi.deleteCluster(id!)
      message.success('Cluster deleted successfully')
      navigate('/clusters')
    } catch (error: any) {
      message.error(`Failed to delete: ${error.response?.data?.message || error.message}`)
    }
  }

  // Get status config
  const getStatusConfig = (status: ClusterStatus) => {
    const configs = {
      connected: { color: 'success', label: 'Connected', icon: <ApiOutlined /> },
      pending: { color: 'default', label: 'Pending', icon: <CloudServerOutlined /> },
      error: { color: 'error', label: 'Error', icon: <DeleteOutlined /> },
      disabled: { color: 'warning', label: 'Disabled', icon: <DeleteOutlined /> },
    }
    return configs[status] || configs.pending
  }

  // Node table columns
  const nodeColumns: ColumnsType<ClusterNode> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <NodeIndexOutlined />
          {name}
        </Space>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={status === 'Ready' ? 'success' : 'error'}>{status}</Tag>
      ),
    },
    {
      title: 'Roles',
      dataIndex: 'roles',
      key: 'roles',
      width: 100,
    },
    {
      title: 'Version',
      dataIndex: 'version',
      key: 'version',
      width: 100,
    },
    {
      title: 'OS Image',
      dataIndex: 'osImage',
      key: 'osImage',
      ellipsis: true,
    },
    {
      title: 'CPU',
      key: 'cpu',
      width: 120,
      render: (_, record) => (
        <div style={{ fontSize: '12px' }}>
          <div>Capacity: {record.cpuCapacity}</div>
          <div style={{ color: '#888' }}>Allocatable: {record.cpuAllocatable}</div>
        </div>
      ),
    },
    {
      title: 'Memory',
      key: 'memory',
      width: 150,
      render: (_, record) => (
        <div style={{ fontSize: '12px' }}>
          <div>Capacity: {record.memoryCapacity}</div>
          <div style={{ color: '#888' }}>Allocatable: {record.memoryAllocatable}</div>
        </div>
      ),
    },
    {
      title: 'Internal IP',
      dataIndex: 'internalIp',
      key: 'internalIp',
      width: 120,
    },
    {
      title: 'External IP',
      dataIndex: 'externalIp',
      key: 'externalIp',
      width: 120,
      render: (ip: string) => ip || '-',
    },
  ]

  if (isLoading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center', height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin size="large" tip="Loading cluster details..." />
      </div>
    )
  }

  if (error || !cluster) {
    return (
      <div style={{ padding: '24px' }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/clusters')}>
          Back to Clusters
        </Button>
        <Alert
          style={{ marginTop: '16px' }}
          message="Failed to load cluster"
          description={(error as Error)?.message || 'Cluster not found'}
          type="error"
          showIcon
        />
      </div>
    )
  }

  const statusConfig = getStatusConfig(cluster.status)

  const tabItems = [
    {
      key: 'overview',
      label: 'Overview',
      children: (
        <>
          <Card title="Cluster Information" style={{ marginBottom: '16px' }}>
            <Descriptions column={2} bordered>
              <Descriptions.Item label="Cluster Name" span={2}>
                {cluster.name}
              </Descriptions.Item>
              <Descriptions.Item label="Description" span={2}>
                {cluster.description || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={statusConfig.color} icon={statusConfig.icon}>
                  {statusConfig.label}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Type">
                {cluster.type === 'managed' ? 'Managed' : 'Self-Hosted'}
              </Descriptions.Item>
              <Descriptions.Item label="Provider">
                {cluster.provider ? cluster.provider.toUpperCase() : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Region">
                {cluster.region || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Kubernetes Version">
                {cluster.version || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="API Endpoint">
                {cluster.endpoint || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="Created">
                {new Date(cluster.createdAt).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="Last Connected">
                {cluster.lastConnectedAt ? new Date(cluster.lastConnectedAt).toLocaleString() : 'Never'}
              </Descriptions.Item>
            </Descriptions>
          </Card>

          {cluster.errorMessage && (
            <Alert
              style={{ marginBottom: '16px' }}
              message="Connection Error"
              description={cluster.errorMessage}
              type="error"
              showIcon
              closable
            />
          )}

          {clusterInfo && (
            <>
              <Row gutter={16} style={{ marginBottom: '16px' }}>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="Nodes"
                      value={clusterInfo.nodeCount}
                      prefix={<NodeIndexOutlined />}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="Namespaces"
                      value={clusterInfo.namespaceCount}
                      prefix={<ContainerOutlined />}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic title="Pods" value={clusterInfo.podCount} />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic title="Deployments" value={clusterInfo.deploymentCount} />
                  </Card>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col span={6}>
                  <Card>
                    <Statistic title="Services" value={clusterInfo.serviceCount} />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic title="Ingress" value={clusterInfo.ingressCount} />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic title="ConfigMaps" value={clusterInfo.configMapCount} />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic title="Secrets" value={clusterInfo.secretCount} />
                  </Card>
                </Col>
              </Row>
            </>
          )}
        </>
      ),
    },
    {
      key: 'nodes',
      label: `Nodes (${nodes?.length || 0})`,
      children: (
        <Table
          columns={nodeColumns}
          dataSource={nodes || []}
          rowKey="id"
          loading={!nodes}
          pagination={false}
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
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/clusters')}>
            Back to Clusters
          </Button>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
            <CloudServerOutlined /> {cluster.name}
          </span>
          <Tag color={statusConfig.color} icon={statusConfig.icon}>
            {statusConfig.label}
          </Tag>
        </Space>
        <Space>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => {
              refetch()
              refetchNodes()
            }}
          >
            Refresh
          </Button>
          {cluster.status !== 'connected' && (
            <Button danger icon={<DeleteOutlined />} onClick={handleDelete}>
              Delete
            </Button>
          )}
        </Space>
      </div>

      {/* Auto-refresh indicator */}
      {autoRefresh && (
        <Alert
          style={{ marginBottom: '16px' }}
          message="Auto-refreshing"
          description="Cluster information is being updated automatically."
          type="info"
          showIcon
          closable
          onClose={() => setAutoRefresh(false)}
        />
      )}

      {/* Tabs */}
      <Tabs items={tabItems} />
    </div>
  )
}

export default ClusterDetailPage
