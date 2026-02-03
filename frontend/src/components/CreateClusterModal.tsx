import { useState } from 'react'
import { Modal, Form, Input, Select, message, Steps, Button, Alert, Space } from 'antd'
import type { CreateClusterRequest } from '../types/cluster'
import { clusterApi } from '../api/cluster'

const { TextArea } = Input
const { Option } = Select

interface CreateClusterModalProps {
  visible: boolean
  onCancel: () => void
  onSuccess: () => void
}

export const CreateClusterModal: React.FC<CreateClusterModalProps> = ({
  visible,
  onCancel,
  onSuccess,
}) => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [currentStep, setCurrentStep] = useState(0)
  const [kubeconfig, setKubeconfig] = useState('')
  const [testResult, setTestResult] = useState<any>(null)
  const [testing, setTesting] = useState(false)

  const handleTestConnection = async () => {
    try {
      setTesting(true)
      setTestResult(null)

      const config = form.getFieldValue('kubeconfig')
      if (!config) {
        message.error('Please provide kubeconfig')
        return
      }

      const result = await clusterApi.testConnection({
        kubeconfig: config,
        endpoint: form.getFieldValue('endpoint'),
      })

      setTestResult(result)
      if (result.success) {
        message.success('Connection successful!')
      } else {
        message.error(`Connection failed: ${result.error}`)
      }
    } catch (error: any) {
      setTestResult({ success: false, error: error.response?.data?.message || error.message })
      message.error(`Connection test failed: ${error.response?.data?.message || error.message}`)
    } finally {
      setTesting(false)
    }
  }

  const handleOk = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      const request: CreateClusterRequest = {
        name: values.name,
        description: values.description,
        type: values.type,
        endpoint: values.endpoint,
        kubeconfig: kubeconfig,
        region: values.region,
        provider: values.provider,
      }

      await clusterApi.createCluster(request)
      message.success('Cluster added successfully')
      form.resetFields()
      setCurrentStep(0)
      setKubeconfig('')
      setTestResult(null)
      onSuccess()
    } catch (error: any) {
      message.error(`Failed to create cluster: ${error.response?.data?.message || error.message}`)
    } finally {
      setLoading(false)
    }
  }

  const handleCancel = () => {
    form.resetFields()
    setCurrentStep(0)
    setKubeconfig('')
    setTestResult(null)
    onCancel()
  }

  const steps = [
    {
      title: 'Basic Info',
      content: (
        <>
          <Form.Item
            label="Cluster Name"
            name="name"
            rules={[{ required: true, message: 'Please enter cluster name' }]}
          >
            <Input placeholder="e.g., production-cluster" />
          </Form.Item>

          <Form.Item label="Description" name="description">
            <TextArea rows={2} placeholder="Optional description" />
          </Form.Item>

          <Form.Item
            label="Cluster Type"
            name="type"
            rules={[{ required: true, message: 'Please select cluster type' }]}
          >
            <Select>
              <Option value="managed">Managed (EKS, GKE, AKS, etc.)</Option>
              <Option value="self-hosted">Self-Hosted (kubeadm, k3s, etc.)</Option>
            </Select>
          </Form.Item>

          <Form.Item label="Provider" name="provider">
            <Select placeholder="Select cloud provider (optional)">
              <Option value="aws">AWS (EKS)</Option>
              <Option value="gcp">GCP (GKE)</Option>
              <Option value="azure">Azure (AKS)</Option>
              <Option value="alibaba">Alibaba Cloud (ACK)</Option>
              <Option value="tencent">Tencent Cloud (TKE)</Option>
            </Select>
          </Form.Item>

          <Form.Item label="Region" name="region">
            <Input placeholder="e.g., us-west-2, eu-central-1" />
          </Form.Item>
        </>
      ),
    },
    {
      title: 'Connection',
      content: (
        <>
          <Alert
            style={{ marginBottom: '16px' }}
            message="Paste your kubeconfig file content"
            description="The kubeconfig file contains the connection details for your Kubernetes cluster. You can usually find it at ~/.kube/config"
            type="info"
            showIcon
          />

          <Form.Item
            label="Kubeconfig"
            name="kubeconfig"
            rules={[{ required: true, message: 'Please paste kubeconfig content' }]}
          >
            <TextArea
              rows={10}
              placeholder="Paste kubeconfig YAML content here..."
              onChange={(e) => setKubeconfig(e.target.value)}
              style={{ fontFamily: 'monospace', fontSize: '12px' }}
            />
          </Form.Item>

          <Form.Item label="API Endpoint (Optional)" name="endpoint">
            <Input placeholder="https://cluster.example.com (override default from kubeconfig)" />
          </Form.Item>

          <Space style={{ marginTop: '8px' }}>
            <Button onClick={handleTestConnection} loading={testing}>
              Test Connection
            </Button>
          </Space>

          {testResult && (
            <Alert
              style={{ marginTop: '16px' }}
              message={testResult.success ? 'Connection Successful' : 'Connection Failed'}
              description={
                testResult.success
                  ? `Kubernetes version: ${testResult.version}, Nodes: ${testResult.nodeCount}`
                  : testResult.error
              }
              type={testResult.success ? 'success' : 'error'}
              showIcon
            />
          )}
        </>
      ),
    },
  ]

  return (
    <Modal
      title="Add Kubernetes Cluster"
      open={visible}
      onOk={handleOk}
      onCancel={handleCancel}
      okText="Add Cluster"
      cancelText="Cancel"
      width={700}
      confirmLoading={loading}
      okButtonProps={{ disabled: currentStep === 0 }}
    >
      <Form form={form} layout="vertical">
        <Steps current={currentStep} style={{ marginBottom: '24px' }}>
          <Steps.Step title="Basic Info" />
          <Steps.Step title="Connection" />
        </Steps>

        {steps[currentStep].content}

        <div style={{ marginTop: '24px', textAlign: 'center' }}>
          {currentStep > 0 && (
            <Button style={{ marginRight: '8px' }} onClick={() => setCurrentStep(currentStep - 1)}>
              Previous
            </Button>
          )}
          {currentStep < steps.length - 1 && (
            <Button type="primary" onClick={() => setCurrentStep(currentStep + 1)}>
              Next
            </Button>
          )}
        </div>
      </Form>
    </Modal>
  )
}
