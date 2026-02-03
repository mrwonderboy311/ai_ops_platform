import { Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { PrivateRoute } from './components/PrivateRoute'
import HostListPage from './pages/HostListPage'
import HostDetailPage from './pages/HostDetailPage'
import SSHTerminalPage from './pages/SSHTerminalPage'
import FileManagementPage from './pages/FileManagementPage'
import { BatchTaskListPage } from './pages/BatchTaskListPage'
import BatchTaskDetailPage from './pages/BatchTaskDetailPage'
import { ClusterListPage } from './pages/ClusterListPage'
import ClusterDetailPage from './pages/ClusterDetailPage'
import { ClusterMonitoringPage } from './pages/ClusterMonitoringPage'
import WorkloadListPage from './pages/WorkloadListPage'
import PodDetailPage from './pages/PodDetailPage'
import { HelmRepositoryPage } from './pages/HelmRepositoryPage'
import { HelmApplicationPage } from './pages/HelmApplicationPage'
import AlertListPage from './pages/AlertListPage'
import AuditLogPage from './pages/AuditLogPage'
import PerformanceDashboardPage from './pages/PerformanceDashboardPage'
import NotificationCenterPage from './pages/NotificationCenterPage'
import UserManagementPage from './pages/UserManagementPage'

// Dashboard placeholder component
function Dashboard() {
  return (
    <div style={{ padding: '24px' }}>
      <h1>MyOps AIOps Platform</h1>
      <p>欢迎使用智能运维平台</p>
    </div>
  )
}

// Layout component with logout
function Layout() {
  const logout = () => {
    localStorage.removeItem('myops-auth')
    window.location.href = '/login'
  }

  return (
    <div>
      <header
        style={{
          background: '#ffffff',
          padding: '16px 24px',
          borderBottom: '1px solid #f0f0f0',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <h2 style={{ margin: 0 }}>MyOps AIOps Platform</h2>
        <button onClick={logout}>退出登录</button>
      </header>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/hosts" element={<HostListPage />} />
        <Route path="/hosts/:id" element={<HostDetailPage />} />
        <Route path="/clusters" element={<ClusterListPage />} />
        <Route path="/clusters/:id" element={<ClusterDetailPage />} />
        <Route path="/clusters/:id/monitoring" element={<ClusterMonitoringPage />} />
        <Route path="/clusters/:id/workloads" element={<WorkloadListPage />} />
        <Route path="/clusters/:id/namespaces/:namespace/pods/:podName" element={<PodDetailPage />} />
        <Route path="/helm/repositories" element={<HelmRepositoryPage />} />
        <Route path="/helm/applications" element={<HelmApplicationPage />} />
        <Route path="/alerts" element={<AlertListPage />} />
        <Route path="/audit-logs" element={<AuditLogPage />} />
        <Route path="/performance" element={<PerformanceDashboardPage />} />
        <Route path="/notifications" element={<NotificationCenterPage />} />
        <Route path="/users" element={<UserManagementPage />} />
        <Route path="/batch-tasks" element={<BatchTaskListPage />} />
        <Route path="/batch-tasks/:id" element={<BatchTaskDetailPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </div>
  )
}

function App() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: 1,
        refetchOnWindowFocus: false,
      },
    },
  })

  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider
        theme={{
          token: {
            colorPrimary: '#667eea',
          },
        }}
      >
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route
            path="/hosts/ssh/:id"
            element={
              <PrivateRoute>
                <SSHTerminalPage />
              </PrivateRoute>
            }
          />
          <Route
            path="/hosts/files/:id"
            element={
              <PrivateRoute>
                <FileManagementPage />
              </PrivateRoute>
            }
          />
          <Route
            path="/*"
            element={
              <PrivateRoute>
                <Layout />
              </PrivateRoute>
            }
          />
        </Routes>
      </ConfigProvider>
    </QueryClientProvider>
  )
}

export default App
