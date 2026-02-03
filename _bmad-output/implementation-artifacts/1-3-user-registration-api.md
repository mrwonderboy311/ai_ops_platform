# Story 1.3: 用户注册 API

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

作为一名新用户，
我想要注册一个账号，
以便登录平台使用功能。

## Acceptance Criteria

**Given** API Gateway 和数据库已运行
**When** 发送 POST 请求到 `/api/v1/auth/register`
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "SecurePass123!"
}
```
**Then** 返回 201 状态码
```json
{
  "data": {
    "id": "uuid",
    "username": "testuser",
    "email": "test@example.com"
  },
  "requestId": "req-xxx"
}
```

**And** 密码使用 bcrypt 加密存储（cost factor 10）

**And** 用户名和邮箱唯一性验证：
- 重复用户名返回 400：`{"error": {"code": "USERNAME_EXISTS", "message": "用户名已存在"}}`
- 重复邮箱返回 400：`{"error": {"code": "EMAIL_EXISTS", "message": "邮箱已被注册"}}`

**And** 密码强度验证：
- 最少 8 个字符
- 包含大小写字母、数字
- 不符合返回 400：`{"error": {"code": "WEAK_PASSWORD", "message": "密码强度不足"}}`

**And** 创建 `users` 表索引：
- `idx_users_username` on `username`
- `idx_users_email` on `email`

## Tasks / Subtasks

- [ ] 创建用户数据模型 (AC: #)
  - [ ] 创建 `backend/pkg/model/user.go` 定义 User 结构体
  - [ ] 实现 GORM 模型（表名、字段映射、钩子）
  - [ ] 添加 `BeforeCreate` 钩子更新 `updated_at`
- [ ] 创建数据库层 (AC: #)
  - [ ] 创建 `backend/pkg/db/user.go` 数据库操作
  - [ ] 实现 `CreateUser` 函数
  - [ ] 实现 `FindByUsername` 函数
  - [ ] 实现 `FindByEmail` 函数
  - [ ] 实现 `FindByID` 函数
- [ ] 实现密码加密 (AC: #)
  - [ ] 创建 `backend/pkg/auth/password.go`
  - [ ] 实现 `HashPassword` 使用 bcrypt（cost factor 10）
  - [ ] 实现 `ComparePassword` 验证密码
- [ ] 实现密码验证器 (AC: #)
  - [ ] 创建 `backend/pkg/auth/validator.go`
  - [ ] 实现密码强度验证（最少 8 字符）
  - [ ] 实现大小写字母验证
  - [ ] 实现数字验证
  - [ ] 返回详细的错误信息
- [ ] 实现输入验证 (AC: #)
  - [ ] 创建 `backend/pkg/auth/validator.go`
  - [ ] 验证用户名格式（3-50 字符，字母数字下划线）
  - [ ] 验证邮箱格式（RFC 5322）
  - [ ] 验证必填字段
- [ ] 创建 Protobuf 定义 (AC: #)
  - [ ] 创建 `backend/pkg/proto/auth/auth.proto`
  - [ ] 定义 RegisterRequest 消息
  - [ ] 定义 RegisterResponse 消息
  - [ ] 定义 AuthService 的 Register RPC
  - [ ] 添加 gRPC-Gateway HTTP 映射
- [ ] 实现 gRPC Service (AC: #)
  - [ ] 创建 `backend/api-gateway/internal/service/auth_service.go`
  - [ ] 实现 Register 方法
  - [ ] 处理用户名已存在错误
  - [ ] 处理邮箱已存在错误
  - [ ] 处理密码强度不足错误
  - [ ] 返回创建的用户信息（不含密码）
- [ ] 实现 HTTP 处理器 (AC: #)
  - [ ] 创建 `backend/api-gateway/internal/handler/register.go`
  - [ ] 实现 JSON 请求解析
  - [ ] 调用 gRPC Service
  - [ ] 返回 JSON 响应
  - [ ] 处理错误响应
- [ ] 集成到 API Gateway (AC: #)
  - [ ] 注册路由 `POST /api/v1/auth/register`
  - [ ] 应用日志中间件
  - [ ] 应用恢复中间件
  - [ ] **不需要认证中间件**（公开端点）
- [ ] 编写单元测试 (AC: #)
  - [ ] 测试密码加密功能
  - [ ] 测试密码验证（有效/无效密码）
  - [ ] 测试用户名/邮箱验证
  - [ ] 测试 CreateUser 数据库操作
  - [ ] 测试重复用户名/邮箱错误
- [ ] 编写集成测试 (AC: #)
  - [ ] 测试成功的注册请求
  - [ ] 测试重复用户名
  - [ ] 测试重复邮箱
  - [ ] 测试弱密码
  - [ ] 测试无效邮箱格式
  - [ ] 测试必填字段缺失

## Dev Notes

### 项目背景

这是 Epic 1 的第三个 Story，负责实现用户注册功能。注册 API 是平台的第一个公开 API，允许新用户创建账号。

### 业务逻辑

#### 注册流程

```
1. 接收请求 → 2. 验证输入 → 3. 检查唯一性 → 4. 加密密码 → 5. 创建用户 → 6. 返回响应
```

#### 错误码定义

| 错误码 | HTTP 状态 | 场景 |
|--------|----------|------|
| `USERNAME_EXISTS` | 400 | 用户名已存在 |
| `EMAIL_EXISTS` | 400 | 邮箱已被注册 |
| `WEAK_PASSWORD` | 400 | 密码强度不足 |
| `INVALID_USERNAME` | 400 | 用户名格式无效 |
| `INVALID_EMAIL` | 400 | 邮箱格式无效 |
| `MISSING_FIELDS` | 400 | 必填字段缺失 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |

### 数据模型

#### User 结构体

```go
// backend/pkg/model/user.go
package model

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type User struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    Username     string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"username"`
    Email        string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
    PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"` // 不序列化
    CreatedAt    time.Time `gorm:"default:now()" json:"created_at"`
    UpdatedAt    time.Time `gorm:"default:now()" json:"updated_at"`
}

