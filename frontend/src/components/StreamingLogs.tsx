import { useEffect, useRef, useState } from 'react'
import { Button, Space, Tag, Switch, InputNumber, Select, Typography, message } from 'antd'
import {
  CloseOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  DeleteOutlined,
  CopyOutlined,
  DownloadOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'

const { Text } = Typography
const { Option } = Select

interface LogEntry {
  timestamp: string
  content: string
}

interface StreamingLogsProps {
  clusterId: string
  namespace: string
  podName: string
  containerName?: string
  visible: boolean
  onClose: () => void
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const StreamingLogs: React.FC<StreamingLogsProps> = ({
  clusterId,
  namespace,
  podName,
  containerName,
  visible,
  onClose,
}) => {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [isStreaming, setIsStreaming] = useState(false)
  const [autoScroll, setAutoScroll] = useState(true)
  const [tailLines, setTailLines] = useState(100)
  const [selectedContainer, setSelectedContainer] = useState<string>(containerName || '')
  const [containers, setContainers] = useState<string[]>([])

  const wsRef = useRef<WebSocket | null>(null)
  const logContainerRef = useRef<HTMLDivElement>(null)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout>>()

  // Fetch containers for the pod
  useEffect(() => {
    const fetchContainers = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/v1/clusters/${clusterId}/namespaces/${namespace}/pods/${podName}/detail`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
          },
        })
        if (response.ok) {
          const data = await response.json()
          if (data.data?.containers) {
            const containerNames = data.data.containers.map((c: any) => c.name)
            setContainers(containerNames)
            if (!selectedContainer && containerNames.length > 0) {
              setSelectedContainer(containerNames[0])
            }
          }
        }
      } catch (error) {
        console.error('Failed to fetch containers:', error)
      }
    }

    if (visible && clusterId && namespace && podName) {
      fetchContainers()
    }
  }, [visible, clusterId, namespace, podName])

  // Auto-scroll to bottom when logs are updated
  useEffect(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight
    }
  }, [logs, autoScroll])

  // Connect to websocket for streaming logs
  const startStreaming = () => {
    if (!selectedContainer) {
      message.warning('Please select a container first')
      return
    }

    if (wsRef.current) {
      wsRef.current.close()
    }

    const wsUrl = `${API_BASE_URL.replace('http', 'ws')}/api/v1/clusters/pod-logs/ws?clusterId=${clusterId}&namespace=${namespace}&podName=${podName}&container=${selectedContainer}&tailLines=${tailLines}&follow=true`

    wsRef.current = new WebSocket(wsUrl)

    wsRef.current.onopen = () => {
      setIsStreaming(true)
      setLogs([]) // Clear previous logs
    }

    wsRef.current.onmessage = (event) => {
      const data = event.data
      if (data.startsWith('Error:')) {
        message.error(data)
        setIsStreaming(false)
        return
      }

      // Parse log entries (each line may have a timestamp prefix)
      const lines: string[] = data.split('\n').filter((line: string) => line.trim())
      const newEntries: LogEntry[] = []

      for (const line of lines) {
        // Kubernetes logs may have timestamp prefix like "2024-01-01T12:00:00.000000000Z"
        const timestampMatch = line.match(/^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s*(.*)$/)
        if (timestampMatch) {
          newEntries.push({
            timestamp: timestampMatch[1],
            content: timestampMatch[2],
          })
        } else {
          newEntries.push({
            timestamp: dayjs().format(),
            content: line,
          })
        }
      }

      if (newEntries.length > 0) {
        setLogs((prev) => [...prev, ...newEntries])
      }
    }

    wsRef.current.onerror = (error) => {
      console.error('WebSocket error:', error)
      message.error('WebSocket connection error')
      setIsStreaming(false)
    }

    wsRef.current.onclose = () => {
      setIsStreaming(false)
    }
  }

  const stopStreaming = () => {
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    setIsStreaming(false)

    // Clear reconnect timeout if exists
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
  }

  // Toggle streaming
  const toggleStreaming = () => {
    if (isStreaming) {
      stopStreaming()
    } else {
      startStreaming()
    }
  }

  // Clear logs
  const clearLogs = () => {
    setLogs([])
  }

  // Copy logs to clipboard
  const copyLogs = () => {
    const logText = logs.map((entry) => entry.content).join('\n')
    navigator.clipboard.writeText(logText)
    message.success('Logs copied to clipboard')
  }

  // Download logs
  const downloadLogs = () => {
    const logText = logs.map((entry) => entry.content).join('\n')
    const blob = new Blob([logText], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${podName}-${selectedContainer}-logs.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  // Close handler
  const handleClose = () => {
    stopStreaming()
    onClose()
  }

  if (!visible) {
    return null
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      {/* Header with controls */}
      <div style={{ marginBottom: '12px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          {containers.length > 1 && (
            <Select
              value={selectedContainer}
              onChange={(value) => {
                setSelectedContainer(value as string)
                if (isStreaming) {
                  stopStreaming()
                }
              }}
              style={{ width: 200 }}
              placeholder="Select container"
            >
              {containers.map((c) => (
                <Option key={c} value={c}>{c}</Option>
              ))}
            </Select>
          )}
          <Button
            type="primary"
            icon={isStreaming ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
            onClick={toggleStreaming}
            disabled={!selectedContainer}
          >
            {isStreaming ? 'Pause' : 'Stream'}
          </Button>
          <Button icon={<DeleteOutlined />} onClick={clearLogs} disabled={isStreaming}>
            Clear
          </Button>
          <Switch checked={autoScroll} onChange={setAutoScroll} checkedChildren="Auto-scroll" />
        </Space>
        <Space>
          <InputNumber
            value={tailLines}
            onChange={(value) => setTailLines(value || 100)}
            min={1}
            max={10000}
            addonAfter="lines"
            disabled={isStreaming}
            style={{ width: 130 }}
          />
          <Button icon={<CopyOutlined />} onClick={copyLogs}>
            Copy
          </Button>
          <Button icon={<DownloadOutlined />} onClick={downloadLogs} disabled={logs.length === 0}>
            Download
          </Button>
          <Button icon={<CloseOutlined />} onClick={handleClose}>
            Close
          </Button>
        </Space>
      </div>

      {/* Status */}
      {isStreaming && (
        <div style={{ marginBottom: '8px' }}>
          <Tag color="blue" icon={<PlayCircleOutlined />}>
            Streaming live...
          </Tag>
          <Text type="secondary" style={{ marginLeft: '8px' }}>
            ({logs.length} lines)
          </Text>
        </div>
      )}

      {/* Logs display */}
      <div
        ref={logContainerRef}
        style={{
          flex: 1,
          overflow: 'auto',
          background: '#1e1e1e',
          color: '#d4d4d4',
          padding: '12px',
          borderRadius: '4px',
          fontFamily: 'monospace',
          fontSize: '12px',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-all',
        }}
      >
        {logs.length === 0 ? (
          <div style={{ color: '#888', textAlign: 'center', padding: '40px' }}>
            {isStreaming ? 'Waiting for logs...' : 'Click Stream to start streaming logs'}
          </div>
        ) : (
          logs.map((entry, index) => (
            <div key={index} style={{ marginBottom: '2px' }}>
              <span style={{ color: '#569cd6' }}>{entry.timestamp}</span>
              <span style={{ marginLeft: '8px' }}>{entry.content}</span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

export default StreamingLogs
