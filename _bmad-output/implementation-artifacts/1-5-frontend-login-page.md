# Story 1.5: 前端登录页面

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

作为一名用户，
我想要通过网页登录平台，
以便访问管理界面。

## Acceptance Criteria

**Given** 前端项目已初始化
**When** 访问 `/login` 路由
**Then** 显示登录表单，包含：
- 用户名输入框（必填）
- 密码输入框（必填，type="password"）
- "登录" 按钮
- "忘记密码" 链接（暂不可用，显示"即将推出"）

**And** 表单验证：
- 提交时检查用户名和密码非空
- 空值显示输入框下方红色提示

**And** 登录成功后：
- Token 存储在 localStorage（key: `access_token`）
- 用户信息存储在 Zustand store
- 跳转到 `/dashboard` 首页

**And** 登录失败后：
- 显示 Ant Design message.error 提示错误信息
- 清空密码输入框

**And** Axios 请求拦截器配置：
- 所有请求自动添加 `Authorization: Bearer <token>` 头
- 401 响应自动跳转到登录页

**And** 路由守卫：
- 未登录访问受保护路由自动跳转到 `/login`
- 已登录访问 `/login` 自动跳转到 `/dashboard`

## Tasks / Subtasks

- [ ] 创建登录页面组件 (AC: #)
  - [ ] 创建 `frontend/src/pages/LoginPage.tsx`
  - [ ] 使用 Ant Design Form 组件
  - [ ] 添加用户名输入框（Input）
  - [ ] 添加密码输入框（Input.Password）
  - [ ] 添加登录按钮（Button type="primary"）
  - [ ] 添加"忘记密码"链接（Typography.Link）
- [ ] 实现表单验证 (AC: #)
  - [ ] 配置 Ant Design Form 规则
  - [ ] 用户名必填验证
  - [ ] 密码必填验证
  - [ ] 显示验证错误消息
- [ ] 创建 Zustand Auth Store (AC: #)
  - [ ] 创建 `frontend/src/store/authStore.ts`
  - [ ] 定义 State 接口（user, token, isAuthenticated）
  - [ ] 实现 login action
  - [ ] 实现 logout action
  - [ ] 实现 setUser action
  - [ ] 持久化 token 到 localStorage
- [ ] 创建 API 客户端 (AC: #)
  - [ ] 创建 `frontend/src/api/auth.ts`
  - [ ] 实现 loginAPI 函数
  - [ ] 实现 refreshAPI 函数
  - [ ] 添加 TypeScript 类型定义
- [ ] 配置 Axios 拦截器 (AC: #)
  - [ ] 创建 `frontend/src/api/client.ts`
  - [ ] 配置请求拦截器（添加 Bearer token）
  - [ ] 配置响应拦截器（处理 401）
  - [ ] 实现 token 刷新逻辑
  - [ ] 401 自动跳转登录页
- [ ] 创建路由守卫组件 (AC: #)
  - [ ] 创建 `frontend/src/components/PrivateRoute.tsx`
  - [ ] 检查认证状态
  - [ ] 未登录重定向到 /login
  - [ ] 已登录重定向到 /dashboard
  - [ ] 传递原始组件 props
- [ ] 更新路由配置 (AC: #)
  - [ ] 配置 `/login` 公开路由
  - [ ] 配置 `/dashboard` 受保护路由
  - [ ] 应用 PrivateRoute 组件
  - [ ] 添加默认重定向
- [ ] 创建 Dashboard 占位页面 (AC: #)
  - [ ] 创建 `frontend/src/pages/DashboardPage.tsx`
  - [ ] 显示欢迎消息
  - [ ] 显示用户信息
  - [ ] 添加登出按钮
- [ ] 更新 App 组件 (AC: #)
  - [ ] 集成 authStore
  - [ ] 配置 Routes
  - [ ] 添加全局 message 配置
- [ ] 添加 TypeScript 类型 (AC: #)
  - [ ] 创建 `frontend/src/types/auth.ts`
  - [ ] 定义 LoginRequest 接口
  - [ ] 定义 LoginResponse 接口
  - [ ] 定义 User 接口
  - [ ] 定义 ApiError 接口
- [ ] 编写组件测试 (AC: #)
  - [ ] 测试登录表单渲染
  - [ ] 测试表单验证
  - [ ] 测试登录成功跳转
  - [ ] 测试登录失败处理

## Dev Notes

### 项目背景

这是 Epic 1 的第五个 Story，负责实现前端登录页面。这是用户访问平台的第一界面，需要提供良好的用户体验和安全性。

### 组件设计

#### 页面结构

```
LoginPage
├── Card (登录卡片)
│   ├── Form (登录表单)
│   │   ├── Form.Item (用户名)
│   │   ├── Form.Item (密码)
│   │   └── Button (登录按钮)
│   └── Typography.Text ("忘记密码")
└── Message (全局消息提示)
```

### 登录页面组件

```tsx
// frontend/src/pages/LoginPage.tsx
import { useState } from 'react'
import { Form, Input, Button, Card, Typography, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/store/authStore'
import { loginAPI } from '@/api/auth'
import type { LoginRequest } from '@/types/auth'

const { Link } = Typography

export default function LoginPage() {
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const login = useAuthStore((state) => state.login)

  const onFinish = async (values: LoginRequest) => {
    setLoading(true)
    try {
      const response = await loginAPI(values)
      login(response.data.accessToken, response.data.user)
      message.success('登录成功')
      navigate('/dashboard')
    } catch (error: any) {
      const errorMessage = error.response?.data?.error?.message || '登录失败'
      message.error(errorMessage)
      form.setFieldsValue({ password: '' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={styles.container}>
      <Card style={styles.card} title="MyOps AIOps Platform">
        <Form
          form={form}
          name="login"
          onFinish={onFinish}
          autoComplete="off"
          size="large"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="用户名"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="密码"
            />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              登录
            </Button>
          </Form.Item>

          <Form.Item style={{ marginBottom: 0 }}>
            <Typography.Text style={{ fontSize: '12px' }}>
              忘记密码？<Link onClick={() => message.info('功能即将推出')}>找回密码</Link>
            </Typography.Text>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

const styles = {
  container: {
    minHeight: '100vh',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
  },
  card: {
    width: 400,
    boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
  },
}
```

### Zustand Auth Store

```typescript
// frontend/src/store/authStore.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types/auth'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  login: (token: string, user: User) => void
  logout: () => void
  setUser: (user: User) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,

      login: (token, user) =>
        set({
          token,
          user,
          isAuthenticated: true,
        }),

      logout: () =>
        set({
          token: null,
          user: null,
          isAuthenticated: false,
        }),

      setUser: (user) => set({ user }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ token: state.token, user: state.user }),
    }
  )
)
```

### API 客户端

```typescript
// frontend/src/api/auth.ts
import axios from 'axios'
import type { LoginRequest, LoginResponse, User } from '@/types/auth'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export async function loginAPI(data: LoginRequest): Promise<LoginResponse> {
  const response = await axios.post<LoginResponse>(
    `${API_BASE_URL}/api/v1/auth/login`,
    data
  )
  return response.data
}

export async function refreshAPI(refreshToken: string): Promise<LoginResponse> {
  const response = await axios.post<LoginResponse>(
    `${API_BASE_URL}/api/v1/auth/refresh`,
    { refreshToken }
  )
  return response.data
}
```

### Axios 拦截器

```typescript
// frontend/src/api/client.ts
import axios from 'axios'
import { message } from 'antd'
import { useAuthStore } from '@/store/authStore'
import { refreshAPI } from './auth'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    // 401 错误处理
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      try {
        // 尝试刷新 token
        const authStore = useAuthStore.getState()
        const refreshToken = localStorage.getItem('refresh_token')

        if (refreshToken) {
          const response = await refreshAPI(refreshToken)
          authStore.login(response.data.accessToken, response.data.user)

          // 重试原始请求
          originalRequest.headers.Authorization = `Bearer ${response.data.accessToken}`
          return apiClient(originalRequest)
        }
      } catch (refreshError) {
        // 刷新失败，退出登录
        useAuthStore.getState().logout()
        localStorage.removeItem('refresh_token')
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    // 其他错误
    const errorMessage = error.response?.data?.error?.message || '请求失败'
    message.error(errorMessage)
    return Promise.reject(error)
  }
)
```

### 路由守卫组件

```tsx
// frontend/src/components/PrivateRoute.tsx
import { Navigate } from 'react-router-dom'
import { useAuthStore } from '@/store/authStore'

interface PrivateRouteProps {
  children: React.ReactNode
}

export default function PrivateRoute({ children }: PrivateRouteProps) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}
```

### Dashboard 页面

```tsx
// frontend/src/pages/DashboardPage.tsx
import { Button, Card, Typography, Space } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/store/authStore'

const { Title, Text } = Typography

export default function DashboardPage() {
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()

  const handleLogout = () => {
    logout()
    localStorage.removeItem('refresh_token')
    navigate('/login')
  }

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Title level={2}>欢迎来到 MyOps AIOps Platform</Title>
          <Text>当前用户: {user?.username}</Text>
          <Text>邮箱: {user?.email}</Text>
          <Button type="primary" danger onClick={handleLogout}>
            退出登录
          </Button>
        </Space>
      </Card>
    </div>
  )
}
```

### 路由配置

```tsx
// frontend/src/App.tsx (更新)
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import LoginPage from '@/pages/LoginPage'
import DashboardPage from '@/pages/DashboardPage'
import PrivateRoute from '@/components/PrivateRoute'

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/dashboard"
            element={
              <PrivateRoute>
                <DashboardPage />
              </PrivateRoute>
            }
          />
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  )
}

export default App
```

### TypeScript 类型

```typescript
// frontend/src/types/auth.ts
export interface User {
  id: string
  username: string
  email: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  data: {
    accessToken: string
    refreshToken: string
    expiresIn: number
    user: User
  }
  requestId: string
}

export interface ApiError {
  error: {
    code: string
    message: string
  }
  requestId: string
}
```

### 样式增强

```css
/* frontend/src/pages/LoginPage.css */
.login-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  border-radius: 8px;
}

.login-form .ant-form-item {
  margin-bottom: 24px;
}

.login-button {
  width: 100%;
  height: 40px;
  font-size: 16px;
}
```

### 环境变量

```env
# frontend/.env.example
VITE_API_BASE_URL=http://localhost:8080
```

### 安全注意事项

1. **Token 存储**：Access Token 存储在 localStorage（可考虑 httpOnly cookie）
2. **Refresh Token**：单独存储，用于刷新 Access Token
3. **HTTPS**：生产环境必须使用 HTTPS
4. **XSS 防护**：避免在 localStorage 存储敏感信息
5. **CSRF 防护**：后端应实现 CSRF token

### 用户体验优化

1. **加载状态**：登录按钮显示 loading
2. **错误提示**：使用 Ant Design message 组件
3. **表单重置**：登录失败后清空密码
4. **自动聚焦**：页面加载后聚焦用户名输入框
5. **回车提交**：支持 Enter 键提交表单

### Dev Agent Guardrails

1. **必须使用 Ant Design Form 组件**
2. **必须使用 Zustand 管理认证状态**
3. **必须实现 Axios 请求拦截器**
4. **必须实现 401 自动跳转登录**
5. **必须实现路由守卫**
6. **必须使用 TypeScript 类型**
7. **登录成功必须跳转到 /dashboard**
8. **必须实现 token 刷新逻辑**

### Dependencies

**前端依赖**（已在 Story 1.1 安装）：
```
- antd@^5.22.5
- zustand@^5.0.2
- @tanstack/react-query@^5.62.7
- axios@^1.7.9
- react-router-dom@^6.28.0
- zustand/middleware (persist)
```

### References

- [Source: docs/plans/2025-02-02-aiops-platform-design.md#5-前端技术栈]
- [Source: _bmad-output/planning-artifacts/epics.md#Epic-1]
- [Source: _bmad-output/implementation-artifacts/1-4-user-login-and-token-generation.md]

## Dev Agent Record

### Agent Model Used

glm-4.7 (claude-opus-4-5-20251101)

### Debug Log References

None (story preparation phase)

### Completion Notes List

**Story Created (2026-02-03)**:

Story prepared for development with detailed implementation guidance.
