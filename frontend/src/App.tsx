import { Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { PrivateRoute } from './components/PrivateRoute'
import HostListPage from './pages/HostListPage'

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
