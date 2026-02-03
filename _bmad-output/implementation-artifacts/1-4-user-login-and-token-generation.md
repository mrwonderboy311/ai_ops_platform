# Story 1.4: 用户登录与 Token 生成

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

作为一名已注册用户，
我想要登录并获取访问令牌，
以便使用平台功能。

## Acceptance Criteria

**Given** 用户已注册
**When** 发送 POST 请求到 `/api/v1/auth/login`
```json
{
  "username": "testuser",
  "password": "SecurePass123!"
}
```
**Then** 返回 200 状态码
```json
{
  "data": {
    "accessToken": "eyJhbGc...",
    "refreshToken": "eyJhbGc...",
    "expiresIn": 3600,
    "user": {
      "id": "uuid",
      "username": "testuser",
      "email": "test@example.com"
    }
  },
  "requestId": "req-xxx"
}
```

**And** JWT Token 包含以下 claims：
- `sub`: 用户 ID
- `username`: 用户名
- `exp`: 过期时间（1 小时）
- `iat`: 签发时间

**And** Token 使用 RS256 算法签名

**And** 登录失败处理：
- 用户不存在返回 401：`{"error": {"code": "INVALID_CREDENTIALS", "message": "用户名或密码错误"}}`
- 密码错误返回 401：同上（防止用户名枚举）

**And** Refresh Token 存储在 Redis：
- Key: `refresh_token:{user_id}`
- TTL: 30 天

**And** 实现刷新 Token 端点 `POST /api/v1/auth/refresh`

## Tasks / Subtasks

