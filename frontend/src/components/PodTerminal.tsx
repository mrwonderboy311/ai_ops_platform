import { useEffect, useRef, useState } from 'react'
import { Button, Space, Select, message } from 'antd'
import { CloseOutlined, LoadingOutlined } from '@ant-design/icons'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'

const { Option } = Select

interface PodTerminalProps {
  clusterId: string
  namespace: string
  podName: string
  visible: boolean
  onClose: () => void
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const PodTerminal: React.FC<PodTerminalProps> = ({
  clusterId,
  namespace,
  podName,
  visible,
  onClose,
}) => {
  const [containers, setContainers] = useState<string[]>([])
  const [selectedContainer, setSelectedContainer] = useState<string>('')
  const [selectedShell, setSelectedShell] = useState<string>('/bin/sh')
  const [isConnecting, setIsConnecting] = useState(false)
  const [isConnected, setIsConnected] = useState(false)

  const terminalRef = useRef<HTMLDivElement>(null)
  const terminalRef2 = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const wsRef = useRef<WebSocket | null>(null)

  const shellOptions = ['/bin/sh', '/bin/bash', '/bin/ash', '/sh']

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

  // Initialize terminal
  useEffect(() => {
    if (!visible || !terminalRef.current) {
      return
    }

    // Create terminal
    const terminal = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#d4d4d4',
        black: '#000000',
        red: '#cd3131',
        green: '#0dbc79',
        yellow: '#e5e510',
        blue: '#2472c8',
        magenta: '#bc3fbc',
        cyan: '#11a8cd',
        white: '#e5e5e5',
        brightBlack: '#666666',
        brightRed: '#f14c4c',
        brightGreen: '#23d18b',
        brightYellow: '#f5f543',
        brightBlue: '#3b8eea',
        brightMagenta: '#d670d6',
        brightCyan: '#29b8db',
        brightWhite: '#ffffff',
      },
    })

    // Create fit addon
    const fitAddon = new FitAddon()
    terminal.loadAddon(fitAddon)

    // Mount terminal
    terminal.open(terminalRef.current)
    fitAddon.fit()

    terminalRef2.current = terminal
    fitAddonRef.current = fitAddon

    // Handle terminal resize
    const handleResize = () => {
      if (fitAddonRef.current) {
        fitAddonRef.current.fit()
      }
    }

    window.addEventListener('resize', handleResize)

    // Write welcome message
    terminal.writeln('\x1b[1;34mMyOps Pod Terminal\x1b[0m')
    terminal.writeln('Select a container and shell, then click "Connect" to start.')
    terminal.writeln('')

