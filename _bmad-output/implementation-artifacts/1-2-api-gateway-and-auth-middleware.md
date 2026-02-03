# Story 1.2: API Gateway 与认证中间件

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

作为一名开发者，
我想要搭建 API Gateway 和 JWT 认证中间件，
以便所有 API 请求都经过统一的认证和协议转换。

## Acceptance Criteria

**Given** 项目结构已初始化
**When** 启动 API Gateway 服务
**Then** API Gateway 监听在 `:8080` 端口

**And** 实现以下中间件：
- JWT 认证中间件：验证 `Authorization: Bearer <token>` 头
- 日志中间件：使用 zap 记录所有请求（method、path、status、latency）
- 限流中间件：基于 IP 的令牌桶限流（100 req/min）
- CORS 中间件：允许前端域名跨域访问
- 恢复中间件：捕获 panic 并返回 500 错误

**And** gRPC-Gateway 正确配置：
- Protobuf 文件放在 `backend/pkg/proto/` 目录
- 生成的 gRPC-Gateway 代码可编译
- REST 端点正确映射到 gRPC 服务

**And** 健康检查端点响应：
- `GET /health` 返回 `{"status": "ok"}`

**And** 未认证请求返回 401：
```json
{
  "error": {
    "code": "UNAUTHENTICATED",
    "message": "未提供有效的认证令牌"
  },
  "requestId": "req-xxx"
}
```

## Tasks / Subtasks

