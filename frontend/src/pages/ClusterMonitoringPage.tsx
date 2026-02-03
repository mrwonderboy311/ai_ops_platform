import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Button, Space, message, Card, Row, Col, Statistic, Progress, Select, Tag, Spin, Alert } from 'antd'
import {
  ArrowLeftOutlined,
  ReloadOutlined,
  CloudServerOutlined,
  ClusterOutlined,
  ContainerOutlined,
  NodeIndexOutlined,
} from '@ant-design/icons'
import { clusterMonitoringApi } from '../api/clusterMonitoring'
import type { NodeMetric } from '../types/clusterMonitoring'

const { Option } = Select

export const ClusterMonitoringPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [duration, setDuration] = useState<string>('1h')
  const [autoRefresh, setAutoRefresh] = useState(false)

  // Fetch cluster metrics summary
  const { data: summary, isLoading: summaryLoading, refetch: refetchSummary, error: summaryError } = useQuery({
    queryKey: ['clusterMetricsSummary', id],
    queryFn: () => clusterMonitoringApi.getClusterMetricsSummary(id!),
    enabled: !!id,
    refetchInterval: autoRefresh ? 10000 : false,
  })

  // Fetch node metrics
  const { data: nodeMetrics, isLoading: nodesLoading, refetch: refetchNodes } = useQuery({
    queryKey: ['nodeMetrics', id, duration],
    queryFn: () => clusterMonitoringApi.getNodeMetrics(id!, duration),
    enabled: !!id,
    refetchInterval: autoRefresh ? 10000 : false,
  })

  // Handle refresh
  const handleRefresh = () => {
    refetchSummary()
    refetchNodes()
    clusterMonitoringApi.refreshMetrics(id!).catch(() => {
      // Ignore error, metrics refresh is asynchronous
    })
    message.success('Metrics refresh initiated')
  }

  // Handle auto-refresh toggle
  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(() => {
        refetchSummary()
        refetchNodes()
      }, 10000)
      return () => clearInterval(interval)
    }
  }, [autoRefresh, refetchSummary, refetchNodes])

  if (summaryLoading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center', height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin size="large" tip="Loading monitoring data..." />
      </div>
    )
  }

  if (summaryError || !summary) {
    return (
      <div style={{ padding: '24px' }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/clusters')}>
          Back to Clusters
        </Button>
        <Alert
          style={{ marginTop: '16px' }}
          message="Failed to load monitoring data"
          description={(summaryError as Error)?.message || 'Unknown error'}
          type="error"
          showIcon
        />
      </div>
    )
  }

  // Get latest node metrics (unique nodes with latest data)
  const latestNodeMetrics = React.useMemo(() => {
    if (!nodeMetrics) return []
    const nodeMap = new Map<string, NodeMetric>()
    nodeMetrics.forEach((metric) => {
      const existing = nodeMap.get(metric.nodeName)
      if (!existing || metric.timestamp > existing.timestamp) {
        nodeMap.set(metric.nodeName, metric)
      }
    })
    return Array.from(nodeMap.values())
  }, [nodeMetrics])

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/clusters')}>
            Back to Clusters
          </Button>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
            <CloudServerOutlined /> Cluster Monitoring
          </span>
        </Space>
        <Space>
          <Select
            value={duration}
            onChange={setDuration}
            style={{ width: 120 }}
          >
            <Option value="15m">15 minutes</Option>
            <Option value="1h">1 hour</Option>
            <Option value="6h">6 hours</Option>
            <Option value="24h">24 hours</Option>
            <Option value="7d">7 days</Option>
          </Select>
          <Button onClick={() => setAutoRefresh(!autoRefresh)}>
            {autoRefresh ? 'Disable' : 'Enable'} Auto-Refresh
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            Refresh
          </Button>
        </Space>
      </div>

      {/* Auto-refresh indicator */}
      {autoRefresh && (
        <Alert
          style={{ marginBottom: '16px' }}
          message="Auto-refreshing"
          description="Metrics are being updated automatically every 10 seconds."
          type="info"
          showIcon
          closable
          onClose={() => setAutoRefresh(false)}
        />
      )}

      {/* Cluster Overview */}
      <Card title="Cluster Overview" style={{ marginBottom: '16px' }}>
        <Row gutter={16}>
          <Col span={6}>
            <Statistic
              title="CPU Usage"
              value={summary.cpuUsagePercent.toFixed(2)}
              suffix="%"
              prefix={<ClusterOutlined />}
              valueStyle={{ color: summary.cpuUsagePercent > 80 ? '#cf1322' : summary.cpuUsagePercent > 60 ? '#faad14' : '#3f8600' }}
            />
            <Progress
              percent={Math.round(summary.cpuUsagePercent)}
              status={summary.cpuUsagePercent > 80 ? 'exception' : summary.cpuUsagePercent > 60 ? 'normal' : 'success'}
              showInfo={false}
              style={{ marginTop: '8px' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="Memory Usage"
              value={summary.memoryUsagePercent.toFixed(2)}
              suffix="%"
              valueStyle={{ color: summary.memoryUsagePercent > 80 ? '#cf1322' : summary.memoryUsagePercent > 60 ? '#faad14' : '#3f8600' }}
            />
            <Progress
              percent={Math.round(summary.memoryUsagePercent)}
              status={summary.memoryUsagePercent > 80 ? 'exception' : summary.memoryUsagePercent > 60 ? 'normal' : 'success'}
              showInfo={false}
              style={{ marginTop: '8px' }}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="Nodes"
              value={summary.readyNodeCount}
              suffix={`/ ${summary.nodeCount}`}
              prefix={<NodeIndexOutlined />}
              valueStyle={{ color: summary.readyNodeCount === summary.nodeCount ? '#3f8600' : '#faad14' }}
            />
            <div style={{ marginTop: '8px', fontSize: '12px', color: '#888' }}>
              {summary.nodeCount - summary.readyNodeCount} Not Ready
            </div>
          </Col>
          <Col span={6}>
            <Statistic
              title="Pods"
              value={summary.runningPodCount}
              suffix={`/ ${summary.podCount}`}
              prefix={<ContainerOutlined />}
              valueStyle={{ color: summary.runningPodCount === summary.podCount ? '#3f8600' : '#faad14' }}
            />
            <div style={{ marginTop: '8px', fontSize: '12px', color: '#888' }}>
              {summary.pendingPodCount} Pending · {summary.failedPodCount} Failed
            </div>
          </Col>
        </Row>
      </Card>

      {/* Node Metrics */}
      <Card title={`Node Metrics (${latestNodeMetrics.length})`} loading={nodesLoading}>
        <Row gutter={[16, 16]}>
          {latestNodeMetrics.map((node) => (
            <Col span={8} key={node.id}>
              <Card
                size="small"
                title={
                  <Space>
                    <NodeIndexOutlined />
                    {node.nodeName}
                  </Space>
                }
                extra={
                  <Tag color={node.ready ? 'success' : 'error'}>
                    {node.status}
                  </Tag>
                }
              >
                <Row gutter={8}>
                  <Col span={12}>
                    <div style={{ fontSize: '12px', color: '#888' }}>CPU</div>
                    <Progress
                      percent={Math.round(node.cpuUsagePercent)}
                      size="small"
                      status={node.cpuUsagePercent > 80 ? 'exception' : 'normal'}
                    />
                  </Col>
                  <Col span={12}>
                    <div style={{ fontSize: '12px', color: '#888' }}>Memory</div>
                    <Progress
                      percent={Math.round((node.memoryUsageBytes / node.memoryTotalBytes) * 100)}
                      size="small"
                      status={node.memoryUsageBytes / node.memoryTotalBytes > 0.8 ? 'exception' : 'normal'}
                    />
                  </Col>
                </Row>
                <div style={{ marginTop: '12px', fontSize: '12px', color: '#888' }}>
                  <div>Pods: {node.podCount}</div>
                  {node.diskTotalBytes > 0 && (
                    <div>Disk: {((node.diskUsageBytes / node.diskTotalBytes) * 100).toFixed(1)}%</div>
                  )}
                </div>
              </Card>
            </Col>
          ))}
        </Row>
        {latestNodeMetrics.length === 0 && (
          <div style={{ textAlign: 'center', padding: '24px', color: '#888' }}>
            No node metrics available
          </div>
        )}
      </Card>
    </div>
  )
}

export default ClusterMonitoringPage
