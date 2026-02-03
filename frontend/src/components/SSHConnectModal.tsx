import { useState } from 'react'
import { Modal, Form, Input, Radio } from 'antd'
import { LockOutlined } from '@ant-design/icons'

export interface SSHConnectConfig {
  hostId: string
  username: string
  authType: 'password' | 'key'
  password?: string
  privateKey?: string
  rows: number
  cols: number
}

interface SSHConnectModalProps {
  visible: boolean
  hostId: string
  hostName: string
  onConnect: (config: SSHConnectConfig) => void
  onCancel: () => void
  loading?: boolean
}

export const SSHConnectModal: React.FC<SSHConnectModalProps> = ({
  visible,
  hostId,
  hostName,
  onConnect,
  onCancel,
  loading = false,
}) => {
  const [form] = Form.useForm()
  const [authType, setAuthType] = useState<'password' | 'key'>('password')

  const handleOk = async () => {
    try {
      const values = await form.validateFields()

      const config: SSHConnectConfig = {
        hostId,
        username: values.username || 'root',
        authType,
        password: authType === 'password' ? values.password : undefined,
        privateKey: authType === 'key' ? values.privateKey : undefined,
        rows: 24,
        cols: 80,
      }

      onConnect(config)
    } catch (err) {
      // Validation failed
    }
  }

  const handleCancel = () => {
    form.resetFields()
    onCancel()
  }

  return (
    <Modal
      title={`Connect to SSH: ${hostName}`}
      open={visible}
      onOk={handleOk}
      onCancel={handleCancel}
      okText="Connect"
      cancelText="Cancel"
      confirmLoading={loading}
      width={500}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          authType: 'password',
          username: 'root',
        }}
      >
        <Form.Item
          label="Username"
          name="username"
          rules={[{ required: true, message: 'Please enter username' }]}
        >
          <Input
            prefix={<LockOutlined />}
            placeholder="Enter username (default: root)"
          />
        </Form.Item>

        <Form.Item label="Authentication Method">
          <Radio.Group
            value={authType}
            onChange={(e) => setAuthType(e.target.value)}
          >
            <Radio value="password">Password</Radio>
            <Radio value="key">Private Key</Radio>
          </Radio.Group>
        </Form.Item>

        {authType === 'password' && (
          <Form.Item
            label="Password"
            name="password"
            rules={[{ required: true, message: 'Please enter password' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="Enter SSH password"
            />
          </Form.Item>
        )}

        {authType === 'key' && (
          <Form.Item
            label="Private Key"
            name="privateKey"
            rules={[{ required: true, message: 'Please enter private key' }]}
            extra="Paste your private key content (PEM format)"
          >
            <Input.TextArea
              placeholder="-----BEGIN RSA PRIVATE KEY-----
...
-----END RSA PRIVATE KEY-----"
              rows={8}
              style={{ fontFamily: 'monospace', fontSize: '12px' }}
            />
          </Form.Item>
        )}

        <div style={{ color: '#999', fontSize: '12px', marginTop: '8px' }}>
          <p style={{ margin: 0 }}>
            Terminal size: 24 rows × 80 columns (will auto-adjust)
          </p>
        </div>
      </Form>
    </Modal>
  )
}

export default SSHConnectModal
