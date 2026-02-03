import React, { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Button, Space, Descriptions, Tag, Card, Tabs, Table, Spin, message, Alert, Modal } from 'antd'
import {
  ArrowLeftOutlined,
  ReloadOutlined,
  ContainerOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  FileTextOutlined,
  PlayCircleOutlined,
  CodeOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { workloadApi } from '../api/workload'
import { StreamingLogs } from '../components/StreamingLogs'
import { PodTerminal } from '../components/PodTerminal'
import dayjs from 'dayjs'

interface ContainerInfo {
  name: string
  image: string
  imagePullPolicy: string
  ready: boolean
  restartCount: number
  state: string
  cpuRequest?: string
  memoryRequest?: string
  cpuLimit?: string
  memoryLimit?: string
}

interface EventInfo {
  type: string
  reason: string
  message: string
  firstSeen: string
  lastSeen: string
  count: number
}

interface PodDetail {
  name: string
  namespace: string
  status: string
  phase: string
  podIp: string
  hostIp: string
  nodeName: string
  ready: boolean
  restartCount: number
  ownerType?: string
  ownerName?: string
  containers: ContainerInfo[]
  events: EventInfo[]
  labels: Record<string, string>
  annotations: Record<string, string>
  serviceAccount: string
  restartPolicy: string
  dnsPolicy: string
  createdAt: string
  startTime?: string
  qosClass: string
}

export const PodDetailPage: React.FC = () => {
  const { id: clusterId, namespace, podName } = useParams<{ id: string; namespace: string; podName: string }>()
  const navigate = useNavigate()
  const [logsModal, setLogsModal] = useState<{ visible: boolean; containerName: string; logs: string }>({
    visible: false,
    containerName: '',
    logs: '',
  })
  const [streamingLogsModal, setStreamingLogsModal] = useState<{ visible: boolean; containerName: string }>({
    visible: false,
    containerName: '',
  })
  const [terminalModal, setTerminalModal] = useState<{ visible: boolean; containerName: string }>({
    visible: false,
    containerName: '',
  })

  // Fetch pod detail
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['podDetail', clusterId, namespace, podName],
    queryFn: () => workloadApi.getPodDetail(clusterId!, namespace!, podName!),
    enabled: !!clusterId && !!namespace && !!podName,
  })

  // Handle view logs
  const handleViewLogs = async (containerName: string) => {
    try {
      const logs = await workloadApi.getPodLogs(clusterId!, namespace!, podName!, 500)
      setLogsModal({ visible: true, containerName, logs })
    } catch (error: any) {
      message.error(`Failed to fetch logs: ${error.response?.data?.message || error.message}`)
    }
  }

  // Get status tag
  const getStatusTag = (pod: PodDetail) => {
    if (pod.ready) {
      return <Tag color="success" icon={<CheckCircleOutlined />}>Ready</Tag>
    }
    if (pod.phase === 'Pending') {
      return <Tag color="warning" icon={<ClockCircleOutlined />}>Pending</Tag>
    }
    if (pod.phase === 'Failed' || pod.phase === 'Error') {
      return <Tag color="error" icon={<CloseCircleOutlined />}>{pod.phase}</Tag>
    }
    return <Tag color="default" icon={<ExclamationCircleOutlined />}>{pod.phase}</Tag>
  }

  // Get container state tag
  const getContainerStateTag = (container: ContainerInfo) => {
    if (container.ready) {
      return <Tag color="success" icon={<CheckCircleOutlined />}>Running</Tag>
    }
    if (container.state.startsWith('Waiting')) {
      return <Tag color="warning" icon={<ClockCircleOutlined />}>{container.state}</Tag>
    }
    if (container.state.startsWith('Terminated')) {
      return <Tag color="default" icon={<ExclamationCircleOutlined />}>{container.state}</Tag>
    }
    return <Tag color="default">{container.state}</Tag>
  }

  // Container table columns
  const containerColumns: ColumnsType<ContainerInfo> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <ContainerOutlined />
          {name}
        </Space>
      ),
    },
    {
      title: 'Image',
      dataIndex: 'image',
      key: 'image',
      ellipsis: true,
    },
    {
      title: 'State',
      key: 'state',
      render: (_: any, record: ContainerInfo) => getContainerStateTag(record),
    },
    {
      title: 'Ready',
      key: 'ready',
      render: (_: any, record: ContainerInfo) => (
        <span>{record.ready ? 'Yes' : 'No'}</span>
      ),
    },
    {
      title: 'Restarts',
      dataIndex: 'restartCount',
      key: 'restartCount',
    },
    {
      title: 'Resources',
      key: 'resources',
      render: (_: any, record: ContainerInfo) => (
        <div style={{ fontSize: '12px' }}>
          {record.cpuRequest && (
            <div>CPU: {record.cpuRequest}
              {record.cpuLimit && ` / ${record.cpuLimit}`}
            </div>
          )}
          {record.memoryRequest && (
            <div>Memory: {record.memoryRequest}
              {record.memoryLimit && ` / ${record.memoryLimit}`}
            </div>
          )}
        </div>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: ContainerInfo) => (
        <Space>
          <Button
            size="small"
            icon={<CodeOutlined />}
            onClick={() => setTerminalModal({ visible: true, containerName: record.name })}
          >
            Terminal
          </Button>
          <Button
            size="small"
            icon={<PlayCircleOutlined />}
            onClick={() => setStreamingLogsModal({ visible: true, containerName: record.name })}
          >
            Stream
          </Button>
          <Button
            size="small"
            icon={<FileTextOutlined />}
            onClick={() => handleViewLogs(record.name)}
          >
            Logs
          </Button>
        </Space>
      ),
    },
  ]

  // Event table columns
  const eventColumns: ColumnsType<EventInfo> = [
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color={type === 'Normal' ? 'blue' : 'orange'}>{type}</Tag>
      ),
    },
    {
      title: 'Reason',
      dataIndex: 'reason',
      key: 'reason',
    },
    {
      title: 'Message',
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
    {
      title: 'Count',
      dataIndex: 'count',
      key: 'count',
      width: 80,
    },
    {
      title: 'Last Seen',
      dataIndex: 'lastSeen',
      key: 'lastSeen',
      render: (date: string) => dayjs(date).fromNow(),
    },
  ]

  if (isLoading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center', height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin size="large" tip="Loading pod details..." />
      </div>
    )
  }

  if (!data) {
    return (
      <div style={{ padding: '24px' }}>
        <Alert message="Pod not found" type="error" />
      </div>
    )
  }

  const pod: PodDetail = data

  // Convert labels and annotations to arrays for display
  const labels = Object.entries(pod.labels || {}).map(([key, value]) => ({ key, value }))
  const annotations = Object.entries(pod.annotations || {}).map(([key, value]) => ({ key, value }))

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/clusters/${clusterId}/workloads`)}>
            Back to Workloads
          </Button>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
            <ContainerOutlined /> Pod: {pod.name}
          </span>
          {getStatusTag(pod)}
        </Space>
        <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
          Refresh
        </Button>
      </div>

      {/* Basic Information */}
      <Card title="Basic Information" style={{ marginBottom: '16px' }}>
        <Descriptions column={2} bordered size="small">
          <Descriptions.Item label="Namespace">{pod.namespace}</Descriptions.Item>
          <Descriptions.Item label="Name">{pod.name}</Descriptions.Item>
          <Descriptions.Item label="Status">{getStatusTag(pod)}</Descriptions.Item>
          <Descriptions.Item label="Phase">{pod.phase}</Descriptions.Item>
          <Descriptions.Item label="Pod IP">{pod.podIp || '-'}</Descriptions.Item>
          <Descriptions.Item label="Host IP">{pod.hostIp || '-'}</Descriptions.Item>
          <Descriptions.Item label="Node">{pod.nodeName}</Descriptions.Item>
          <Descriptions.Item label="QoS Class">{pod.qosClass}</Descriptions.Item>
          <Descriptions.Item label="Service Account">{pod.serviceAccount}</Descriptions.Item>
          <Descriptions.Item label="Restart Policy">{pod.restartPolicy}</Descriptions.Item>
          <Descriptions.Item label="DNS Policy">{pod.dnsPolicy}</Descriptions.Item>
          <Descriptions.Item label="Ready">{pod.ready ? 'Yes' : 'No'}</Descriptions.Item>
          <Descriptions.Item label="Total Restarts">{pod.restartCount}</Descriptions.Item>
          <Descriptions.Item label="Created">{dayjs(pod.createdAt).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
          <Descriptions.Item label="Started">{pod.startTime ? dayjs(pod.startTime).format('YYYY-MM-DD HH:mm:ss') : '-'}</Descriptions.Item>
          {pod.ownerType && pod.ownerName && (
            <>
              <Descriptions.Item label="Owner Type">{pod.ownerType}</Descriptions.Item>
              <Descriptions.Item label="Owner Name">{pod.ownerName}</Descriptions.Item>
            </>
          )}
        </Descriptions>
      </Card>

      {/* Tabs */}
      <Card>
        <Tabs items={[
          {
            key: 'containers',
            label: `Containers (${pod.containers.length})`,
            children: (
              <Table
                columns={containerColumns}
                dataSource={pod.containers}
                rowKey="name"
                pagination={false}
                size="small"
              />
            ),
          },
          {
            key: 'events',
            label: `Events (${pod.events.length})`,
            children: (
              <Table
                columns={eventColumns}
                dataSource={pod.events}
                rowKey={(record) => `${record.type}-${record.reason}-${record.lastSeen}`}
                pagination={false}
                size="small"
              />
            ),
          },
          {
            key: 'labels',
            label: `Labels (${labels.length})`,
            children: (
              <Table
                columns={[
                  { title: 'Key', dataIndex: 'key', key: 'key' },
                  { title: 'Value', dataIndex: 'value', key: 'value', ellipsis: true },
                ]}
                dataSource={labels}
                rowKey="key"
                pagination={false}
                size="small"
              />
            ),
          },
          {
            key: 'annotations',
            label: `Annotations (${annotations.length})`,
            children: (
              <Table
                columns={[
                  { title: 'Key', dataIndex: 'key', key: 'key' },
                  { title: 'Value', dataIndex: 'value', key: 'value', ellipsis: true },
                ]}
                dataSource={annotations}
                rowKey="key"
                pagination={false}
                size="small"
              />
            ),
          },
        ]} />
      </Card>

      {/* Logs Modal */}
      <Modal
        title={`Logs: ${logsModal.containerName}`}
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

      {/* Streaming Logs Modal */}
      <Modal
        title={`Streaming Logs: ${pod.name}`}
        open={streamingLogsModal.visible}
        onCancel={() => setStreamingLogsModal({ visible: false, containerName: '' })}
        footer={null}
        width={1000}
        style={{ top: 20 }}
        styles={{ body: { height: 'calc(100vh - 200px)' } }}
      >
        <StreamingLogs
          clusterId={clusterId!}
          namespace={namespace!}
          podName={podName!}
          containerName={streamingLogsModal.containerName}
          visible={streamingLogsModal.visible}
          onClose={() => setStreamingLogsModal({ visible: false, containerName: '' })}
        />
      </Modal>

      {/* Terminal Modal */}
      <Modal
        title={`Terminal: ${pod.name}`}
        open={terminalModal.visible}
        onCancel={() => setTerminalModal({ visible: false, containerName: '' })}
        footer={null}
        width={1000}
        style={{ top: 20 }}
        styles={{ body: { height: 'calc(100vh - 200px)' } }}
      >
        <PodTerminal
          clusterId={clusterId!}
          namespace={namespace!}
          podName={podName!}
          visible={terminalModal.visible}
          onClose={() => setTerminalModal({ visible: false, containerName: '' })}
        />
      </Modal>
    </div>
  )
}

export default PodDetailPage
