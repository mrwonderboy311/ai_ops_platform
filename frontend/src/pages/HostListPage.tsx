import React, { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Table,
  Button,
  Input,
  Tag,
  Space,
  Modal,
  message,
  Popconfirm,
  Tooltip,
} from 'antd'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import type { FilterValue, SorterResult } from 'antd/es/table/interface'
import { SearchOutlined, PlusOutlined, ReloadOutlined, CheckOutlined, CloseOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { hostApi } from '../api/host'
import type { Host, HostStatus, HostListParams } from '../types/host'

const { Search } = Input

const HostListPage: React.FC = () => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  // State for filters and pagination
  const [params, setParams] = useState<HostListParams>({
    page: 1,
    pageSize: 20,
    status: undefined,
    hostname: '',
    ipAddress: '',
  })

  // Search input states
  const [hostnameSearch, setHostnameSearch] = useState('')
  const [ipAddressSearch, setIpAddressSearch] = useState('')

  // Fetch hosts
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['hosts', params],
    queryFn: () => hostApi.listHosts(params),
    refetchInterval: 30000, // Auto-refresh every 30 seconds
  })

  // Approve host mutation
  const approveMutation = useMutation({
    mutationFn: (id: string) => hostApi.approveHost(id),
    onSuccess: () => {
      message.success('Host approved successfully')
      queryClient.invalidateQueries({ queryKey: ['hosts'] })
    },
    onError: (error: Error) => {
      message.error(`Failed to approve host: ${error.message}`)
    },
  })

  // Reject host mutation
  const rejectMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason?: string }) =>
      hostApi.rejectHost(id, { reason }),
    onSuccess: () => {
      message.success('Host rejected successfully')
      queryClient.invalidateQueries({ queryKey: ['hosts'] })
      setRejectModalVisible(false)
      setRejectingHostId(null)
      setRejectReason('')
    },
    onError: (error: Error) => {
      message.error(`Failed to reject host: ${error.message}`)
    },
  })

  // Delete host mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) => hostApi.deleteHost(id),
    onSuccess: () => {
      message.success('Host deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['hosts'] })
    },
    onError: (error: Error) => {
      message.error(`Failed to delete host: ${error.message}`)
    },
  })

  // Reject modal state
  const [rejectModalVisible, setRejectModalVisible] = useState(false)
  const [rejectingHostId, setRejectingHostId] = useState<string | null>(null)
  const [rejectReason, setRejectReason] = useState('')

  // Status badge renderer
  const renderStatus = (status: HostStatus) => {
    const statusConfig = {
      pending: { color: 'orange', text: 'Pending' },
      approved: { color: 'blue', text: 'Approved' },
      rejected: { color: 'red', text: 'Rejected' },
      offline: { color: 'default', text: 'Offline' },
      online: { color: 'green', text: 'Online' },
    }
    const config = statusConfig[status] || statusConfig.pending
    return <Tag color={config.color}>{config.text}</Tag>
  }

  // Table columns
  const columns: ColumnsType<Host> = [
    {
      title: 'Hostname',
      dataIndex: 'hostname',
      key: 'hostname',
      sorter: true,
      render: (hostname: string, record) => (
        <a onClick={() => navigate(`/hosts/${record.id}`)}>{hostname || record.ipAddress}</a>
      ),
    },
    {
      title: 'IP Address',
      dataIndex: 'ipAddress',
      key: 'ipAddress',
      sorter: true,
      render: (ip: string, record) => `${ip}:${record.port}`,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      filters: [
        { text: 'Pending', value: 'pending' },
        { text: 'Approved', value: 'approved' },
        { text: 'Rejected', value: 'rejected' },
        { text: 'Offline', value: 'offline' },
        { text: 'Online', value: 'online' },
      ],
      render: renderStatus,
    },
    {
      title: 'OS Type',
      dataIndex: 'osType',
      key: 'osType',
      render: (osType: string) => osType || '-',
    },
    {
      title: 'CPU',
      dataIndex: 'cpuCores',
      key: 'cpuCores',
      render: (cores: number | null) => cores ?? '-',
    },
    {
      title: 'Memory (GB)',
      dataIndex: 'memoryGB',
      key: 'memoryGB',
      render: (mem: number | null) => mem ?? '-',
    },
    {
      title: 'Created At',
      dataIndex: 'createdAt',
      key: 'createdAt',
      sorter: true,
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: 'Actions',
      key: 'actions',
      fixed: 'right',
      width: 200,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="View Details">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => navigate(`/hosts/${record.id}`)}
            />
          </Tooltip>

          {record.status === 'pending' && (
            <>
              <Popconfirm
                title="Approve this host?"
                description="This host will be able to report data"
                onConfirm={() => approveMutation.mutate(record.id)}
                okText="Yes"
                cancelText="No"
              >
                <Tooltip title="Approve">
                  <Button type="text" icon={<CheckOutlined />} />
                </Tooltip>
              </Popconfirm>

              <Tooltip title="Reject">
                <Button
                  type="text"
                  danger
                  icon={<CloseOutlined />}
                  onClick={() => {
                    setRejectingHostId(record.id)
                    setRejectModalVisible(true)
                  }}
                />
              </Tooltip>
            </>
          )}

          <Popconfirm
            title="Delete this host?"
            description="This action cannot be undone"
            onConfirm={() => deleteMutation.mutate(record.id)}
            okText="Yes"
            cancelText="No"
          >
            <Tooltip title="Delete">
              <Button type="text" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  // Handle table change
  const handleTableChange = (
    newPagination: TablePaginationConfig,
    filters: Record<string, FilterValue | null>,
    sorter: SorterResult<Host> | SorterResult<Host>[]
  ) => {
    const newParams: HostListParams = {
      ...params,
      page: newPagination.current || 1,
      pageSize: newPagination.pageSize || 20,
    }

    // Handle filters - FilterValue can be string | number | boolean | (string | number | boolean)[]
    const statusFilter = filters.status
    if (statusFilter && typeof statusFilter === 'string') {
      newParams.status = statusFilter as HostStatus
    }

    // Handle sorter
    if (Array.isArray(sorter)) {
      if (sorter.length > 0 && sorter[0].field) {
        newParams.sortBy = String(sorter[0].field)
        newParams.sortDesc = sorter[0].order === 'descend'
      }
    } else if (sorter?.field) {
      newParams.sortBy = String(sorter.field)
      newParams.sortDesc = sorter.order === 'descend'
    }

    setParams(newParams)
  }

  // Handle search
  const handleSearch = () => {
    setParams({
      ...params,
      hostname: hostnameSearch,
      ipAddress: ipAddressSearch,
      page: 1,
    })
  }

  const handleReset = () => {
    setHostnameSearch('')
    setIpAddressSearch('')
    setParams({
      page: 1,
      pageSize: 20,
      status: undefined,
      hostname: '',
      ipAddress: '',
    })
  }

  const handleRejectConfirm = () => {
    if (rejectingHostId) {
      rejectMutation.mutate({
        id: rejectingHostId,
        reason: rejectReason,
      })
    }
  }

  if (error) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <p style={{ color: 'red' }}>Failed to load hosts: {(error as Error).message}</p>
        <Button onClick={() => refetch()}>Retry</Button>
      </div>
    )
  }

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1>Host Management</h1>
        <Space>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/hosts/new')}
          >
            Add Host
          </Button>
        </Space>
      </div>

      {/* Filters */}
      <div style={{
        background: '#fff',
        padding: '16px',
        borderRadius: '8px',
        marginBottom: '16px',
        display: 'flex',
        gap: '16px',
        alignItems: 'center',
      }}>
        <Search
          placeholder="Search by hostname"
          value={hostnameSearch}
          onChange={(e) => setHostnameSearch(e.target.value)}
          onSearch={handleSearch}
          style={{ width: 200 }}
          allowClear
        />
        <Search
          placeholder="Search by IP address"
          value={ipAddressSearch}
          onChange={(e) => setIpAddressSearch(e.target.value)}
          onSearch={handleSearch}
          style={{ width: 200 }}
          allowClear
        />
        <Button onClick={handleSearch} icon={<SearchOutlined />}>
          Search
        </Button>
        <Button onClick={handleReset}>Reset</Button>
        <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
          Refresh
        </Button>
      </div>

      {/* Table */}
      <Table
        columns={columns}
        dataSource={data?.hosts || []}
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: params.page,
          pageSize: params.pageSize,
          total: data?.total || 0,
          showSizeChanger: true,
          showTotal: (total) => `Total ${total} hosts`,
        }}
        onChange={handleTableChange}
        scroll={{ x: 1200 }}
      />

      {/* Reject Modal */}
      <Modal
        title="Reject Host"
        open={rejectModalVisible}
        onOk={handleRejectConfirm}
        onCancel={() => {
          setRejectModalVisible(false)
          setRejectingHostId(null)
          setRejectReason('')
        }}
        okText="Reject"
        okButtonProps={{ danger: true }}
      >
        <p>Are you sure you want to reject this host? The agent will no longer be able to report data.</p>
        <Input.TextArea
          rows={4}
          placeholder="Reason for rejection (optional)"
          value={rejectReason}
          onChange={(e) => setRejectReason(e.target.value)}
          maxLength={500}
          showCount
        />
      </Modal>
    </div>
  )
}

export default HostListPage