- [ ] 实现 JWT Token 生成 (AC: #)
  - [ ] 创建 `backend/pkg/auth/jwt.go`
  - [ ] 实现 `GenerateAccessToken` 函数（RS256）
  - [ ] 实现 `GenerateRefreshToken` 函数
  - [ ] 实现 `ValidateToken` 函数
  - [ ] 定义 JWT Claims 结构体
- [ ] 生成 RSA 密钥对 (AC: #)
  - [ ] 创建密钥生成脚本 `backend/scripts/generate-jwt-keys.sh`
  - [ ] 生成 2048 位 RSA 私钥
  - [ ] 导出对应的公钥
  - [ ] 存储在 `backend/api-gateway/keys/` 目录
- [ ] 实现 Refresh Token 存储 (AC: #)
  - [ ] 创建 `backend/pkg/auth/redis.go`
  - [ ] 实现 `StoreRefreshToken` 函数
  - [ ] 实现 `GetRefreshToken` 函数
  - [ ] 实现 `DeleteRefreshToken` 函数
  - [ ] 配置 Redis 连接
- [ ] 实现登录 Service (AC: #)
  - [ ] 在 `backend/api-gateway/internal/service/auth_service.go` 添加 `Login` 方法
  - [ ] 根据用户名查找用户
  - [ ] 验证密码（bcrypt 比较）
  - [ ] 生成 Access Token（1 小时过期）
  - [ ] 生成 Refresh Token（30 天过期）
  - [ ] 存储 Refresh Token 到 Redis
  - [ ] 返回登录响应
- [ ] 实现刷新 Token Service (AC: #)
  - [ ] 在 auth_service 添加 `RefreshToken` 方法
  - [ ] 验证 Refresh Token
  - [ ] 从 Redis 获取存储的 token
  - [ ] 检查 token 是否匹配
  - [ ] 生成新的 Access Token
  - [ ] 可选：轮换 Refresh Token
- [ ] 实现 HTTP 处理器 (AC: #)
  - [ ] 创建 `backend/api-gateway/internal/handler/login.go`
  - [ ] 实现 login POST 处理器
  - [ ] 创建 `backend/api-gateway/internal/handler/refresh.go`
  - [ ] 实现 refresh POST 处理器
- [ ] 集成到 API Gateway (AC: #)
  - [ ] 注册路由 `POST /api/v1/auth/login`
  - [ ] 注册路由 `POST /api/v1/auth/refresh`
  - [ ] 应用日志和恢复中间件
  - [ ] **不需要认证中间件**（公开端点）
- [ ] 配置 Redis 连接 (AC: #)
  - [ ] 更新配置结构添加 Redis 配置
  - [ ] 实现 Redis 客户端初始化
  - [ ] 添加连接池配置
  - [ ] 实现优雅关闭
- [ ] 编写单元测试 (AC: #)
  - [ ] 测试 JWT 生成和验证
  - [ ] 测试过期 Token 验证
  - [ ] 测试 Refresh Token 存储和检索
  - [ ] 测试登录成功场景
  - [ ] 测试用户不存在
  - [ ] 测试密码错误
- [ ] 编写集成测试 (AC: #)
  - [ ] 测试完整登录流程
  - [ ] 测试刷新 Token 流程
  - [ ] 测试无效 Refresh Token

## Dev Notes

### 项目背景

这是 Epic 1 的第四个 Story，负责实现用户登录和 JWT Token 生成。登录是认证流程的核心，需要生成符合规范的 JWT Token 并管理 Refresh Token。

### 业务逻辑

#### 登录流程

```
1. 接收请求 → 2. 查找用户 → 3. 验证密码 → 4. 生成 Token → 5. 存储 Refresh Token → 6. 返回响应
```

#### 刷新 Token 流程

```
1. 接收请求 → 2. 验证 Refresh Token → 3. 检查 Redis 存储 → 4. 生成新 Access Token → 5. 返回响应
```

### JWT 设计

#### Claims 结构

```go
// backend/pkg/auth/jwt.go
package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

// Claims 自定义 JWT Claims
type Claims struct {
    Subject  string `json:"sub"`      // 用户 ID
    Username string `json:"username"` // 用户名
    jwt.RegisteredClaims
}

// RegisteredClaims 标准声明
type RegisteredClaims struct {
    ExpiresAt time.Time `json:"exp"` // 过期时间
    IssuedAt  time.Time `json:"iat"` // 签发时间
    NotBefore time.Time `json:"nbf"` // 生效时间
}

// TokenConfig Token 配置
type TokenConfig struct {
    Issuer          string        // 签发者
    AccessDuration  time.Duration // Access Token 有效期
    RefreshDuration time.Duration // Refresh Token 有效期
}

type JWTManager struct {
    privateKey *rsa.PrivateKey
    publicKey  *rsa.PublicKey
    config     TokenConfig
}
```

#### JWT Manager 实现

```go
// backend/pkg/auth/jwt.go
package auth

import (
    "crypto/rsa"
    errors2 "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

type JWTManager struct {
    privateKey *rsa.PrivateKey
    publicKey  *rsa.PublicKey
    config     TokenConfig
}

func NewJWTManager(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, config TokenConfig) *JWTManager {
    return &JWTManager{
        privateKey: privateKey,
        publicKey:  publicKey,
        config:     config,
    }
}

// GenerateAccessToken 生成 Access Token
func (m *JWTManager) GenerateAccessToken(userID, username string) (string, error) {
    now := time.Now()
    claims := Claims{
        Subject:  userID,
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(m.config.AccessDuration)),
            IssuedAt:  jwt.NewNumericDate(now),
            NotBefore: jwt.NewNumericDate(now),
            Issuer:    m.config.Issuer,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(m.privateKey)
}

// GenerateRefreshToken 生成 Refresh Token
func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
    // Refresh Token 使用 UUID
    tokenID := uuid.New().String()

    // 将 tokenID 存储为 claims，以便可以撤销
    now := time.Now()
    claims := jwt.MapClaims{
        "sub": tokenID,
        "uid": userID,
        "typ": "refresh",
        "exp": now.Add(m.config.RefreshDuration).Unix(),
        "iat": now.Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(m.privateKey)
}

// ValidateAccessToken 验证 Access Token
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, errors2.New("unexpected signing method")
        }
        return m.publicKey, nil
    })

    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors2.New("invalid token")
    }

    return claims, nil
}

// ValidateRefreshToken 验证 Refresh Token 并返回用户 ID
func (m *JWTManager) ValidateRefreshToken(tokenString string) (string, string, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, errors2.New("unexpected signing method")
        }
        return m.publicKey, nil
    })

    if err != nil {
        return "", "", err
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return "", "", errors2.New("invalid token")
    }

    tokenType, ok := claims["typ"].(string)
    if !ok || tokenType != "refresh" {
        return "", "", errors2.New("not a refresh token")
    }

    tokenID, _ := claims["sub"].(string)
    userID, _ := claims["uid"].(string)

    return tokenID, userID, nil
}
```

### RSA 密钥生成

#### 密钥生成脚本

```bash
#!/bin/bash
# backend/scripts/generate-jwt-keys.sh

set -e

KEYS_DIR="./backend/api-gateway/keys"
mkdir -p "$KEYS_DIR"

# 生成 2048 位 RSA 私钥
openssl genrsa -out "$KEYS_DIR/private.pem" 2048

# 导出公钥
openssl rsa -in "$KEYS_DIR/private.pem" -pubout -out "$KEYS_DIR/public.pem"

# 设置权限
chmod 600 "$KEYS_DIR/private.pem"
chmod 644 "$KEYS_DIR/public.pem"

echo "JWT keys generated successfully:"
echo "  Private: $KEYS_DIR/private.pem"
echo "  Public:  $KEYS_DIR/public.pem"
```

### Redis Token 存储

#### Refresh Token Repository

```go
// backend/pkg/auth/redis.go
package auth

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RefreshTokenRepository struct {
    client *redis.Client
}

func NewRefreshTokenRepository(client *redis.Client) *RefreshTokenRepository {
    return &RefreshTokenRepository{client: client}
}

// Store 存储 Refresh Token
func (r *RefreshTokenRepository) Store(ctx context.Context, userID, tokenID string, ttl time.Duration) error {
    key := fmt.Sprintf("refresh_token:%s", userID)
    return r.client.Set(ctx, key, tokenID, ttl).Err()
}

// Get 获取 Refresh Token
func (r *RefreshTokenRepository) Get(ctx context.Context, userID string) (string, error) {
    key := fmt.Sprintf("refresh_token:%s", userID)
    return r.client.Get(ctx, key).Result()
}

// Delete 删除 Refresh Token（注销）
func (r *RefreshTokenRepository) Delete(ctx context.Context, userID string) error {
    key := fmt.Sprintf("refresh_token:%s", userID)
    return r.client.Del(ctx, key).Err()
}

// Verify 验证 Refresh Token 是否匹配
func (r *RefreshTokenRepository) Verify(ctx context.Context, userID, tokenID string) (bool, error) {
    stored, err := r.Get(ctx, userID)
    if err != nil {
        if err == redis.Nil {
            return false, nil
        }
        return false, err
    }
    return stored == tokenID, nil
}
```

### 登录 Service 实现

```go
// backend/api-gateway/internal/service/auth_service.go (扩展)
package service

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"
    authv1 "github.com/wangjialin/myops/pkg/proto/auth"
    "github.com/wangjialin/myops/pkg/auth"
    "github.com/wangjialin/myops/pkg/db"
    "github.com/wangjialin/myops/pkg/model"
)

type AuthService struct {
    userRepo    *db.UserRepository
    jwtManager  *auth.JWTManager
    tokenRepo   *auth.RefreshTokenRepository
    authv1.UnimplementedAuthServiceServer
}

func NewAuthService(
    userRepo *db.UserRepository,
    jwtManager *auth.JWTManager,
    tokenRepo *auth.RefreshTokenRepository,
) *AuthService {
    return &AuthService{
        userRepo:   userRepo,
        jwtManager: jwtManager,
        tokenRepo:  tokenRepo,
    }
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
    // 1. 查找用户
    user, err := s.userRepo.FindByUsername(ctx, req.Username)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // 返回通用错误，防止用户名枚举
            return nil, status.Error(codes.Unauthenticated, "INVALID_CREDENTIALS")
        }
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    // 2. 验证密码
    if !auth.ComparePassword(user.PasswordHash, req.Password) {
        return nil, status.Error(codes.Unauthenticated, "INVALID_CREDENTIALS")
    }

    // 3. 生成 Access Token（1 小时）
    accessToken, err := s.jwtManager.GenerateAccessToken(user.ID.String(), user.Username)
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    // 4. 生成 Refresh Token（30 天）
    refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String())
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    // 5. 验证 Refresh Token 并提取 tokenID
    tokenID, userID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    // 6. 存储 Refresh Token 到 Redis
    err = s.tokenRepo.Store(ctx, userID, tokenID, 30*24*time.Hour)
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    // 7. 返回响应
    return &authv1.LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    3600, // 1 小时
        User: &authv1.User{
            Id:       user.ID.String(),
            Username: user.Username,
            Email:    user.Email,
        },
    }, nil
}

// RefreshToken 刷新 Access Token
func (s *AuthService) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
    // 1. 验证 Refresh Token
    tokenID, userID, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "INVALID_REFRESH_TOKEN")
    }

    // 2. 检查 Redis 中是否存在
    valid, err := s.tokenRepo.Verify(ctx, userID, tokenID)
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }
    if !valid {
        return nil, status.Error(codes.Unauthenticated, "REFRESH_TOKEN_EXPIRED")
    }

    // 3. 获取用户信息
    userUUID, err := uuid.Parse(userID)
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    user, err := s.userRepo.FindByID(ctx, userUUID)
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    // 4. 生成新的 Access Token
    accessToken, err := s.jwtManager.GenerateAccessToken(user.ID.String(), user.Username)
    if err != nil {
        return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
    }

    // 5. 可选：轮换 Refresh Token
    // newRefreshToken, err := s.jwtManager.GenerateRefreshToken(userID)
    // ...

    return &authv1.RefreshTokenResponse{
        AccessToken:  accessToken,
        RefreshToken: req.RefreshToken, // 返回原 token 或新 token
        ExpiresIn:    3600,
    }, nil
}
```

### Protobuf 定义更新

```protobuf
// backend/pkg/proto/auth/auth.proto (扩展)

service AuthService {
  // ... Register RPC

  // 用户登录
  rpc Login(LoginRequest) returns (LoginResponse) {
    option (google.api.http) = {
      post: "/api/v1/auth/login"
      body: "*"
    };
  }

  // 刷新 Token
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse) {
    option (google.api.http) = {
      post: "/api/v1/auth/refresh"
      body: "*"
    };
  }
}

// 登录请求
message LoginRequest {
  string username = 1;
  string password = 2;
}

// 登录响应
message LoginResponse {
  string access_token = 1;
  string refresh_token = 2;
  int32 expires_in = 3;
  User user = 4;
}

// 刷新 Token 请求
message RefreshTokenRequest {
  string refresh_token = 1;
}

// 刷新 Token 响应
message RefreshTokenResponse {
  string access_token = 1;
  string refresh_token = 2;
  int32 expires_in = 3;
}
```

### 配置更新

```go
// backend/api-gateway/internal/config/config.go (扩展)

type Config struct {
    Server   ServerConfig   `yaml:"server"`
    JWT      JWTConfig      `yaml:"jwt"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
}

type JWTConfig struct {
    PrivateKeyPath string        `yaml:"private_key_path"`
    PublicKeyPath  string        `yaml:"public_key_path"`
    Issuer         string        `yaml:"issuer"`
    AccessDuration time.Duration `yaml:"access_duration"`
    RefreshDuration time.Duration `yaml:"refresh_duration"`
}

type RedisConfig struct {
    Addr     string `yaml:"addr"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
    PoolSize int    `yaml:"pool_size"`
}
```

配置文件示例 `config/config.yaml`:

```yaml
server:
  host: 0.0.0.0
  port: 8080

jwt:
  private_key_path: ./keys/private.pem
  public_key_path: ./keys/public.pem
  issuer: myops
  access_duration: 1h
  refresh_duration: 720h  # 30 天

database:
  host: localhost
  port: 5432
  user: myops
  password: myops_dev_pass
  database: myops

redis:
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10
```

### 服务器初始化

```go
// backend/api-gateway/cmd/server/main.go (更新)
package main

func main() {
    // 加载配置
    config := config.Load()

    // 初始化日志
    logger := zap.NewProduction()

    // 初始化数据库
    db := db.Connect(config.Database)
    userRepo := db.NewUserRepository(db)

    // 初始化 Redis
    redisClient := redis.NewClient(&redis.Options{
        Addr:     config.Redis.Addr,
        Password: config.Redis.Password,
        DB:       config.Redis.DB,
        PoolSize: config.Redis.PoolSize,
    })

    // 加载 RSA 密钥
    privateKey, publicKey := loadJWTKeys(config.JWT)

    // 初始化 JWT Manager
    jwtManager := auth.NewJWTManager(privateKey, publicKey, auth.TokenConfig{
        Issuer:          config.JWT.Issuer,
        AccessDuration:  config.JWT.AccessDuration,
        RefreshDuration: config.JWT.RefreshDuration,
    })

    // 初始化 Refresh Token Repository
    tokenRepo := auth.NewRefreshTokenRepository(redisClient)

    // 初始化 Service
    authService := service.NewAuthService(userRepo, jwtManager, tokenRepo)

    // 启动服务器
    server := server.New(config.Server, logger, authService)
    server.Start()
}
```

### 测试

#### JWT 测试

```go
// backend/pkg/auth/jwt_test.go
package auth_test

import (
    "testing"
    "time"

    "github.com/wangjialin/myops/pkg/auth"
)

func TestJWTManager(t *testing.T) {
    // 生成测试密钥对
    privateKey, publicKey := generateTestKeys()

    manager := auth.NewJWTManager(privateKey, publicKey, auth.TokenConfig{
        Issuer:          "test",
        AccessDuration:  time.Hour,
        RefreshDuration: 30 * 24 * time.Hour,
    })

    t.Run("Generate and Validate Access Token", func(t *testing.T) {
        token, err := manager.GenerateAccessToken("user-123", "testuser")
        if err != nil {
            t.Fatalf("GenerateAccessToken failed: %v", err)
        }

        claims, err := manager.ValidateAccessToken(token)
        if err != nil {
            t.Fatalf("ValidateAccessToken failed: %v", err)
        }

        if claims.Subject != "user-123" {
            t.Errorf("Expected subject 'user-123', got '%s'", claims.Subject)
        }
        if claims.Username != "testuser" {
            t.Errorf("Expected username 'testuser', got '%s'", claims.Username)
        }
    })

    t.Run("Generate Refresh Token", func(t *testing.T) {
        token, err := manager.GenerateRefreshToken("user-123")
        if err != nil {
            t.Fatalf("GenerateRefreshToken failed: %v", err)
        }

        tokenID, userID, err := manager.ValidateRefreshToken(token)
        if err != nil {
            t.Fatalf("ValidateRefreshToken failed: %v", err)
        }

        if userID != "user-123" {
            t.Errorf("Expected userID 'user-123', got '%s'", userID)
        }
        if tokenID == "" {
            t.Error("Expected non-empty tokenID")
        }
    })
}
```

### 安全注意事项

1. **密钥保护**：RSA 私钥文件权限设置为 600
2. **错误消息统一**：登录失败使用相同错误消息防止用户名枚举
3. **Token 过期**：Access Token 短期（1 小时），Refresh Token 长期（30 天）
4. **Redis 持久化**：配置 Redis 持久化防止重启丢失 token
5. **HTTPS 传输**：生产环境必须使用 HTTPS

### Dev Agent Guardrails

1. **必须使用 RS256 算法签名 JWT**
2. **Access Token 必须设置 1 小时过期**
3. **Refresh Token 必须存储在 Redis**
4. **登录失败必须返回统一错误消息**
5. **密码验证必须使用 bcrypt.ComparePassword**
6. **必须实现刷新 Token 端点**
7. **必须配置优雅关闭 Redis 连接**
8. **必须编写单元测试**

### Dependencies

**Go 依赖**：
```
- github.com/golang-jwt/jwt/v5 (JWT)
- github.com/redis/go-redis/v9 (Redis 客户端)
- github.com/google/uuid (UUID)
- gorm.io/gorm (ORM)
- crypto/rsa (标准库)
- crypto/rand (标准库)
```

### References

- [Source: docs/plans/2025-02-02-aiops-platform-design.md#13-安全与权限-认证]
- [Source: _bmad-output/planning-artifacts/epics.md#Epic-1]
- [Source: _bmad-output/implementation-artifacts/1-2-api-gateway-and-auth-middleware.md]
- [Source: _bmad-output/implementation-artifacts/1-3-user-registration-api.md]

## Dev Agent Record

### Agent Model Used

glm-4.7 (claude-opus-4-5-20251101)

### Debug Log References

None (story preparation phase)

### Completion Notes List

**Story Created (2026-02-03)**:

Story prepared for development with detailed implementation guidance.