func (User) TableName() string {
    return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
    return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
    u.UpdatedAt = time.Now()
    return nil
}
```

### 密码加密

#### bcrypt 加密实现

```go
// backend/pkg/auth/password.go
package auth

import (
    "golang.org/x/crypto/bcrypt"
)

const (
    bcryptCost = 10
)

// HashPassword 对密码进行 bcrypt 加密
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}

// ComparePassword 验证密码
func ComparePassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}
```

### 输入验证

#### 密码强度验证

```go
// backend/pkg/auth/validator.go
package auth

import (
    "errors"
    "regexp"
)

var (
    errPasswordTooShort      = errors.New("密码至少需要 8 个字符")
    errPasswordMissingUpper  = errors.New("密码必须包含至少一个大写字母")
    errPasswordMissingLower  = errors.New("密码必须包含至少一个小写字母")
    errPasswordMissingDigit  = errors.New("密码必须包含至少一个数字")
    errInvalidUsernameFormat = errors.New("用户名只能包含字母、数字和下划线，3-50 个字符")
    errInvalidEmailFormat    = errors.New("邮箱格式无效")
)

// PasswordStrength 验证密码强度
func PasswordStrength(password string) error {
    if len(password) < 8 {
        return errPasswordTooShort
    }

    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

    if !hasUpper {
        return errPasswordMissingUpper
    }
    if !hasLower {
        return errPasswordMissingLower
    }
    if !hasDigit {
        return errPasswordMissingDigit
    }

    return nil
}

// UsernameFormat 验证用户名格式
func UsernameFormat(username string) error {
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,50}$`, username)
    if !matched {
        return errInvalidUsernameFormat
    }
    return nil
}

// EmailFormat 验证邮箱格式
func EmailFormat(email string) error {
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
    if !matched {
        return errInvalidEmailFormat
    }
    return nil
}
```

### 数据库操作

#### 用户数据库层

```go
// backend/pkg/db/user.go
package db

import (
    "context"
    "github.com/google/uuid"
    "github.com/wangjialin/myops/pkg/model"
    "gorm.io/gorm"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// FindByID 根据 ID 查找用户
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// ExistsUsername 检查用户名是否存在
func (r *UserRepository) ExistsUsername(ctx context.Context, username string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error
    return count > 0, err
}

// ExistsEmail 检查邮箱是否存在
func (r *UserRepository) ExistsEmail(ctx context.Context, email string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error
    return count > 0, err
}
```

### Protobuf 定义

#### auth.proto

