import { useEffect, useRef, useState, useCallback, forwardRef, useImperativeHandle } from 'react'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import 'xterm/css/xterm.css'
import { message } from 'antd'

// WebSocket message types
type WSMessageType = 'input' | 'resize' | 'ping' | 'pong' | 'output' | 'connected' | 'error'

interface WSMessage {
  type: WSMessageType
  data?: string
  rows?: number
  cols?: number
  sessionId?: string
  message?: string
  error?: string
}

export interface SSHConnectConfig {
  hostId: string
  username: string
  password?: string
  privateKey?: string
  rows: number
  cols: number
}

export interface SSHTerminalRef {
  connect: (config: SSHConnectConfig) => void
  disconnect: () => void
  isConnected: () => boolean
  isConnecting: () => boolean
}

interface SSHTerminalProps {
  hostId: string
  token: string
  onConnect?: () => void
  onDisconnect?: () => void
  onError?: (error: string) => void
}

export const SSHTerminal = forwardRef<SSHTerminalRef, SSHTerminalProps>(({
  hostId,
  token,
  onConnect,
  onDisconnect,
  onError,
}, ref) => {
  const terminalRef = useRef<HTMLDivElement>(null)
  const terminalRefInstance = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const [isConnected, setIsConnected] = useState(false)
  const [isConnecting, setIsConnecting] = useState(false)

  // Cleanup function
  const cleanup = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    if (terminalRefInstance.current) {
      terminalRefInstance.current.dispose()
      terminalRefInstance.current = null
    }
    setIsConnected(false)
  }, [])

  // Connect to SSH WebSocket
  const connect = useCallback(
    (config: SSHConnectConfig) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        return
      }

      setIsConnecting(true)

      // Add token to WebSocket handshake via query param
      // Note: For WebSocket, we typically pass the token via query param
      const wsUrl = `ws://localhost:8080/ws/ssh/${hostId}?token=${token}`
      const wsWithAuth = new WebSocket(wsUrl)

      wsRef.current = wsWithAuth

      wsWithAuth.onopen = () => {
        console.log('[SSH Terminal] WebSocket connected')

        // Send connect message with credentials
        const connectMsg: WSMessage = {
          type: 'input',
          data: JSON.stringify(config),
        }
        wsWithAuth.send(JSON.stringify(connectMsg))
      }

      wsWithAuth.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data)

          switch (msg.type) {
            case 'connected':
              console.log('[SSH Terminal] SSH session connected:', msg.sessionId)
              setIsConnected(true)
              setIsConnecting(false)
              onConnect?.()

              // Send initial ping
              wsWithAuth.send(JSON.stringify({ type: 'ping' }))
              break

            case 'output':
              if (terminalRefInstance.current && msg.data) {
                terminalRefInstance.current.write(msg.data)
              }
              break

            case 'error':
              console.error('[SSH Terminal] Server error:', msg.message, msg.error)
              message.error(`SSH Error: ${msg.message}`)
              setIsConnecting(false)
              onError?.(msg.error || msg.message || 'Unknown error')
              cleanup()
              break

            case 'pong':
              // Handle pong - keep connection alive
              break

            default:
              console.warn('[SSH Terminal] Unknown message type:', msg.type)
          }
        } catch (err) {
          console.error('[SSH Terminal] Failed to parse message:', err)
        }
      }

      wsWithAuth.onerror = (event) => {
        console.error('[SSH Terminal] WebSocket error:', event)
        message.error('WebSocket connection error')
        setIsConnecting(false)
        onError?.('WebSocket connection error')
        cleanup()
      }

      wsWithAuth.onclose = (event) => {
        console.log('[SSH Terminal] WebSocket closed:', event.code, event.reason)
        setIsConnected(false)
        setIsConnecting(false)
        onDisconnect?.()

        if (event.code !== 1000) {
          // Non-normal close - could reconnect here if desired
          message.warning(`Connection closed: ${event.reason || 'Unknown reason'}`)
        }
      }
    },
    [hostId, token, onConnect, onDisconnect, onError, cleanup]
  )

  // Initialize terminal
  useEffect(() => {
    if (!terminalRef.current) return

    // Create terminal instance
    const terminal = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#ffffff',
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
      scrollback: 1000,
      tabStopWidth: 4,
    })

    // Create and load addons
    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()

    terminal.loadAddon(fitAddon)
    terminal.loadAddon(webLinksAddon)

    // Open terminal in the DOM
    terminal.open(terminalRef.current)
    fitAddon.fit()

    // Store refs
    terminalRefInstance.current = terminal
    fitAddonRef.current = fitAddon

    // Handle user input
    terminal.onData((data) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        const msg: WSMessage = {
          type: 'input',
          data: data,
        }
        wsRef.current.send(JSON.stringify(msg))
      }
    })

    // Handle terminal resize
    const handleResize = () => {
      if (fitAddonRef.current) {
        fitAddonRef.current.fit()
        const dims = fitAddonRef.current.proposeDimensions()

        if (dims && dims.rows && dims.cols && wsRef.current?.readyState === WebSocket.OPEN) {
          const msg: WSMessage = {
            type: 'resize',
            rows: dims.rows,
            cols: dims.cols,
          }
          wsRef.current.send(JSON.stringify(msg))
        }
      }
    }

    // Listen for window resize
    window.addEventListener('resize', handleResize)

    // Cleanup on unmount
    return () => {
      window.removeEventListener('resize', handleResize)
      cleanup()
      terminal.dispose()
    }
  }, [cleanup])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      cleanup()
    }
  }, [cleanup])

  // Expose methods via ref
  useImperativeHandle(ref, () => ({
    connect,
    disconnect: cleanup,
    isConnected: () => isConnected,
    isConnecting: () => isConnecting,
  }), [connect, cleanup, isConnected, isConnecting])

  return (
    <div
      ref={terminalRef}
      style={{
        width: '100%',
        height: '100%',
        backgroundColor: '#1e1e1e',
        borderRadius: '4px',
        overflow: 'hidden',
      }}
    />
  )
})

SSHTerminal.displayName = 'SSHTerminal'

export default SSHTerminal
