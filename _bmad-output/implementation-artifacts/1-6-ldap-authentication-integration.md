# Story 1.6: LDAP 认证集成

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

作为一名企业用户，
我想要使用公司 LDAP/AD 账号登录，
以便无需单独注册账号。

## Acceptance Criteria

**Given** 系统已配置 LDAP 连接
**When** 发送 POST 请求到 `/api/v1/auth/ldap-login`
```json
{
  "username": "companyuser",
  "password": "LdapPass123!"
}
```
**Then** 系统连接到 LDAP 服务器验证用户

**And** LDAP 验证成功时：
- 如果用户不存在于 `users` 表，自动创建用户记录
- 返回标准的登录响应（包含 JWT Token）
- 用户类型标记为 "ldap"

**And** LDAP 验证失败时：
- 返回 401：`{"error": {"code": "LDAP_AUTH_FAILED", "message": "LDAP 认证失败"}}`

**And** 配置文件支持：
```yaml
ldap:
  url: "ldap://ldap.example.com:389"
  baseDN: "dc=example,dc=com"
  bindDN: "cn=admin,dc=example,dc=com"
  bindPassword: "secret"
  userFilter: "(uid=%s)"
  userAttributes:
    username: "uid"
    email: "mail"
    displayName: "cn"
```

**And** 前端登录页面添加 "LDAP 登录" 标签页
- 用户可以选择"本地账号"或"LDAP 账号"登录方式

## Tasks / Subtasks

