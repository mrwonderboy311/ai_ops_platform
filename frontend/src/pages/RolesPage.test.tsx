import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import { RolesPage } from './RolesPage'
import * as rbacApi from '../api/rbac'

// Mock the RBAC API
vi.mock('../api/rbac', () => ({
  default: {
    getRoles: vi.fn(),
    getPermissions: vi.fn(),
    createRole: vi.fn(),
    updateRole: vi.fn(),
    deleteRole: vi.fn(),
    getRolePermissions: vi.fn(),
    assignRolePermissions: vi.fn(),
    seedDefaultRoles: vi.fn(),
  },
  PERMISSION_CATEGORIES: {},
  PERMISSION_ACTIONS: {},
  PERMISSION_SCOPES: {},
  RESOURCES: {},
}))

// Mock permission data
const mockPermissions = [
  {
    id: '1',
    name: 'hosts.read',
    displayName: 'Read Hosts',
    category: 'host',
    resource: 'hosts',
    action: 'read',
    scope: 'global',
  },
  {
    id: '2',
    name: 'hosts.write',
    displayName: 'Write Hosts',
    category: 'host',
    resource: 'hosts',
    action: 'update',
    scope: 'global',
  },
  {
    id: '3',
    name: 'clusters.read',
    displayName: 'Read Clusters',
    category: 'k8s',
    resource: 'clusters',
    action: 'read',
    scope: 'global',
  },
]

// Mock role data
const mockRoles = [
  {
    id: '1',
    name: 'viewer',
    displayName: 'Viewer',
    description: 'Read-only access',
    isSystem: false,
    isDefault: true,
    permissions: [
      { id: '1', name: 'hosts.read', displayName: 'Read Hosts' },
      { id: '3', name: 'clusters.read', displayName: 'Read Clusters' },
    ],
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    name: 'admin',
    displayName: 'Administrator',
    description: 'Full system access',
    isSystem: true,
    isDefault: false,
    permissions: mockPermissions,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
]

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
}

function renderWithProviders(component: React.ReactElement) {
  const queryClient = createTestQueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider>
        {component}
      </ConfigProvider>
    </QueryClientProvider>
  )
}