    return () => {
      window.removeEventListener('resize', handleResize)
      terminal.dispose()
      terminalRef2.current = null
      fitAddonRef.current = null
    }
  }, [visible])

  // Connect to websocket
  const connect = () => {
    if (!selectedContainer) {
      message.warning('Please select a container first')
      return
    }

    if (wsRef.current) {
      wsRef.current.close()
    }

    setIsConnecting(true)

    const wsUrl = `${API_BASE_URL.replace('http', 'ws')}/api/v1/clusters/pod-terminal/ws?clusterId=${clusterId}&namespace=${namespace}&podName=${podName}&container=${selectedContainer}&shell=${selectedShell}`

    wsRef.current = new WebSocket(wsUrl)

    wsRef.current.onopen = () => {
      setIsConnecting(false)
      setIsConnected(true)

      if (terminalRef2.current) {
        terminalRef2.current.writeln(`\x1b[1;32mConnected to ${selectedContainer}\x1b[0m`)
        terminalRef2.current.writeln('')
      }
    }

    wsRef.current.onmessage = (event) => {
      const data = event.data
      try {
        const msg = JSON.parse(data)
        if (msg.type === 'output') {
          if (terminalRef2.current) {
            terminalRef2.current.write(msg.data)
          }
        } else if (msg.type === 'error') {
          if (terminalRef2.current) {
            terminalRef2.current.writeln(`\x1b[1;31mError: ${msg.data}\x1b[0m`)
          }
          message.error(msg.data)
        }
      } catch {
        // Not JSON, write as-is
        if (terminalRef2.current) {
          terminalRef2.current.write(data)
        }
      }
    }

    wsRef.current.onerror = (error) => {
      console.error('WebSocket error:', error)
      setIsConnecting(false)
      setIsConnected(false)
      message.error('WebSocket connection error')
    }

    wsRef.current.onclose = () => {
      setIsConnecting(false)
      setIsConnected(false)
      if (terminalRef2.current) {
        terminalRef2.current.writeln('\x1b[1;33mConnection closed\x1b[0m')
      }
    }
  }

  // Handle terminal input
  useEffect(() => {
    if (!terminalRef2.current || !isConnected) {
      return
    }

    const handleData = (data: string) => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({
          type: 'input',
          data: data,
        }))
      }
    }

    terminalRef2.current.onData(handleData)

    return () => {
      if (terminalRef2.current) {
        terminalRef2.current.onData(() => {})
      }
    }
  }, [isConnected])

  // Handle terminal resize
  useEffect(() => {
    if (!terminalRef2.current || !isConnected) {
      return
    }

    const handleResize = () => {
      if (fitAddonRef.current && wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        const dims = fitAddonRef.current.proposeDimensions()
        if (dims && dims.cols && dims.rows) {
          wsRef.current.send(JSON.stringify({
            type: 'resize',
            rows: dims.rows,
            cols: dims.cols,
          }))
        }
      }
    }

    const resizeObserver = new ResizeObserver(handleResize)
    if (terminalRef.current) {
      resizeObserver.observe(terminalRef.current)
    }

    // Initial resize
    handleResize()

    return () => {
      resizeObserver.disconnect()
    }
  }, [isConnected])

  // Disconnect on close
  const handleClose = () => {
    if (wsRef.current) {
      wsRef.current.close()
    }
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
                if (isConnected) {
                  if (wsRef.current) {
                    wsRef.current.close()
                  }
                  setIsConnected(false)
                }
              }}
              style={{ width: 200 }}
              placeholder="Select container"
              disabled={isConnected}
            >
              {containers.map((c) => (
                <Option key={c} value={c}>{c}</Option>
              ))}
            </Select>
          )}
          <Select
            value={selectedShell}
            onChange={(value) => setSelectedShell(value as string)}
            style={{ width: 120 }}
            disabled={isConnected}
          >
            {shellOptions.map((shell) => (
              <Option key={shell} value={shell}>{shell}</Option>
            ))}
          </Select>
          <Button
            type="primary"
            icon={isConnecting ? <LoadingOutlined /> : isConnected ? <div style={{ width: 8, height: 8, borderRadius: '50%', background: '#52c41a' }} /> : undefined}
            onClick={connect}
            disabled={!selectedContainer || isConnecting || isConnected}
            loading={isConnecting}
          >
            {isConnected ? 'Connected' : isConnecting ? 'Connecting...' : 'Connect'}
          </Button>
          {isConnected && (
            <Button danger onClick={() => {
              if (wsRef.current) {
                wsRef.current.close()
              }
              setIsConnected(false)
            }}>
              Disconnect
            </Button>
          )}
        </Space>
        <Space>
          <Button icon={<CloseOutlined />} onClick={handleClose}>
            Close
          </Button>
        </Space>
      </div>

      {/* Status */}
      {isConnected && (
        <div style={{ marginBottom: '8px' }}>
          <span style={{ color: '#52c41a', fontSize: '12px' }}>
            ● Connected to {selectedContainer}
          </span>
        </div>
      )}

      {/* Terminal */}
      <div
        ref={terminalRef}
        style={{
          flex: 1,
          background: '#1e1e1e',
          borderRadius: '4px',
          overflow: 'hidden',
          minHeight: '400px',
        }}
      />
    </div>
  )
}

export default PodTerminal