```protobuf
syntax = "proto3";

package myops.auth.v1;

import "google/api/annotations.proto";
import "protoc-gen-openapiv2/options/annotations.proto";

option go_package = "github.com/wangjialin/myops/pkg/proto/auth";

// OpenAPI 配置
option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_swagger) = {
  info: {
    title: "MyOps Auth API";
    version: "1.0";
    contact: {
      name: "MyOps Team";
    };
  };
  schemes: HTTP;
  schemes: HTTPS;
  consumes: "application/json";
  produces: "application/json";
  security_definitions: {
    security: {
      key: "Bearer";
      value: {
        type: TYPE_API_KEY;
        in: IN_HEADER;
        name: "Authorization";
        description: "Bearer token for authentication";
      }
    }
  }
};

service AuthService {
  // 用户注册
  rpc Register(RegisterRequest) returns (RegisterResponse) {
    option (google.api.http) = {
      post: "/api/v1/auth/register"
      body: "*"
    };
  }

  // 用户登录（后续 Story 实现）
  rpc Login(LoginRequest) returns (LoginResponse) {
    option (google.api.http) = {
      post: "/api/v1/auth/login"
      body: "*"
    };
  }
}

// 注册请求
message RegisterRequest {
  string username = 1 [(grpc.gateway.protoc_gen_openapiv2.options.openapiv2_field) = {
    description: "用户名（3-50 字符，字母数字下划线）";
    required: ["username"];
  }];
  string email = 2 [(grpc.gateway.protoc_gen_openapiv2.options.openapiv2_field) = {
    description: "邮箱地址";
    required: ["email"];
  }];
  string password = 3 [(grpc.gateway.protoc_gen_openapiv2.options.openapiv2_field) = {
    description: "密码（最少 8 字符，包含大小写字母和数字）";
    format: "password";
    required: ["password"];
    min_length: 8;
  }];
}

// 注册响应
message RegisterResponse {
  User user = 1;
}

// 登录请求（占位）
message LoginRequest {
  string username = 1;
  string password = 2;
}

// 登录响应（占位）
message LoginResponse {
  string access_token = 1;
  string refresh_token = 2;
  int32 expires_in = 3;
  User user = 4;
}

// 用户信息
message User {
  string id = 1 [(grpc.gateway.protoc_gen_openapiv2.options.openapiv2_field) = {
    description: "用户 ID (UUID)";
    example: "\"550e8400-e29b-41d4-a716-446655440000\"";
  }];
  string username = 2 [(grpc.gateway.protoc_gen_openapiv2.options.openapiv2_field) = {
    description: "用户名";
    example: "\"testuser\"";
  }];
  string email = 3 [(grpc.gateway.protoc_gen_openapiv2.options.openapiv2_field) = {
    description: "邮箱地址";
    example: "\"test@example.com\"";
  }];
}
```

### gRPC Service 实现

```go
// backend/api-gateway/internal/service/auth_service.go
package service

import (
    "context"
    "errors"

    "github.com/google/uuid"
    authv1 "github.com/wangjialin/myops/pkg/proto/auth"
    "github.com/wangjialin/myops/pkg/auth"
    "github.com/wangjialin/myops/pkg/db"
    "github.com/wangjialin/myops/pkg/model"
)

type AuthService struct {
    userRepo *db.UserRepository
    authv1.UnimplementedAuthServiceServer
}

func NewAuthService(userRepo *db.UserRepository) *AuthService {
    return &AuthService{
        userRepo: userRepo,
    }
}

func (s *AuthService) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
    // 1. 验证输入
    if err := auth.UsernameFormat(req.Username); err != nil {
        return nil, invalidArgumentError("INVALID_USERNAME", err.Error())
    }
    if err := auth.EmailFormat(req.Email); err != nil {
        return nil, invalidArgumentError("INVALID_EMAIL", err.Error())
    }
    if err := auth.PasswordStrength(req.Password); err != nil {
        return nil, invalidArgumentError("WEAK_PASSWORD", err.Error())
    }

    // 2. 检查唯一性
    exists, err := s.userRepo.ExistsUsername(ctx, req.Username)
    if err != nil {
        return nil, internalError(err)
    }
    if exists {
        return nil, invalidArgumentError("USERNAME_EXISTS", "用户名已存在")
    }

    exists, err = s.userRepo.ExistsEmail(ctx, req.Email)
    if err != nil {
        return nil, internalError(err)
    }
    if exists {
        return nil, invalidArgumentError("EMAIL_EXISTS", "邮箱已被注册")
    }

    // 3. 加密密码
    passwordHash, err := auth.HashPassword(req.Password)
    if err != nil {
        return nil, internalError(err)
    }

    // 4. 创建用户
    user := &model.User{
        ID:           uuid.New(),
        Username:     req.Username,
        Email:        req.Email,
        PasswordHash: passwordHash,
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, internalError(err)
    }

    // 5. 返回响应（不含密码）
    return &authv1.RegisterResponse{
        User: &authv1.User{
            Id:       user.ID.String(),
            Username: user.Username,
            Email:    user.Email,
        },
    }, nil
}

// 错误辅助函数
func invalidArgumentError(code, message string) error {
    // 返回 gRPC status.Error
}

func internalError(err error) error {
    // 返回 gRPC Internal 错误
}
```

