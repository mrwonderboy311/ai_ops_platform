import React, { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Button, Space, message, Table, Tag, Tabs, Select, Spin, Modal } from 'antd'
import {
  ArrowLeftOutlined,
  ReloadOutlined,
  ContainerOutlined,
  AppstoreOutlined,
  ClusterOutlined,
  DeleteOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { workloadApi } from '../api/workload'
import type { K8sPod, K8sDeployment, K8sService } from '../types/workload'

const { Option } = Select

export const WorkloadListPage: React.FC = () => {
  const { id: clusterId } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [namespace, setNamespace] = useState<string>('default')
  const [activeTab, setActiveTab] = useState('pods')
  const [logsModal, setLogsModal] = useState<{ visible: boolean; podName: string; logs: string }>({
    visible: false,
    podName: '',
    logs: '',
  })

  // Fetch namespaces
  const { data: namespaces = [], isLoading: namespacesLoading, refetch: refetchNamespaces } = useQuery({
    queryKey: ['namespaces', clusterId],
    queryFn: () => workloadApi.getNamespaces(clusterId!),
    enabled: !!clusterId,
  })

  // Fetch pods
  const { data: pods = [], isLoading: podsLoading, refetch: refetchPods } = useQuery({
    queryKey: ['pods', clusterId, namespace],
    queryFn: () => workloadApi.getPods(clusterId!, namespace),
    enabled: !!clusterId && !!namespace,
  })

  // Fetch deployments
  const { data: deployments = [], isLoading: deploymentsLoading, refetch: refetchDeployments } = useQuery({
    queryKey: ['deployments', clusterId, namespace],
    queryFn: () => workloadApi.getDeployments(clusterId!, namespace),
    enabled: !!clusterId && !!namespace && activeTab === 'deployments',
  })

  // Fetch services
  const { data: services = [], isLoading: servicesLoading, refetch: refetchServices } = useQuery({
    queryKey: ['services', clusterId, namespace],
    queryFn: () => workloadApi.getServices(clusterId!, namespace),
    enabled: !!clusterId && !!namespace && activeTab === 'services',
  })

  // Handle refresh
  const handleRefresh = () => {
    refetchNamespaces()
    refetchPods()
    refetchDeployments()
    refetchServices()
  }

  // Handle view logs
  const handleViewLogs = async (podName: string) => {
    try {
      const logs = await workloadApi.getPodLogs(clusterId!, namespace, podName, 100)
      setLogsModal({ visible: true, podName, logs })
    } catch (error: any) {
      message.error(`Failed to fetch logs: ${error.response?.data?.message || error.message}`)
    }
  }

  // Handle delete pod
  const handleDeletePod = async (podName: string) => {
    Modal.confirm({
      title: 'Delete Pod',
      content: `Are you sure you want to delete pod "${podName}"?`,
      onOk: async () => {
        try {
          await workloadApi.deletePod(clusterId!, namespace, podName)
          message.success('Pod deleted successfully')
          refetchPods()
        } catch (error: any) {
          message.error(`Failed to delete pod: ${error.response?.data?.message || error.message}`)
        }
      },
    })
  }

  // Pod table columns
  const podColumns: ColumnsType<K8sPod> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record) => (
        <Space>
          <ContainerOutlined />
          <a onClick={() => navigate(`/clusters/${clusterId}/namespaces/${namespace}/pods/${name}`)}>
            {name}
          </a>
          {record.ownerName && <Tag color="blue">{record.ownerType}/{record.ownerName}</Tag>}
        </Space>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string, record) => (
        <Tag color={record.ready ? 'success' : 'error'} icon={record.ready && <ContainerOutlined />}>
          {status}
        </Tag>
      ),
    },
    {
      title: 'Phase',
      dataIndex: 'phase',
      key: 'phase',
      width: 100,
    },
    {
      title: 'Node',
      dataIndex: 'nodeName',
      key: 'nodeName',
      width: 150,
      ellipsis: true,
    },
    {
      title: 'IP',
      dataIndex: 'podIp',
      key: 'podIp',
      width: 120,
      render: (ip: string) => ip || '-',
    },
    {
      title: 'Restarts',
      dataIndex: 'restartCount',
      key: 'restartCount',
      width: 80,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 150,
      render: (_: any, record) => (
        <Space>
          <Button size="small" icon={<FileTextOutlined />} onClick={() => handleViewLogs(record.name)}>
            Logs
          </Button>
          <Button size="small" danger icon={<DeleteOutlined />} onClick={() => handleDeletePod(record.name)}>
            Delete
          </Button>
        </Space>
      ),
    },
  ]

  // Deployment table columns
  const deploymentColumns: ColumnsType<K8sDeployment> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <AppstoreOutlined />
          {name}
        </Space>
      ),
    },
    {
      title: 'Ready',
      key: 'ready',
      width: 120,
      render: (_: any, record) => (
        <span>
          {record.readyReplicas} / {record.replicas}
        </span>
      ),
    },
    {
      title: 'Up-to-date',
      dataIndex: 'updatedReplicas',
      key: 'updatedReplicas',
      width: 100,
    },
    {
      title: 'Available',
      dataIndex: 'availableReplicas',
      key: 'availableReplicas',
      width: 100,
    },
    {
      title: 'Image',
      dataIndex: 'image',
      key: 'image',
      ellipsis: true,
    },
  ]

  // Service table columns
  const serviceColumns: ColumnsType<K8sService> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <ClusterOutlined />
          {name}
        </Space>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: 120,
      render: (type: string) => <Tag color="blue">{type}</Tag>,
    },
    {
      title: 'Cluster IP',
      dataIndex: 'clusterIp',
      key: 'clusterIp',
      width: 150,
    },
    {
      title: 'Ports',
      key: 'ports',
      width: 200,
      render: (_: any, record) => (
        <Space direction="vertical" size={0}>
          {record.ports.map((port, i) => (
            <span key={i} style={{ fontSize: '12px' }}>
              {port.name && <span>{port.name}: </span>}
              {port.port}/{port.protocol}
              {port.nodePort && <span> (NodePort: {port.nodePort})</span>}
            </span>
          ))}
        </Space>
      ),
    },
  ]

  if (namespacesLoading && namespaces.length === 0) {
    return (
      <div style={{ padding: '24px', textAlign: 'center', height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin size="large" tip="Loading namespaces..." />
      </div>
    )
  }

  const tabItems = [
    {
      key: 'pods',
      label: `Pods (${pods.length})`,
      children: (
        <Table
          columns={podColumns}
          dataSource={pods}
          rowKey="name"
          loading={podsLoading}
          pagination={false}
          size="small"
        />
      ),
    },
    {
      key: 'deployments',
      label: `Deployments (${deployments.length})`,
      children: (
        <Table
          columns={deploymentColumns}
          dataSource={deployments}
          rowKey="name"
          loading={deploymentsLoading}
          pagination={false}
          size="small"
        />
      ),
    },
    {
      key: 'services',
      label: `Services (${services.length})`,
      children: (
        <Table
          columns={serviceColumns}
          dataSource={services}
          rowKey="name"
          loading={servicesLoading}
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
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/clusters/${clusterId}`)}>
            Back to Cluster
          </Button>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
            <ContainerOutlined /> Workloads
          </span>
        </Space>
        <Space>
          <Select
            value={namespace}
            onChange={setNamespace}
            style={{ width: 200 }}
            loading={namespacesLoading}
          >
            {namespaces.map((ns) => (
              <Option key={ns} value={ns}>
                {ns}
              </Option>
            ))}
          </Select>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            Refresh
          </Button>
        </Space>
      </div>

      {/* Tabs */}
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />

      {/* Logs Modal */}
      <Modal
        title={`Logs: ${logsModal.podName}`}
        open={logsModal.visible}
        onCancel={() => setLogsModal({ ...logsModal, visible: false })}
        footer={[
          <Button key="close" onClick={() => setLogsModal({ ...logsModal, visible: false })}>
            Close
          </Button>,
        ]}
        width={800}
      >
        <div style={{ maxHeight: '400px', overflow: 'auto', background: '#1e1e1e', padding: '12px' }}>
          <pre style={{ color: '#d4d4d4', margin: 0, fontSize: '12px', whiteSpace: 'pre-wrap' }}>
            {logsModal.logs || 'No logs available'}
          </pre>
        </div>
      </Modal>
    </div>
  )
}

export default WorkloadListPage
