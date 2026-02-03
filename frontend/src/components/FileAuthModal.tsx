import { useState } from 'react'
import { Modal, Form, Input, Radio } from 'antd'
import { LockOutlined } from '@ant-design/icons'

interface FileAuthModalProps {
  visible: boolean
  hostName: string
  onSubmit: (credentials: { username: string; password: string; key: string }) => void
  onCancel: () => void
}

export const FileAuthModal: React.FC<FileAuthModalProps> = ({
  visible,
  hostName,
  onSubmit,
  onCancel,
}) => {
  const [form] = Form.useForm()
  const [authType, setAuthType] = useState<'password' | 'key'>('password')

  const handleOk = async () => {
    try {
      const values = await form.validateFields()
      onSubmit({
        username: values.username || 'root',
        password: authType === 'password' ? values.password : '',
        key: authType === 'key' ? values.key : '',
      })
      form.resetFields()
    } catch {
      // Validation failed
    }
  }

  return (
    <Modal
      title={`Connect to ${hostName}`}
      open={visible}
      onOk={handleOk}
      onCancel={onCancel}
      okText="Connect"
      cancelText="Cancel"
      width={450}
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
            name="key"
            rules={[{ required: true, message: 'Please enter private key' }]}
            extra="Paste your private key content (PEM format)"
          >
            <Input.TextArea
              placeholder="-----BEGIN RSA PRIVATE KEY-----
...
-----END RSA PRIVATE KEY-----"
              rows={6}
              style={{ fontFamily: 'monospace', fontSize: '12px' }}
            />
          </Form.Item>
        )}
      </Form>
    </Modal>
  )
}