### 错误处理

#### 统一错误响应

```go
// backend/pkg/errors/errors.go
package errors

import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type Error struct {
    Code    string
    Message string
}

func (e *Error) Error() string {
    return e.Message
}

func NewError(code, message string) *Error {
    return &Error{Code: code, Message: message}
}

// 预定义错误
var (
    ErrUsernameExists = NewError("USERNAME_EXISTS", "用户名已存在")
    ErrEmailExists    = NewError("EMAIL_EXISTS", "邮箱已被注册")
    ErrWeakPassword   = NewError("WEAK_PASSWORD", "密码强度不足")
    ErrInvalidEmail   = NewError("INVALID_EMAIL", "邮箱格式无效")
)

// 转换为 gRPC status
func ToGRPCError(err error) error {
    if e, ok := err.(*Error); ok {
        return status.Error(codes.InvalidArgument, e.Code)
    }
    return status.Error(codes.Internal, "INTERNAL_ERROR")
}
```

### 测试

#### 单元测试示例

```go
// backend/pkg/auth/password_test.go
package auth_test

import (
    "testing"
    "github.com/wangjialin/myops/pkg/auth"
)

func TestHashPassword(t *testing.T) {
    password := "SecurePass123!"
    hash, err := auth.HashPassword(password)

    if err != nil {
        t.Fatalf("HashPassword failed: %v", err)
    }

    if hash == password {
        t.Error("Hash should not equal password")
    }

    // 验证可以正确比较
    if !auth.ComparePassword(hash, password) {
        t.Error("ComparePassword should succeed with correct password")
    }

    // 验证错误密码失败
    if auth.ComparePassword(hash, "WrongPassword") {
        t.Error("ComparePassword should fail with wrong password")
    }
}

func TestPasswordStrength(t *testing.T) {
    tests := []struct {
        name     string
        password string
        wantErr  bool
    }{
        {"Valid password", "SecurePass123!", false},
        {"Too short", "Pass1!", true},
        {"Missing uppercase", "securepass123!", true},
        {"Missing lowercase", "SECUREPASS123!", true},
        {"Missing digit", "SecurePass!", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := auth.PasswordStrength(tt.password)
            if (err != nil) != tt.wantErr {
                t.Errorf("PasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 配置更新

需要在 `backend/pkg/go.mod` 中添加：

```go
require (
    golang.org/x/crypto v0.25.0 // bcrypt
    github.com/google/uuid v1.6.0
)
```

### Dev Agent Guardrails

1. **密码必须使用 bcrypt 加密，cost factor = 10**
2. **密码必须至少 8 个字符，包含大小写字母和数字**
3. **用户名和邮箱必须唯一**
4. **用户名格式：3-50 字符，只允许字母数字下划线**
5. **响应中不能包含密码字段**
6. **注册端点不需要认证**
7. **必须返回标准的错误响应格式**
8. **必须编写单元测试**

### Dependencies

**Go 依赖**：
```
- golang.org/x/crypto (bcrypt)
- github.com/google/uuid (UUID)
- gorm.io/gorm (ORM)
- github.com/grpc-ecosystem/grpc-gateway/v2 (gRPC-Gateway)
- google.golang.org/protobuf (Protobuf)
```

### References

- [Source: docs/plans/2025-02-02-aiops-platform-design.md#13-安全与权限-认证]
- [Source: _bmad-output/planning-artifacts/epics.md#Epic-1]
- [Source: _bmad-output/implementation-artifacts/1-1-project-initialization-and-database-setup.md]

## Dev Agent Record

### Agent Model Used

glm-4.7 (claude-opus-4-5-20251101)

### Debug Log References

None (story preparation phase)

### Completion Notes List

**Story Created (2026-02-03)**:

Story prepared for development with detailed implementation guidance.
