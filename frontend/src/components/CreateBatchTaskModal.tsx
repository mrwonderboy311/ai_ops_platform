import { useState } from 'react'
import { Modal, Form, Input, Select, message } from 'antd'
import { useQuery } from '@tanstack/react-query'
import type { CreateBatchTaskRequest } from '../types/batchTask'
import { batchTaskApi } from '../api/batchTask'
import { hostApi } from '../api/host'

const { TextArea } = Input
const { Option } = Select

interface CreateBatchTaskModalProps {
  visible: boolean
  onCancel: () => void
  onSuccess: () => void
}

export const CreateBatchTaskModal: React.FC<CreateBatchTaskModalProps> = ({
  visible,
  onCancel,
  onSuccess,
}) => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)

  // Fetch available hosts
  const { data: hosts } = useQuery({
    queryKey: ['hosts'],
    queryFn: async () => {
      const response = await hostApi.listHosts({ page: 1, pageSize: 1000 })
      return response.hosts.filter((h) => h.status === 'approved' || h.status === 'online')
    },
    enabled: visible,
  })

  const handleOk = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      const request: CreateBatchTaskRequest = {
        name: values.name,
        description: values.description,
        type: values.type,
        strategy: values.strategy,
        command: values.command,
        script: values.script,
        timeout: values.timeout || 60,
        maxRetries: values.maxRetries || 0,
        parallelism: values.parallelism || 0,
        hostIds: values.hostIds,
      }

      await batchTaskApi.createBatchTask(request)

      // Also execute the task
      const task = await batchTaskApi.createBatchTask(request)
      await batchTaskApi.executeBatchTask({ taskId: task.id })

      message.success('Batch task created and execution started')
      form.resetFields()
      onSuccess()
    } catch (error: any) {
      message.error(`Failed to create batch task: ${error.response?.data?.message || error.message}`)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title="Create Batch Task"
      open={visible}
      onOk={handleOk}
      onCancel={onCancel}
      okText="Create & Execute"
      cancelText="Cancel"
      width={600}
      confirmLoading={loading}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          type: 'command',
          strategy: 'parallel',
          timeout: 60,
          maxRetries: 0,
          parallelism: 0,
        }}
      >
        <Form.Item
          label="Task Name"
          name="name"
          rules={[{ required: true, message: 'Please enter task name' }]}
        >
          <Input placeholder="Enter task name" />
        </Form.Item>

        <Form.Item label="Description" name="description">
          <TextArea rows={2} placeholder="Enter task description (optional)" />
        </Form.Item>

        <Form.Item
          label="Task Type"
          name="type"
          rules={[{ required: true, message: 'Please select task type' }]}
        >
          <Select>
            <Option value="command">Command</Option>
            <Option value="script">Script</Option>
            <Option value="file_op">File Operation</Option>
          </Select>
        </Form.Item>

        <Form.Item
          label="Execution Strategy"
          name="strategy"
          rules={[{ required: true, message: 'Please select execution strategy' }]}
          extra="Parallel: Execute on all hosts at once | Serial: Execute one at a time | Rolling: Execute in batches"
        >
          <Select>
            <Option value="parallel">Parallel (All hosts at once)</Option>
            <Option value="serial">Serial (One at a time)</Option>
            <Option value="rolling">Rolling (In batches)</Option>
          </Select>
        </Form.Item>

        <Form.Item noStyle shouldUpdate={(prev, curr) => prev.type !== curr.type}>
          {({ getFieldValue }) => {
            const taskType = getFieldValue('type')
            return (
              <>
                {taskType === 'command' && (
                  <Form.Item
                    label="Command"
                    name="command"
                    rules={[{ required: true, message: 'Please enter command' }]}
                  >
                    <TextArea
                      rows={3}
                      placeholder="Enter command to execute (e.g., 'ls -la /tmp')"
                      style={{ fontFamily: 'monospace' }}
                    />
                  </Form.Item>
                )}
                {taskType === 'script' && (
                  <Form.Item
                    label="Script"
                    name="script"
                    rules={[{ required: true, message: 'Please enter script' }]}
                  >
                    <TextArea
                      rows={6}
                      placeholder="#!/bin/bash&#10;# Enter your script here"
                      style={{ fontFamily: 'monospace' }}
                    />
                  </Form.Item>
                )}
              </>
            )
          }}
        </Form.Item>

        <Form.Item noStyle shouldUpdate={(prev, curr) => prev.strategy !== curr.strategy}>
          {({ getFieldValue }) => {
            const strategy = getFieldValue('strategy')
            return strategy === 'rolling' ? (
              <Form.Item
                label="Parallelism"
                name="parallelism"
                rules={[{ required: true, message: 'Please specify parallelism' }]}
                extra="Number of hosts to execute simultaneously"
              >
                <Input type="number" min={1} placeholder="e.g., 5" />
              </Form.Item>
            ) : null
          }}
        </Form.Item>

        <Form.Item
          label="Target Hosts"
          name="hostIds"
          rules={[{ required: true, message: 'Please select at least one host' }]}
        >
          <Select
            mode="multiple"
            placeholder="Select hosts to execute on"
            showSearch
            filterOption={(input, option) =>
              String(option?.label ?? option?.children ?? '').toLowerCase().includes(input.toLowerCase())
            }
          >
            {hosts?.map((host) => (
              <Option key={host.id} value={host.id}>
                {host.hostname || host.ipAddress} ({host.ipAddress})
              </Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item label="Timeout (seconds)" name="timeout">
          <Input type="number" min={1} placeholder="60" />
        </Form.Item>

        <Form.Item label="Max Retries" name="maxRetries" extra="Number of retries on failure">
          <Input type="number" min={0} placeholder="0" />
        </Form.Item>
      </Form>
    </Modal>
  )
}