- [ ] 创建 API Gateway 项目结构 (AC: #)
  - [ ] 创建 `backend/api-gateway/internal/` 目录结构
  - [ ] 创建 `internal/server/` - HTTP 服务器
  - [ ] 创建 `internal/middleware/` - 中间件目录
  - [ ] 创建 `internal/handler/` - HTTP 处理器
  - [ ] 创建 `internal/config/` - 配置加载
- [ ] 实现 HTTP 服务器和路由 (AC: #)
  - [ ] 创建 `internal/server/server.go` 实现 HTTP 服务器
  - [ ] 配置服务器监听在 `:8080` 端口
  - [ ] 实现优雅关闭（graceful shutdown）
  - [ ] 添加健康检查端点 `GET /health`
- [ ] 实现 JWT 认证中间件 (AC: #)
  - [ ] 创建 `internal/middleware/auth.go`
  - [ ] 实现 JWT token 验证逻辑
  - [ ] 从 `Authorization: Bearer <token>` 头提取 token
  - [ ] 验证 token 签名和过期时间
  - [ ] 无效 token 返回 401 响应
  - [ ] 将用户信息存入 context
- [ ] 实现日志中间件 (AC: #)
  - [ ] 创建 `internal/middleware/logger.go`
  - [ ] 使用 zap 记录请求信息（method、path、status、latency）
  - [ ] 生成唯一的 request_id
  - [ ] 记录响应到日志
- [ ] 实现限流中间件 (AC: #)
  - [ ] 创建 `internal/middleware/ratelimit.go`
  - [ ] 实现基于 IP 的令牌桶算法
  - [ ] 配置限制：100 req/min
  - [ ] 超限返回 429 状态码
- [ ] 实现 CORS 中间件 (AC: #)
  - [ ] 创建 `internal/middleware/cors.go`
  - [ ] 配置允许的源、方法、头部
  - [ ] 支持预检请求（OPTIONS）
- [ ] 实现恢复中间件 (AC: #)
  - [ ] 创建 `internal/middleware/recovery.go`
  - [ ] 捕获 panic 并记录日志
  - [ ] 返回 500 错误响应
- [ ] 配置 gRPC-Gateway (AC: #)
  - [ ] 创建 `backend/pkg/proto/` 目录
  - [ ] 创建基础 Protobuf 定义文件
  - [ ] 配置 protoc 生成代码
  - [ ] 生成 gRPC 和 gRPC-Gateway 代码
- [ ] 更新依赖和配置 (AC: #)
  - [ ] 更新 `backend/api-gateway/go.mod` 添加 JWT 库
  - [ ] 更新 `backend/api-gateway/go.mod` 添加 HTTP router
  - [ ] 添加配置文件示例

## Dev Notes

### 项目背景

这是 Epic 1 的第二个 Story，负责搭建 API Gateway 基础设施。API Gateway 是所有外部请求的统一入口，负责认证、限流、日志、协议转换等横切关注点。

### 架构设计

#### API Gateway 职责

1. **认证和授权**：验证 JWT Token
2. **限流**：防止 API 滥用
3. **日志**：记录所有请求和响应
4. **协议转换**：REST → gRPC
5. **CORS 处理**：支持跨域请求
6. **错误恢复**：防止服务崩溃

#### 技术选型

**HTTP 框架**：
- 使用标准库 `net/http` + 路由库（如 `gorilla/mux` 或 `chi`）
- 保持轻量级，避免过度封装

**JWT 库**：
- `github.com/golang-jwt/jwt/v5` - 官方推荐的 JWT 库

**日志库**：
- `go.uber.org/zap` - 高性能结构化日志

**限流库**：
- `golang.org/x/time/rate` - 标准库限流实现

**gRPC-Gateway**：
- `github.com/grpc-ecosystem/grpc-gateway/v2` - REST 到 gRPC 的代理

### 目录结构

```
backend/api-gateway/
├── cmd/server/
│   └── main.go              # 入口文件
├── internal/
│   ├── server/
│   │   └── server.go        # HTTP 服务器
│   ├── middleware/
│   │   ├── auth.go          # JWT 认证
│   │   ├── logger.go        # 日志记录
│   │   ├── ratelimit.go     # 限流
│   │   ├── cors.go          # CORS
│   │   └── recovery.go      # Panic 恢复
│   ├── handler/
│   │   └── health.go        # 健康检查处理器
│   └── config/
│       └── config.go        # 配置加载
├── migrations/              # 数据库迁移（已在 Story 1.1 创建）
└── go.mod
```

### 中间件设计

#### 中间件执行顺序

```
Request → Recovery → Logger → RateLimit → CORS → Auth → Handler → Response
```

#### 1. Recovery 中间件

```go
func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                logger.Error("panic recovered", zap.Any("error", err))
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

#### 2. Logger 中间件

```go
func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := r.Context().Value("request_id")

        // 使用 ResponseWriter 包装器捕获状态码
        ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

        next.ServeHTTP(ww, r)

        logger.Info("request completed",
            zap.String("request_id", requestID),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.Int("status", ww.status),
            zap.Duration("latency", time.Since(start)),
        )
    })
}
```

#### 3. RateLimit 中间件

```go
type IPRateLimiter struct {
    ips map[string]*rate.Limiter
    mu  sync.Mutex
    r   rate.Limit
    b   int
}

func (rl *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    if limiter, exists := rl.ips[ip]; exists {
        return limiter
    }

    limiter := rate.NewLimiter(rl.r, rl.b)
    rl.ips[ip] = limiter
    return limiter
}

func RateLimit(next http.Handler) http.Handler {
    limiter := &IPRateLimiter{
        ips: make(map[string]*rate.Limiter),
        r:   rate.Every(time.Minute / 100), // 100 req/min
        b:   10, // burst
    }

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr
        if !limiter.GetLimiter(ip).Allow() {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

#### 4. CORS 中间件

```go
func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")

        // 允许的源（从配置读取）
        allowedOrigins := []string{"http://localhost:3000"}

        if contains(allowedOrigins, origin) {
            w.Header().Set("Access-Control-Allow-Origin", origin)
        }

        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Credentials", "true")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

#### 5. Auth 中间件

```go
const (
    authorizationHeader = "Authorization"
    bearerScheme        = "Bearer "
)

func Auth(authService AuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 跳过健康检查端点
            if r.URL.Path == "/health" {
                next.ServeHTTP(w, r)
                return
            }

            authHeader := r.Header.Get(authorizationHeader)
            if authHeader == "" {
                respondWithError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "未提供认证令牌")
                return
            }

            if !strings.HasPrefix(authHeader, bearerScheme) {
                respondWithError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "无效的认证格式")
                return
            }

            token := strings.TrimPrefix(authHeader, bearerScheme)
            claims, err := validateToken(token)
            if err != nil {
                respondWithError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "令牌验证失败")
                return
            }

            // 将用户信息存入 context
            ctx := context.WithValue(r.Context(), "user_id", claims.Subject)
            ctx = context.WithValue(ctx, "username", claims.Username)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func respondWithError(w http.ResponseWriter, status int, code, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": map[string]interface{}{
            "code":    code,
            "message": message,
        },
        "requestId": generateRequestID(),
    })
}
```

### JWT 设计

#### Token 结构

```go
type Claims struct {
    Subject  string `json:"sub"`      // 用户 ID
    Username string `json:"username"` // 用户名
    jwt.RegisteredClaims
}

type RegisteredClaims struct {
    ExpiresAt time.Time `json:"exp"` // 过期时间
    IssuedAt  time.Time `json:"iat"` // 签发时间
    NotBefore time.Time `json:"nbf"` // 生效时间
}
```

#### 密钥管理

```go
// 开发环境使用固定的 RSA 密钥对
// 生产环境应从环境变量或密钥管理系统读取

var (
    privateKey *rsa.PrivateKey
    publicKey  *rsa.PublicKey
)

func initJWTKeys() error {
    // 从配置文件加载密钥
    // 开发环境可以生成测试密钥
}
```

### gRPC-Gateway 配置

#### Protobuf 目录结构

```
backend/pkg/proto/
├── auth/
│   └── auth.proto           # 认证服务定义
├── host/
│   └── host.proto           # 主机服务定义
├── k8s/
│   └── k8s.proto            # K8s 服务定义
└── buf.yaml                 # buf 工具配置
```

#### auth.proto 示例

```protobuf
syntax = "proto3";

package myops.auth.v1;

import "google/api/annotations.proto";

option go_package = "github.com/wangjialin/myops/pkg/proto/auth";

service AuthService {
  rpc Register(RegisterRequest) returns (RegisterResponse) {
    option (google.api.http) = {
      post: "/api/v1/auth/register"
      body: "*"
    };
  }

  rpc Login(LoginRequest) returns (LoginResponse) {
    option (google.api.http) = {
      post: "/api/v1/auth/login"
      body: "*"
    };
  }
}

message RegisterRequest {
  string username = 1;
  string email = 2;
  string password = 3;
}

message RegisterResponse {
  User user = 1;
}

message LoginRequest {
  string username = 1;
  string password = 2;
}

message LoginResponse {
  string access_token = 1;
  string refresh_token = 2;
  int32 expires_in = 3;
  User user = 4;
}

message User {
  string id = 1;
  string username = 2;
  string email = 3;
}
```

### 服务器实现

#### server.go

```go
package server

import (
    "context"
    "net/http"
    "time"

    "github.com/wangjialin/myops/api-gateway/internal/middleware"
    "github.com/wangjialin/myops/api-gateway/internal/handler"
    "go.uber.org/zap"
)

type Server struct {
    httpServer *http.Server
    logger     *zap.Logger
}

func New(addr string, logger *zap.Logger) *Server {
    mux := http.NewServeMux()

    // 注册处理器
    mux.HandleFunc("/health", handler.Health)

    // 应用中间件
    handler := middleware.Chain(
        middleware.Recovery(logger),
        middleware.Logger(logger),
        middleware.RateLimit(),
        middleware.CORS(),
        middleware.Auth(), // 跳过健康检查
    )(mux)

    httpServer := &http.Server{
        Addr:         addr,
        Handler:      handler,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    return &Server{
        httpServer: httpServer,
        logger:     logger,
    }
}

func (s *Server) Start() error {
    s.logger.Info("starting API Gateway", zap.String("addr", s.httpServer.Addr))
    return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    s.logger.Info("shutting down API Gateway")
    return s.httpServer.Shutdown(ctx)
}
```

### 配置管理

#### config.go

```go
package config

type Config struct {
    Server   ServerConfig   `yaml:"server"`
    JWT      JWTConfig      `yaml:"jwt"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
}

type ServerConfig struct {
    Host            string `yaml:"host"`
    Port            int    `yaml:"port"`
    ReadTimeout     int    `yaml:"read_timeout"`
    WriteTimeout    int    `yaml:"write_timeout"`
    ShutdownTimeout int    `yaml:"shutdown_timeout"`
}

type JWTConfig struct {
    Secret     string `yaml:"secret"`      // HMAC 密钥或 RSA 密钥路径
    Algorithm  string `yaml:"algorithm"`   // RS256 或 HS256
    ExpiresIn  int    `yaml:"expires_in"`  // access token 过期时间（秒）
    RefreshIn  int    `yaml:"refresh_in"`  // refresh token 过期时间（秒）
    Issuer     string `yaml:"issuer"`
}

type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
    Database string `yaml:"database"`
}

type RedisConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
}
```

### 环境变量

创建 `.env.example`：

```env
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=15
SERVER_WRITE_TIMEOUT=15

# JWT
JWT_ALGORITHM=RS256
JWT_PRIVATE_KEY_PATH=/etc/myops/jwt/private.pem
JWT_PUBLIC_KEY_PATH=/etc/myops/jwt/public.pem
JWT_EXPIRES_IN=3600
JWT_REFRESH_IN=2592000
JWT_ISSUER=myops

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=myops
DB_PASSWORD=myops_dev_pass
DB_NAME=myops

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

### 测试要求

#### 单元测试

每个中间件和处理器都应有单元测试：

```go
func TestAuthMiddleware_ValidToken(t *testing.T) {
    // 生成有效 token
    // 创建测试请求
    // 验证响应
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
    // 使用无效 token
    // 验证返回 401
}

func TestRateLimitMiddleware(t *testing.T) {
    // 发送超过限制的请求
    // 验证返回 429
}
```

#### 集成测试

```go
func TestServer_Integration(t *testing.T) {
    // 启动测试服务器
    // 发送请求到 /health
    // 验证响应
}
```

### 安全注意事项

1. **密钥管理**：RSA 密钥对不应提交到代码仓库
2. **错误消息**：避免在错误响应中泄露敏感信息
3. **限流配置**：根据实际需求调整限流阈值
4. **CORS 配置**：生产环境应明确指定允许的源
5. **请求日志**：注意不要记录敏感信息（密码、token）

### 性能考虑

1. **连接池**：HTTP 客户端使用连接池
2. **日志缓冲**：使用 zap 的缓冲写入
3. **限流器清理**：定期清理不再使用的 IP 限流器
4. **Context 超时**：所有外部调用应设置超时

### Dev Agent Guardrails

1. **必须使用 gorilla/mux 或类似路由库**（不使用 gin、echo 等重型框架）
2. **必须使用 zap 记录结构化日志**
3. **必须实现所有 5 个中间件**
4. **JWT 必须使用 RS256 算法**
5. **必须生成唯一的 request_id**
6. **必须实现优雅关闭**
7. **中间件必须按正确顺序执行**
8. **健康检查端点不需要认证**

### Dependencies

**Go 依赖**：
```
- github.com/golang-jwt/jwt/v5 (JWT)
- github.com/gorilla/mux (HTTP 路由)
- go.uber.org/zap (日志)
- golang.org/x/time/rate (限流)
- github.com/grpc-ecosystem/grpc-gateway/v2 (gRPC-Gateway)
- google.golang.org/grpc (gRPC)
- google.golang.org/protobuf (Protobuf)
- github.com/google/uuid (UUID 生成)
```

### References

- [Source: docs/plans/2025-02-02-aiops-platform-design.md#8-通信协议]
- [Source: _bmad-output/planning-artifacts/epics.md#Epic-1]
- [Source: _bmad-output/implementation-artifacts/1-1-project-initialization-and-database-setup.md]

## Dev Agent Record

### Agent Model Used

glm-4.7 (claude-opus-4-5-20251101)

### Debug Log References

None (story preparation phase)

### Completion Notes List

**Story Created (2026-02-02)**:

Story prepared for development with detailed implementation guidance.