- [ ] 创建 LDAP 配置结构 (AC: #)
  - [ ] 创建 `backend/pkg/ldap/config.go`
  - [ ] 定义 LDAPConfig 结构体
  - [ ] 定义连接参数
  - [ ] 定义用户属性映射
- [ ] 实现 LDAP 客户端 (AC: #)
  - [ ] 创建 `backend/pkg/ldap/client.go`
  - [ ] 实现 Connect 函数
  - [ ] 实现 Authenticate 函数
  - [ ] 实现 GetUserAttributes 函数
  - [ ] 实现 Close 函数
- [ ] 实现 LDAP 认证逻辑 (AC: #)
  - [ ] 创建 `backend/pkg/ldap/authenticator.go`
  - [ ] 实现用户查找逻辑
  - [ ] 实现密码验证逻辑
  - [ ] 实现属性提取逻辑
  - [ ] 处理连接错误
- [ ] 扩展用户模型 (AC: #)
  - [ ] 在 User 模型添加 UserType 字段
  - [ ] 定义用户类型枚举（local, ldap）
  - [ ] 更新数据库迁移
- [ ] 实现 LDAP 登录 Service (AC: #)
  - [ ] 在 auth_service 添加 LDAPLogin 方法
  - [ ] 连接 LDAP 服务器
  - [ ] 验证用户凭据
  - [ ] 获取用户属性
  - [ ] 自动创建或更新用户
  - [ ] 生成 JWT Token
- [ ] 添加 Protobuf 定义 (AC: #)
  - [ ] 添加 LdapLoginRequest 消息
  - [ ] 添加 LdapLogin RPC
  - [ ] 添加 HTTP 映射
- [ ] 实现 HTTP 处理器 (AC: #)
  - [ ] 创建 `backend/api-gateway/internal/handler/ldap_login.go`
  - [ ] 实现 ldap-login POST 处理器
- [ ] 集成到 API Gateway (AC: #)
  - [ ] 注册路由 `POST /api/v1/auth/ldap-login`
  - [ ] 应用日志和恢复中间件
- [ ] 前端登录页面改造 (AC: #)
  - [ ] 添加 Tabs 组件（本地/LDAP）
  - [ ] 创建 LDAP 登录表单
  - [ ] 实现 ldapLoginAPI 函数
  - [ ] 更新登录逻辑
- [ ] 编写单元测试 (AC: #)
  - [ ] 测试 LDAP 连接
  - [ ] 测试认证成功
  - [ ] 测试认证失败
  - [ ] 测试用户自动创建
- [ ] 编写集成测试 (AC: #)
  - [ ] 使用测试 LDAP 服务器
  - [ ] 测试完整登录流程

## Dev Notes

### 项目背景

这是 Epic 1 的第六个 Story，负责集成 LDAP/AD 认证。企业用户通常使用统一的目录服务管理员工账号，LDAP 集成可以让用户使用现有账号登录平台。

### LDAP 认证流程

```
1. 接收请求 → 2. 连接 LDAP → 3. 查找用户 → 4. 验证密码 → 5. 获取属性 → 6. 创建/更新用户 → 7. 生成 Token
```

### LDAP 配置

#### 配置结构

```go
// backend/pkg/ldap/config.go
package ldap

import "time"

type Config struct {
    // 连接配置
    URL          string        `yaml:"url" env:"LDAP_URL"`                     // ldap://host:port
    BindDN       string        `yaml:"bind_dn" env:"LDAP_BIND_DN"`             // 管理员 DN
    BindPassword string        `yaml:"bind_password" env:"LDAP_BIND_PASSWORD"` // 管理员密码
    BaseDN       string        `yaml:"base_dn" env:"LDAP_BASE_DN"`             // 搜索基础 DN
    Timeout      time.Duration `yaml:"timeout" env:"LDAP_TIMEOUT"`             // 连接超时

    // 用户搜索配置
    UserFilter string `yaml:"user_filter" env:"LDAP_USER_FILTER"` // (uid=%s)
    UserScope  string `yaml:"user_scope" env:"LDAP_USER_SCOPE"`   // sub, one, base

    // 用户属性映射
    Attributes AttributeMapping `yaml:"attributes"`
}

type AttributeMapping struct {
    Username    string `yaml:"username"`    // uid
    Email       string `yaml:"email"`       // mail
    DisplayName string `yaml:"display_name"` // cn
    FirstName   string `yaml:"first_name"`  // givenName
    LastName    string `yaml:"last_name"`   // sn
}

// 用户类型枚举
type UserType string

const (
    UserTypeLocal UserType = "local"
    UserTypeLDAP  UserType = "ldap"
)
```

### LDAP 客户端实现

```go
// backend/pkg/ldap/client.go
package ldap

import (
    "crypto/tls"
    "fmt"
    "time"

    "github.com/go-ldap/ldap/v3"
)

type Client struct {
    conn   *ldap.Conn
    config Config
}

// NewClient 创建 LDAP 客户端
func NewClient(config Config) (*Client, error) {
    c := &Client{config: config}

    if err := c.Connect(); err != nil {
        return nil, err
    }

    return c, nil
}

// Connect 连接 LDAP 服务器
func (c *Client) Connect() error {
    var err error
    var tlsConfig *tls.Config

    // 根据协议选择连接方式
    if c.config.URL[:5] == "ldaps" {
        // LDAPS (636)
        tlsConfig = &tls.Config{
            InsecureSkipVerify: true, // 生产环境应验证证书
        }
        c.conn, err = ldap.DialTLS("tcp", getHostPort(c.config.URL), tlsConfig)
    } else {
        // LDAP (389)
        c.conn, err = ldap.Dial("tcp", getHostPort(c.config.URL))
    }

    if err != nil {
        return fmt.Errorf("ldap connect failed: %w", err)
    }

    // 设置超时
    c.conn.SetTimeout(c.config.Timeout)

    // 绑定管理员账号（用于搜索用户）
    if c.config.BindDN != "" {
        err = c.conn.Bind(c.config.BindDN, c.config.BindPassword)
        if err != nil {
            c.conn.Close()
            return fmt.Errorf("ldap bind failed: %w", err)
        }
    }

    return nil
}

// Authenticate 验证用户凭据
func (c *Client) Authenticate(username, password string) (bool, error) {
    // 1. 查找用户 DN
    userDN, err := c.findUserDN(username)
    if err != nil {
        return false, err
    }

    // 2. 使用用户凭据重新绑定
    err = c.conn.Bind(userDN, password)
    if err != nil {
        if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
            return false, nil // 密码错误
        }
        return false, err
    }

    // 3. 重新绑定管理员账号
    if c.config.BindDN != "" {
        c.conn.Bind(c.config.BindDN, c.config.BindPassword)
    }

    return true, nil
}

// GetUserAttributes 获取用户属性
func (c *Client) GetUserAttributes(username string) (*UserAttributes, error) {
    userDN, err := c.findUserDN(username)
    if err != nil {
        return nil, err
    }

    // 构建属性列表
    attributes := []string{
        c.config.Attributes.Username,
        c.config.Attributes.Email,
        c.config.Attributes.DisplayName,
        c.config.Attributes.FirstName,
        c.config.Attributes.LastName,
    }

    // 搜索用户
    searchRequest := ldap.NewSearchRequest(
        userDN,
        ldap.ScopeBaseObject,
        ldap.NeverDerefAliases,
        0,
        0,
        false,
        "(objectClass=*)",
        attributes,
        nil,
    )

    sr, err := c.conn.Search(searchRequest)
    if err != nil {
        return nil, err
    }

    if len(sr.Entries) == 0 {
        return nil, fmt.Errorf("user not found")
    }

    entry := sr.Entries[0]
    return &UserAttributes{
        Username:    getAttributeValue(entry, c.config.Attributes.Username),
        Email:       getAttributeValue(entry, c.config.Attributes.Email),
        DisplayName: getAttributeValue(entry, c.config.Attributes.DisplayName),
        FirstName:   getAttributeValue(entry, c.config.Attributes.FirstName),
        LastName:    getAttributeValue(entry, c.config.Attributes.LastName),
    }, nil
}

// findUserDN 查找用户 DN
func (c *Client) findUserDN(username string) (string, error) {
    filter := fmt.Sprintf(c.config.UserFilter, username)

    searchRequest := ldap.NewSearchRequest(
        c.config.BaseDN,
        ldap.ScopeWholeSubtree,
        ldap.NeverDerefAliases,
        0,
        0,
        false,
        filter,
        []string{"dn"},
        nil,
    )

    sr, err := c.conn.Search(searchRequest)
    if err != nil {
        return "", err
    }

    if len(sr.Entries) == 0 {
        return "", fmt.Errorf("user not found")
    }

    return sr.Entries[0].DN, nil
}

// Close 关闭连接
func (c *Client) Close() error {
    if c.conn != nil {
        return c.conn.Close()
    }
    return nil
}

type UserAttributes struct {
    Username    string
    Email       string
    DisplayName string
    FirstName   string
    LastName    string
}

func getHostPort(url string) string {
    // 移除协议前缀
    if len(url) > 7 && url[:7] == "ldap://" {
        return url[7:]
    }
    if len(url) > 8 && url[:8] == "ldaps://" {
        return url[8:]
    }
    return url
}

func getAttributeValue(entry *ldap.Entry, attr string) string {
    if len(entry.GetAttributeValues(attr)) > 0 {
        return entry.GetAttributeValue(attr)
    }
    return ""
}
```

### 用户模型扩展

```go
// backend/pkg/model/user.go (扩展)
package model

type UserType string

const (
    UserTypeLocal UserType = "local"
    UserTypeLDAP  UserType = "ldap"
)

type User struct {
    ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    Username     string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"username"`
    Email        string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
    PasswordHash string     `gorm:"type:varchar(255)" json:"-"` // LDAP 用户可为空
    UserType     UserType   `gorm:"type:varchar(50);not null;default:'local'" json:"user_type"`
    DisplayName  string     `gorm:"type:varchar(255)" json:"display_name"`
    CreatedAt    time.Time  `gorm:"default:now()" json:"created_at"`
    UpdatedAt    time.Time  `gorm:"default:now()" json:"updated_at"`
}
```

### 数据库迁移

```sql
-- backend/api-gateway/migrations/20260203XXXXXX_add_user_type.up.sql
ALTER TABLE users ADD COLUMN user_type VARCHAR(50) DEFAULT 'local' NOT NULL;
ALTER TABLE users ADD COLUMN display_name VARCHAR(255);

CREATE INDEX idx_users_user_type ON users(user_type);

-- 添加注释
COMMENT ON COLUMN users.user_type IS '用户类型: local, ldap';
COMMENT ON COLUMN users.display_name IS '显示名称';
```

### LDAP 登录 Service

```go
// backend/api-gateway/internal/service/auth_service.go (扩展)
package service

import (
    "context"
    "errors"
    "time"

    ldapclient "github.com/wangjialin/myops/pkg/ldap"
    authv1 "github.com/wangjialin/myops/pkg/proto/auth"
    "github.com/wangjialin/myops/pkg/auth"
    "github.com/wangjialin/myops/pkg/db"
    "github.com/wangjialin/myops/pkg/model"
)

type AuthService struct {
    userRepo    *db.UserRepository
    jwtManager  *auth.JWTManager
    tokenRepo   *auth.RefreshTokenRepository
    ldapConfig  *ldapclient.Config
    authv1.UnimplementedAuthServiceServer
}

// LDAPLogin LDAP 登录
func (s *AuthService) LDAPLogin(ctx context.Context, req *authv1.LdapLoginRequest) (*authv1.LdapLoginResponse, error) {
    // 1. 连接 LDAP 服务器
    ldapClient, err := ldapclient.NewClient(*s.ldapConfig)
    if err != nil {
        return nil, status.Error(codes.Internal, "LDAP_CONNECTION_FAILED")
    }
    defer ldapClient.Close()

    // 2. 验证用户凭据
    authenticated, err := ldapClient.Authenticate(req.Username, req.Password)
    if err != nil {
        return nil, status.Error(codes.Internal, "LDAP_ERROR")
    }
    if !authenticated {
        return nil, status.Error(codes.Unauthenticated, "LDAP_AUTH_FAILED")
    }

    // 3. 获取用户属性
    attrs, err := ldapClient.GetUserAttributes(req.Username)
    if err != nil {
        return nil, status.Error(codes.Internal, "LDAP_USER_FETCH_FAILED")
    }

    // 4. 查找或创建用户
    user, err := s.userRepo.FindByUsername(ctx, req.Username)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // 自动创建用户
            user = &model.User{
                Username:    attrs.Username,
                Email:       attrs.Email,
                UserType:    model.UserTypeLDAP,
                DisplayName: attrs.DisplayName,
            }
            if err := s.userRepo.Create(ctx, user); err != nil {
                return nil, status.Error(codes.Internal, "USER_CREATE_FAILED")
            }
        } else {
            return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
        }
    }

    // 验证用户类型
    if user.UserType != model.UserTypeLDAP {
        return nil, status.Error(codes.Unauthenticated, "USER_NOT_LDAP")
    }

    // 5. 生成 Token（与本地登录相同）
    accessToken, err := s.jwtManager.GenerateAccessToken(user.ID.String(), user.Username)
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    tokenID, userID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    err = s.tokenRepo.Store(ctx, userID, tokenID, 30*24*time.Hour)
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    // 6. 返回响应
    return &authv1.LdapLoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    3600,
        User: &authv1.User{
            Id:       user.ID.String(),
            Username: user.Username,
            Email:    user.Email,
        },
    }, nil
}
```

### Protobuf 定义

```protobuf
// backend/pkg/proto/auth/auth.proto (扩展)

service AuthService {
  // ... Register, Login, RefreshToken

  // LDAP 登录
  rpc LdapLogin(LdapLoginRequest) returns (LdapLoginResponse) {
    option (google.api.http) = {
      post: "/api/v1/auth/ldap-login"
      body: "*"
    };
  }
}

// LDAP 登录请求
message LdapLoginRequest {
  string username = 1;
  string password = 2;
}

// LDAP 登录响应
message LdapLoginResponse {
  string access_token = 1;
  string refresh_token = 2;
  int32 expires_in = 3;
  User user = 4;
}
```

### 前端登录页面改造

```tsx
// frontend/src/pages/LoginPage.tsx (更新)
import { useState } from 'react'
import { Form, Input, Button, Card, Typography, message, Tabs } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/store/authStore'
import { loginAPI, ldapLoginAPI } from '@/api/auth'

const { Link } = Typography

export default function LoginPage() {
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('local')
  const [localForm] = Form.useForm()
  const [ldapForm] = Form.useForm()
  const navigate = useNavigate()
  const login = useAuthStore((state) => state.login)

  const handleLocalLogin = async (values: any) => {
    setLoading(true)
    try {
      const response = await loginAPI(values)
      login(response.data.accessToken, response.data.user)
      message.success('登录成功')
      navigate('/dashboard')
    } catch (error: any) {
      message.error(error.response?.data?.error?.message || '登录失败')
      localForm.setFieldsValue({ password: '' })
    } finally {
      setLoading(false)
    }
  }

  const handleLdapLogin = async (values: any) => {
    setLoading(true)
    try {
      const response = await ldapLoginAPI(values)
      login(response.data.accessToken, response.data.user)
      message.success('登录成功')
      navigate('/dashboard')
    } catch (error: any) {
      message.error(error.response?.data?.error?.message || 'LDAP 登录失败')
      ldapForm.setFieldsValue({ password: '' })
    } finally {
      setLoading(false)
    }
  }

  const renderLoginForm = (isLdap: boolean) => (
    <Form
      form={isLdap ? ldapForm : localForm}
      name={isLdap ? 'ldap-login' : 'local-login'}
      onFinish={isLdap ? handleLdapLogin : handleLocalLogin}
      autoComplete="off"
      size="large"
    >
      <Form.Item
        name="username"
        rules={[{ required: true, message: '请输入用户名' }]}
      >
        <Input prefix={<UserOutlined />} placeholder="用户名" />
      </Form.Item>

      <Form.Item
        name="password"
        rules={[{ required: true, message: '请输入密码' }]}
      >
        <Input.Password prefix={<LockOutlined />} placeholder="密码" />
      </Form.Item>

      <Form.Item>
        <Button type="primary" htmlType="submit" loading={loading} block>
          登录
        </Button>
      </Form.Item>

      {!isLdap && (
        <Form.Item style={{ marginBottom: 0 }}>
          <Typography.Text style={{ fontSize: '12px' }}>
            忘记密码？<Link onClick={() => message.info('功能即将推出')}>找回密码</Link>
          </Typography.Text>
        </Form.Item>
      )}
    </Form>
  )

  return (
    <div style={styles.container}>
      <Card style={styles.card} title="MyOps AIOps Platform">
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          centered
          items={[
            {
              key: 'local',
              label: '本地账号',
              children: renderLoginForm(false),
            },
            {
              key: 'ldap',
              label: 'LDAP 账号',
              children: renderLoginForm(true),
            },
          ]}
        />
      </Card>
    </div>
  )
}
```

### LDAP API 函数

```typescript
// frontend/src/api/auth.ts (扩展)
export async function ldapLoginAPI(data: LoginRequest): Promise<LoginResponse> {
  const response = await axios.post<LoginResponse>(
    `${API_BASE_URL}/api/v1/auth/ldap-login`,
    data
  )
  return response.data
}
```

### 配置文件

```yaml
# config/config.yaml
ldap:
  url: "ldap://ldap.example.com:389"
  bind_dn: "cn=admin,dc=example,dc=com"
  bind_password: "secret"
  base_dn: "dc=example,dc=com"
  timeout: 10s
  user_filter: "(uid=%s)"
  user_scope: "sub"
  attributes:
    username: "uid"
    email: "mail"
    display_name: "cn"
    first_name: "givenName"
    last_name: "sn"
```

### 安全注意事项

1. **LDAPS**：生产环境必须使用 LDAPS (636) 或 StartTLS
2. **证书验证**：不应跳过证书验证
3. **密码保护**：配置文件中的 bind_password 应加密存储
4. **错误消息**：避免泄露 LDAP 服务器信息
5. **速率限制**：LDAP 认证应实施速率限制

### Dev Agent Guardrails

1. **必须使用 go-ldap/v3 库**
2. **必须支持 LDAP 和 LDAPS 协议**
3. **必须实现用户自动创建**
4. **必须标记用户类型为 ldap**
5. **必须实现连接超时**
6. **必须正确处理连接错误**
7. **前端必须提供标签页切换**
8. **必须编写单元测试**

### Dependencies

**Go 依赖**：
```
- github.com/go-ldap/ldap/v3 (LDAP 客户端)
- github.com/google/uuid (UUID)
- gorm.io/gorm (ORM)
```

### References

- [Source: docs/plans/2025-02-02-aiops-platform-design.md#13-安全与权限-认证]
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
