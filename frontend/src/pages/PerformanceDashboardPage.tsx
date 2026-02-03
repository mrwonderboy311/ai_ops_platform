import React, { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button, Space, Card, Row, Col, Statistic, Progress, Select, Tag } from 'antd'
import {
  DashboardOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  CloseCircleOutlined,
  CloudServerOutlined,
  ApiOutlined,
} from '@ant-design/icons'
import { Line } from '@ant-design/plots'
import dayjs from 'dayjs'
import { performanceApi } from '../api/performance'
import type { TrendPoint } from '../types/performance'

const { Option } = Select

export const PerformanceDashboardPage: React.FC = () => {
  const [timeWindow, setTimeWindow] = useState('5m')
  const [metricType, setMetricType] = useState('cpu')

  // Fetch system health
  const { data: healthData, isLoading: healthLoading, refetch: refetchHealth } = useQuery({
    queryKey: ['systemHealth'],
    queryFn: () => performanceApi.getSystemHealth(),
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  // Fetch performance summary
  const { data: summaryData, isLoading: summaryLoading, refetch: refetchSummary } = useQuery({
    queryKey: ['performanceSummary', timeWindow],
    queryFn: () => performanceApi.getPerformanceSummary(timeWindow),
    refetchInterval: 15000, // Refresh every 15 seconds
  })

  // Fetch trend data
  const { data: trendData, isLoading: trendLoading } = useQuery({
    queryKey: ['trendData', metricType],
    queryFn: () => performanceApi.getTrendData({ metricType, points: 50 }),
    refetchInterval: 15000,
  })

  // Handle refresh
  const handleRefresh = () => {
    refetchHealth()
    refetchSummary()
  }

  const health = healthData?.data
  const summary = summaryData?.data
  const trendPoints = trendData?.data.points || []

  // Get health status config
  const getHealthConfig = (status?: string) => {
    switch (status) {
      case 'healthy':
        return { color: '#52c41a', icon: <CheckCircleOutlined />, text: 'Healthy' }
      case 'warning':
        return { color: '#faad14', icon: <WarningOutlined />, text: 'Warning' }
      case 'critical':
        return { color: '#ff4d4f', icon: <CloseCircleOutlined />, text: 'Critical' }
      default:
        return { color: '#d9d9d9', icon: null, text: 'Unknown' }
    }
  }

  const healthConfig = getHealthConfig(health?.overallStatus)

  // Prepare chart data
  const chartData = trendPoints.map((point: TrendPoint) => ({
    time: dayjs(point.timestamp * 1000).format('HH:mm:ss'),
    value: point.value,
  }))

  const chartConfig = {
    data: chartData,
    xField: 'time',
    yField: 'value',
    smooth: true,
    height: 200,
    animation: {
      appear: {
        animation: 'path-in',
        duration: 1000,
      },
    },
  }

  return (
    <div style={{ padding: '24px' }}>
      {/* Header */}
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>
            <DashboardOutlined /> Performance Dashboard
          </span>
          {health && (
            <Tag color={healthConfig.color} icon={healthConfig.icon}>
              {healthConfig.text}
            </Tag>
          )}
        </Space>
        <Space>
          <Select value={timeWindow} onChange={setTimeWindow} style={{ width: 120 }}>
            <Option value="5m">5 minutes</Option>
            <Option value="15m">15 minutes</Option>
            <Option value="1h">1 hour</Option>
            <Option value="1d">1 day</Option>
          </Select>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            Refresh
          </Button>
        </Space>
      </div>

      {/* System Health Cards */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card loading={healthLoading}>
            <Statistic
              title="Overall Status"
              value={healthConfig.text}
              prefix={healthConfig.icon}
              valueStyle={{ color: healthConfig.color }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={healthLoading}>
            <Statistic
              title="Hosts"
              value={health?.healthyHosts || 0}
              suffix={`/ ${health?.hostCount || 0}`}
              prefix={<CloudServerOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={healthLoading}>
            <Statistic
              title="Clusters"
              value={health?.healthyClusters || 0}
              suffix={`/ ${health?.clusterCount || 0}`}
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={healthLoading}>
            <Statistic
              title="Issues"
              value={(health?.warningCount || 0) + (health?.criticalCount || 0)}
              prefix={<WarningOutlined />}
              valueStyle={{ color: (health?.warningCount || 0) + (health?.criticalCount || 0) > 0 ? '#cf1322' : '#3f8600' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Performance Metrics */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={12}>
          <Card title="CPU Usage" loading={summaryLoading}>
            <Row gutter={16}>
              <Col span={12}>
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ marginBottom: '8px', color: '#888' }}>Average</div>
                  <Progress
                    type="circle"
                    percent={Math.round(summary?.avgCPUUsage || 0)}
                    size={100}
                  />
                </div>
              </Col>
              <Col span={12}>
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ marginBottom: '8px', color: '#888' }}>Peak</div>
                  <Progress
                    type="circle"
                    percent={Math.round(summary?.maxCPUUsage || 0)}
                    size={100}
                    strokeColor="#cf1322"
                  />
                </div>
              </Col>
            </Row>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="Memory Usage" loading={summaryLoading}>
            <Row gutter={16}>
              <Col span={12}>
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ marginBottom: '8px', color: '#888' }}>Average</div>
                  <Progress
                    type="circle"
                    percent={Math.round(summary?.avgMemoryUsage || 0)}
                    size={100}
                  />
                </div>
              </Col>
              <Col span={12}>
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ marginBottom: '8px', color: '#888' }}>Peak</div>
                  <Progress
                    type="circle"
                    percent={Math.round(summary?.maxMemoryUsage || 0)}
                    size={100}
                    strokeColor="#cf1322"
                  />
                </div>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={12}>
          <Card title="Disk Usage" loading={summaryLoading}>
            <Row gutter={16}>
              <Col span={12}>
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ marginBottom: '8px', color: '#888' }}>Average</div>
                  <Progress
                    type="circle"
                    percent={Math.round(summary?.avgDiskUsage || 0)}
                    size={100}
                  />
                </div>
              </Col>
              <Col span={12}>
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ marginBottom: '8px', color: '#888' }}>Peak</div>
                  <Progress
                    type="circle"
                    percent={Math.round(summary?.maxDiskUsage || 0)}
                    size={100}
                    strokeColor="#cf1322"
                  />
                </div>
              </Col>
            </Row>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="Response & Errors" loading={summaryLoading}>
            <Row gutter={16}>
              <Col span={12}>
                <Statistic
                  title="Avg Response Time"
                  value={summary?.avgResponseTime?.toFixed(2) || '0.00'}
                  suffix="ms"
                  valueStyle={{ color: summary?.avgResponseTime && summary.avgResponseTime > 500 ? '#cf1322' : '#3f8600' }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="Error Rate"
                  value={summary?.errorRate?.toFixed(2) || '0.00'}
                  suffix="%"
                  valueStyle={{ color: summary?.errorRate && summary.errorRate > 1 ? '#cf1322' : '#3f8600' }}
                />
              </Col>
            </Row>
            <div style={{ marginTop: '16px' }}>
              <Statistic
                title="Throughput"
                value={summary?.throughput?.toFixed(2) || '0.00'}
                suffix="req/s"
              />
            </div>
          </Card>
        </Col>
      </Row>

      {/* Trend Chart */}
      <Card
        title={
          <Space>
            <span>Trend</span>
            <Select value={metricType} onChange={setMetricType} style={{ width: 120 }}>
              <Option value="cpu">CPU</Option>
              <Option value="memory">Memory</Option>
              <Option value="disk">Disk</Option>
              <Option value="response_time">Response Time</Option>
            </Select>
          </Space>
        }
        loading={trendLoading}
      >
        {chartData.length > 0 ? (
          <Line {...chartConfig} />
        ) : (
          <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>
            No trend data available
          </div>
        )}
      </Card>
    </div>
  )
}

export default PerformanceDashboardPage