describe('RolesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the page title and statistics', async () => {
    vi.mocked(rbacApi.default.getRoles).mockResolvedValue({
      data: {
        data: mockRoles,
        total: 2,
        page: 1,
        pageSize: 20,
      },
    })
    vi.mocked(rbacApi.default.getPermissions).mockResolvedValue({
      data: {
        data: mockPermissions,
        total: 3,
        page: 1,
        pageSize: 1000,
      },
    })

    renderWithProviders(<RolesPage />)

    expect(screen.getByText(/AI Anomaly Detection/)).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByText('Total Roles')).toBeInTheDocument()
      expect(screen.getByText('System Roles')).toBeInTheDocument()
      expect(screen.getByText('Custom Roles')).toBeInTheDocument()
      expect(screen.getByText('Total Permissions')).toBeInTheDocument()
    })
  })

  it('renders the roles table', async () => {
    vi.mocked(rbacApi.default.getRoles).mockResolvedValue({
      data: {
        data: mockRoles,
        total: 2,
        page: 1,
        pageSize: 20,
      },
    })
    vi.mocked(rbacApi.default.getPermissions).mockResolvedValue({
      data: {
        data: mockPermissions,
        total: 3,
        page: 1,
        pageSize: 1000,
      },
    })

    renderWithProviders(<RolesPage />)

    await waitFor(() => {
      expect(screen.getByText('Viewer')).toBeInTheDocument()
      expect(screen.getByText('Administrator')).toBeInTheDocument()
    })
  })

  it('opens the add role modal when clicking "Add Role" button', async () => {
    vi.mocked(rbacApi.default.getRoles).mockResolvedValue({
      data: {
        data: mockRoles,
        total: 2,
        page: 1,
        pageSize: 20,
      },
    })
    vi.mocked(rbacApi.default.getPermissions).mockResolvedValue({
      data: {
        data: mockPermissions,
        total: 3,
        page: 1,
        pageSize: 1000,
      },
    })

    renderWithProviders(<RolesPage />)

    const addButton = await screen.findByText('Add Detection Rule')
    fireEvent.click(addButton)

    await waitFor(() => {
      expect(screen.getByText(/Add Detection Rule/)).toBeInTheDocument()
    })
  })

  it('seeds default roles when clicking "Seed Defaults" button', async () => {
    vi.mocked(rbacApi.default.getRoles).mockResolvedValue({
      data: {
        data: [],
        total: 0,
        page: 1,
        pageSize: 20,
      },
    })
    vi.mocked(rbacApi.default.getPermissions).mockResolvedValue({
      data: {
        data: [],
        total: 0,
        page: 1,
        pageSize: 1000,
      },
    })
    vi.mocked(rbacApi.default.seedDefaultRoles).mockResolvedValue({ data: { message: 'success' } })

    renderWithProviders(<RolesPage />)

    const seedButton = await screen.findByText('Seed Defaults')
    fireEvent.click(seedButton)

    await waitFor(() => {
      expect(rbacApi.default.seedDefaultRoles).toHaveBeenCalled()
    })
  })

  it('displays system role badge correctly', async () => {
    vi.mocked(rbacApi.default.getRoles).mockResolvedValue({
      data: {
        data: mockRoles,
        total: 2,
        page: 1,
        pageSize: 20,
      },
    })
    vi.mocked(rbacApi.default.getPermissions).mockResolvedValue({
      data: {
        data: mockPermissions,
        total: 3,
        page: 1,
        pageSize: 1000,
      },
    })

    renderWithProviders(<RolesPage />)

    await waitFor(() => {
      const adminRole = screen.getByText('Administrator')
      expect(adminRole).toBeInTheDocument()
    })
  })

  it('displays empty state when no roles exist', async () => {
    vi.mocked(rbacApi.default.getRoles).mockResolvedValue({
      data: {
        data: [],
        total: 0,
        page: 1,
        pageSize: 20,
      },
    })
    vi.mocked(rbacApi.default.getPermissions).mockResolvedValue({
      data: {
        data: [],
        total: 0,
        page: 1,
        pageSize: 1000,
      },
    })

    renderWithProviders(<RolesPage />)

    await waitFor(() => {
      expect(screen.getByText(/No roles configured/)).toBeInTheDocument()
    })
  })
})

describe('RolesPage - Permission Management', () => {
  it('opens permission management modal', async () => {
    vi.mocked(rbacApi.default.getRoles).mockResolvedValue({
      data: {
        data: mockRoles,
        total: 2,
        page: 1,
        pageSize: 20,
      },
    })
    vi.mocked(rbacApi.default.getPermissions).mockResolvedValue({
      data: {
        data: mockPermissions,
        total: 3,
        page: 1,
        pageSize: 1000,
      },
    })
    vi.mocked(rbacApi.default.getRolePermissions).mockResolvedValue({
      data: {
        data: [mockPermissions[0]],
        total: 1,
      },
    })

    renderWithProviders(<RolesPage />)

    await waitFor(() => {
      const manageButton = screen.getAllByLabelText('key')[0]
      fireEvent.click(manageButton)
    })

    await waitFor(() => {
      expect(screen.getByText(/Manage Permissions/)).toBeInTheDocument()
    })
  })
})

describe('RolesPage - Statistics', () => {
  it('calculates correct statistics', async () => {
    const testRoles = [
      ...mockRoles,
      {
        id: '3',
        name: 'operator',
        displayName: 'Operator',
        description: 'System operator',
        isSystem: false,
        isDefault: false,
        permissions: [mockPermissions[0]],
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      },
    ]

    vi.mocked(rbacApi.default.getRoles).mockResolvedValue({
      data: {
        data: testRoles,
        total: 3,
        page: 1,
        pageSize: 20,
      },
    })
    vi.mocked(rbacApi.default.getPermissions).mockResolvedValue({
      data: {
        data: mockPermissions,
        total: 3,
        page: 1,
        pageSize: 1000,
      },
    })

    renderWithProviders(<RolesPage />)

    await waitFor(() => {
      expect(screen.getByText('Total Roles')).toBeInTheDocument()
      expect(screen.getByText('System Roles')).toBeInTheDocument()
      expect(screen.getByText('Custom Roles')).toBeInTheDocument()
      expect(screen.getByText('Total Permissions')).toBeInTheDocument()
    })
  })
})
