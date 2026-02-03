import { useState, useRef, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Button, Space, message, Spin } from 'antd'
import { ArrowLeftOutlined, DisconnectOutlined, ReloadOutlined } from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import SSHTerminal, { SSHTerminalRef, SSHConnectConfig } from '../components/SSHTerminal'
import { SSHConnectModal } from '../components/SSHConnectModal'
import { hostApi } from '../api/host'

const SSHTerminalPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const terminalRef = useRef<SSHTerminalRef>(null)

  const [connectModalVisible, setConnectModalVisible] = useState(false)
  const [isConnected, setIsConnected] = useState(false)
  const [isConnecting, setIsConnecting] = useState(false)

  // Get auth token from localStorage
  const token = localStorage.getItem('myops-auth') || ''

  // Fetch host details
  const { data: host, isLoading } = useQuery({
    queryKey: ['host', id],
    queryFn: () => hostApi.getHost(id!),
    enabled: !!id,
  })

  // Show connect modal when page loads
  useEffect(() => {
    if (host && !isConnected) {
      setConnectModalVisible(true)
    }
  }, [host, isConnected])

  const handleConnect = (config: SSHConnectConfig) => {
    setIsConnecting(true)
    setConnectModalVisible(false)

    if (terminalRef.current) {
      terminalRef.current.connect(config)
    }
  }

  const handleDisconnect = () => {
    if (terminalRef.current) {
      terminalRef.current.disconnect()
    }
    setIsConnected(false)
    message.info('Disconnected from SSH session')
  }

  const handleReconnect = () => {
    setConnectModalVisible(true)
  }

  const handleTerminalConnect = () => {
    setIsConnected(true)
    setIsConnecting(false)
    message.success('Connected to SSH session')
  }

  const handleTerminalDisconnect = () => {
    setIsConnected(false)
    setIsConnecting(false)
  }

  const handleTerminalError = () => {
    setIsConnecting(false)
    setIsConnected(false)
  }

  // Loading state
  if (isLoading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center', height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin size="large" tip="Loading host details..." />
      </div>
    )
  }

  if (!host) {
    return (
      <div style={{ padding: '24px' }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/hosts')}>
          Back to Hosts
        </Button>
        <div style={{ marginTop: '16px', color: '#999' }}>Host not found</div>
      </div>
    )
  }

  return (
    <div style={{ padding: '24px', height: 'calc(100vh - 48px)', display: 'flex', flexDirection: 'column' }}>
      {/* Header */}
      <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/hosts')}>
            Back
          </Button>
          <span style={{ fontSize: '18px', fontWeight: 'bold' }}>
            SSH Terminal: {host.hostname || host.ipAddress}
          </span>
          {isConnected && (
            <span style={{ color: '#52c41a', fontSize: '14px' }}>● Connected</span>
          )}
        </Space>

        <Space>
          {isConnected ? (
            <Button danger icon={<DisconnectOutlined />} onClick={handleDisconnect}>
              Disconnect
            </Button>
          ) : (
            <Button icon={<ReloadOutlined />} onClick={handleReconnect}>
              Reconnect
            </Button>
          )}
        </Space>
      </div>

      {/* Terminal */}
      <Card
        styles={{ body: { padding: 0, height: '100%' } }}
        style={{ flex: 1, overflow: 'hidden' }}
      >
        <div style={{ width: '100%', height: '100%', position: 'relative' }}>
          {isConnecting && (
            <div style={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: 'rgba(0, 0, 0, 0.7)',
              zIndex: 10,
              color: '#fff',
            }}>
              <Space direction="vertical" align="center">
                <Spin size="large" />
                <span>Connecting to SSH...</span>
              </Space>
            </div>
          )}

          <SSHTerminal
            ref={terminalRef}
            hostId={id!}
            token={token}
            onConnect={handleTerminalConnect}
            onDisconnect={handleTerminalDisconnect}
            onError={handleTerminalError}
          />
        </div>
      </Card>

      {/* Connect Modal */}
      <SSHConnectModal
        visible={connectModalVisible}
        hostId={id!}
        hostName={host.hostname || host.ipAddress}
        onConnect={handleConnect}
        onCancel={() => {
          setConnectModalVisible(false)
          if (!isConnected) {
            navigate('/hosts')
          }
        }}
        loading={isConnecting}
      />
    </div>
  )
}

export default SSHTerminalPage
