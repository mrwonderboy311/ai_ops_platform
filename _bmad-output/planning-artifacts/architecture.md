---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - docs/plans/2025-02-02-aiops-platform-design.md
  - _bmad-output/brainstorming/brainstorming-session-2026-02-02.md
workflowType: 'architecture'
lastStep: 8
status: 'complete'
completedAt: '2026-02-02'
project_name: 'myops-k8s-platform'
user_name: 'Wangjialin'
date: '2026-02-02'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**

基于 PRD 和设计文档，AIOps 平台包含四大核心功能模块：

1. **主机管理服务 (host-svc)**
   - 主机发现（SSH 被动扫描 + Agent 主动上报）
   - 自动注册与资产台账（硬件信息、操作系统、网络配置、标签分组）
   - 运维操作（SSH 终端、文件传输、进程管理、服务控制）
   - 批量任务（并发控制、任务编排、执行历史）

2. **Kubernetes 管理服务 (k8s-svc)**
   - 多集群接入（Kubeconfig 管理、Service Account）
   - CRD 资源管理（Workload、Service、Ingress、ConfigMap、Secret）
   - Pod 交互（日志流、终端执行，通过 WebSocket 代理）
   - Helm 管理（Chart 仓库、应用安装/升级/回滚）

3. **可观测性服务 (obs-svc)**
   - 数据采集（OpenTelemetry Collector 接入）
   - 统一查询 API（指标/日志/链路关联查询）
   - 数据路由（对接 Prometheus/Loki/Tempo）

4. **AI 分析服务 (ai-svc)**
   - 异常检测（时序预测、日志异常模式识别）
   - 智能告警（告警聚合、降噪、根因关联）
   - LLM 集成（自然语言查询、智能诊断助手）

**Non-Functional Requirements:**

| 类别 | 要求 | 架构影响 |
|-----|------|---------|
| **性能** | API P95 <200ms, 日志查询 <2s | 需要高效的数据存储和查询优化 |
| **并发** | 支持 500+ 并发用户 | 需要水平扩展能力、连接池管理 |
| **可用性** | >99.5% | 需要高可用架构、故障恢复机制 |
| **实时性** | WebSocket 用于多场景 | 需要稳定的实时通信架构 |
| **安全** | TLS 1.3, mTLS, RBAC, 审计日志 | 需要完整的安全架构 |
| **可扩展性** | 服务水平扩展、读写分离 | 需要无状态服务设计 |

**Scale & Complexity:**

- 目标用户规模：中型企业（50-500 用户）
- 目标资源规模：1000-10000 台主机
- 主要技术领域：全栈分布式系统
- 复杂度级别：中-高
- 预估架构组件数量：15+（4 个核心微服务 + API Gateway + 6 个依赖服务 + 前端）

### Technical Constraints & Dependencies

**技术选型约束：**
- 后端：Go + gRPC（内部通信）
- 前端：React 18 + TypeScript
- 实时通信：WebSocket
- 部署：Kubernetes 容器化

**外部依赖：**
- PostgreSQL（业务数据）
- Prometheus（指标存储）
- Loki（日志存储）
- Tempo（链路追踪）
- Redis（缓存）
- Kafka（消息队列）
- LLM API（AI 分析）

### Cross-Cutting Concerns Identified

1. **实时通信架构** - 用于主机状态、Pod 日志流、终端、告警推送
2. **安全与权限** - RBAC 模型、审计日志、数据加密、多种认证方式
3. **可观测性集成** - OpenTelemetry 采集、多数据源关联查询
4. **AI/ML 集成** - 异常检测模型、LLM API 调用、RAG 检索
5. **高可用与扩展** - 服务无状态化、数据持久化、自动故障恢复

## Starter Template Evaluation

### Primary Technology Domain

基于项目需求分析，这是一个 **全栈分布式微服务架构** 项目：
- **后端**: Go + gRPC 微服务（4 个独立服务 + API Gateway）
- **前端**: React 18 + TypeScript 管理后台
- **实时通信**: WebSocket（终端、日志流、告警推送）

### Starter Options Considered

#### 后端评估

| 选项 | 评价 |
|-----|------|
| golang-standards/project-layout | ⭐ 行业标准，清晰分层 |
| gmicro 框架 | ⚠️ 功能封装过多，限制灵活性 |
| 自定义微服务结构 | ⭐⭐ 专注微服务场景 |

#### 前端评估

| 选项 | 评价 |
|-----|------|
| Vite 官方 + Ant Design 手动配置 | ⭐⭐ 最新技术，无包袱 |
| react-admin-dashboard | ⭐ 匹配完整技术栈 |
| Ant Design 官方脚手架 | ⚠️ 可能使用过时工具 |

### Selected Starter: 组合方案

**选择理由**：
- 后端使用标准 Go 项目布局，适配微服务场景
- 前端使用 Vite + React + TypeScript，手动集成 Ant Design
- 灵活性高，完全掌控技术栈演进

---

#### Backend: Go + gRPC 微服务

**Initialization Commands:**

```bash
# 创建 Monorepo 结构
mkdir -p backend/{pkg,proto,host-svc,k8s-svc,obs-svc,ai-svc,api-gateway}

# 为每个服务初始化 Go 模块
cd backend/host-svc
go mod init github.com/wangjialin/myops-k8s-platform/backend/host-svc

# 安装核心依赖
go get google.golang.org/grpc
go get google.golang.org/protobuf
go get go.uber.org/zap
go get github.com/prometheus/client_golang
```

**Architectural Decisions Provided:**

| 类别 | 决策 |
|-----|------|
| **语言与运行时** | Go 1.23+, gRPC-Go, Protobuf v3 |
| **项目结构** | 标准项目布局（微服务适配） |
| **分层架构** | Repository → Service → gRPC Handler |
| **日志** | zap 结构化日志 |
| **追踪** | OpenTelemetry Go |
| **配置管理** | Viper 或环境变量 |
| **依赖注入** | wire 或 fx |

**Project Structure Pattern:**

```
backend/
├── pkg/                    # 共享代码
│   ├── auth/              # 认证工具
│   ├── db/                # 数据库
│   ├── kafka/             # 消息队列
│   └── proto/             # Protobuf 定义
├── proto/                 # 共享 .proto 文件
├── api-gateway/           # API 网关
├── host-svc/              # 主机管理服务
├── k8s-svc/               # K8s 管理服务
├── obs-svc/               # 可观测性服务
└── ai-svc/                # AI 分析服务
```

---

#### Frontend: React + TypeScript + Vite + Ant Design

**Initialization Commands:**

```bash
# 创建 Vite React TypeScript 项目
npm create vite@latest frontend -- --template react-ts

# 进入项目目录
cd frontend

# 安装核心依赖
npm install antd @ant-design/icons
npm install react-router-dom zustand @tanstack/react-query axios

# 安装 UI 组件
npm install xterm xterm-addon-fit
npm install @monaco-editor/react

# 安装图表库
npm install echarts echarts-for-react

# 开发依赖
npm install -D @types/node
```

**Architectural Decisions Provided:**

| 类别 | 决策 |
|-----|------|
| **语言与运行时** | TypeScript 5.x, React 18 |
| **构建工具** | Vite 6.x |
| **UI 组件库** | Ant Design 5.x |
| **状态管理** | Zustand |
| **数据请求** | React Query + Axios |
| **路由** | React Router v6 |
| **样式** | Ant Design CSS-in-JS + 可选 Tailwind CSS |
| **代码规范** | ESLint + Prettier |

**Project Structure Pattern:**

```
frontend/
├── src/
│   ├── pages/            # 页面组件
│   ├── components/       # 通用组件
│   ├── hooks/            # 自定义 Hooks
│   ├── api/              # API 调用
│   ├── store/            # 状态管理
│   ├── types/            # TypeScript 类型
│   ├── utils/            # 工具函数
│   └── main.tsx
├── public/
├── package.json
└── vite.config.ts
```

---

### Note

项目初始化应该作为第一个实施故事（Story）来完成。建议按以下顺序：
1. Phase 1: 基础框架 - 先创建后端共享包（pkg）和前端基础项目
2. 然后逐个服务实现

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
- 数据库迁移工具选择
- 认证架构（JWT + Redis）
- API Gateway 实现（gRPC-Gateway + 中间件）

**Important Decisions (Shape Architecture):**
- 错误处理标准（统一错误码和格式）
- WebSocket 架构（Gateway 代理模式）
- 日志策略（zap 结构化日志）

**Deferred Decisions (Post-MVP):**
- CI/CD 具体流程（可使用 GitHub Actions）
- 监控告警规则（运行时调整）

---

### Data Architecture

| 决策 | 选择 | 版本 | 理由 |
|-----|------|------|------|
| **数据库迁移工具** | golang-migrate | latest | 成熟稳定，CLI 友好，CI/CD 易集成 |
| **ORM** | GORM | v1.25+ | 功能完善，社区活跃，支持多数据库 |
| **连接池** | pgx | v5+ | 高性能 PostgreSQL 驱动 |

**迁移文件组织：**
```
backend/{service-name}/migrations/
├── 000001_init_schema.up.sql
├── 000001_init_schema.down.sql
├── 000002_add_hosts_table.up.sql
└── 000002_add_hosts_table.down.sql
```

---

### Authentication & Security

| 决策 | 选择 | 理由 | 影响组件 |
|-----|------|------|---------|
| **Token 类型** | JWT | 无状态，易于水平扩展 | API Gateway, 所有服务 |
| **Token 存储** | Redis | 支持撤销和黑名单 | API Gateway, 认证服务 |
| **认证验证位置** | API Gateway | 统一验证，减少后端负担 | API Gateway |
| **密码加密** | bcrypt | 行业标准，安全性高 | 认证服务 |
| **Agent 认证** | API Token | 支持轮换机制 | Agent, API Gateway |

**认证流程：**
```
用户 → API Gateway → JWT 验证 → gRPC 调用后端服务
Agent → API Gateway → Token 验证 → gRPC 调用后端服务
```

**支持的认证方式（Phase 6 实现）：**
- 本地账号（用户名/密码）
- LDAP/AD（企业级 SSO）
- OAuth 2.0 / SAML（第三方 IdP）

---

### API & Communication Patterns

| 决策 | 选择 | 理由 |
|-----|------|------|
| **API Gateway** | gRPC-Gateway + 自定义中间件 | 自动生成 REST，完全控制 |
| **内部通信** | gRPC | 高性能，类型安全 |
| **外部通信** | REST (JSON) | 前端友好，易于调试 |
| **WebSocket** | Gateway 代理模式 | 统一认证，便于管理 |

**错误响应格式：**
```json
{
  "code": "HOST_NOT_FOUND",
  "message": "主机不存在",
  "details": {"host_id": "123"},
  "request_id": "req-abc123",
  "timestamp": "2026-02-02T10:00:00Z"
}
```

**错误码分类：**
- `AUTH_*` - 认证授权错误
- `HOST_*` - 主机相关错误
- `K8S_*` - K8s 相关错误
- `OBS_*` - 可观测性错误
- `AI_*` - AI 服务错误

---

### Frontend Architecture

| 决策 | 选择 | 理由 |
|-----|------|------|
| **状态管理** | Zustand | 轻量级，TypeScript 友好 |
| **数据请求** | React Query | 缓存、重试、自动刷新 |
| **HTTP 客户端** | Axios | 拦截器支持，成熟稳定 |
| **路由** | React Router v6 | 最新版本，嵌套路由支持 |

**组件模式：**
- 页面组件：功能模块级别
- 业务组件：可复用的业务逻辑
- 基础组件：通用 UI 组件

---

### Infrastructure & Deployment

| 决策 | 选择 | 理由 |
|-----|------|------|
| **容器化** | Docker | 标准化部署 |
| **编排** | Kubernetes | 微服务标准 |
| **Helm Charts** | 自定义 Charts | 灵活控制部署 |
| **CI/CD** | GitHub Actions (建议) | 与代码仓库集成 |

**服务部署模式：**
- 无状态服务：Deployment + HPA
- 有状态服务：StatefulSet（PostgreSQL, Redis, Kafka）

---

### Logging & Observability

**结构化日志标准（使用 zap）：**
```go
{
  "level": "info",
  "service": "host-svc",
  "user_id": "u123",
  "request_id": "req-abc123",
  "trace_id": "trace-xyz",
  "message": "主机注册成功",
  "timestamp": "2026-02-02T10:00:00Z"
}
```

**必填字段：**
- `level` - 日志级别
- `service` - 服务名称
- `message` - 日志消息
- `timestamp` - 时间戳

**可选字段：**
- `user_id` - 用户 ID
- `request_id` - 请求 ID（用于链路追踪）
- `trace_id` - 分布式追踪 ID
- `error` - 错误信息

---

### Decision Impact Analysis

**Implementation Sequence:**

1. **Phase 1 - 基础框架**
   - 初始化 Go Monorepo 和前端 Vite 项目
   - 设置共享包（pkg/auth, pkg/db, pkg/kafka）
   - 配置 gRPC-Gateway 基础中间件

2. **Phase 2 - 主机管理**
   - 实现 host-svc 数据库迁移（golang-migrate）
   - 实现 JWT 认证中间件
   - 实现基础错误处理

3. **Phase 3-6 - 其他服务**
   - 每个服务复用认证和错误处理模式
   - 统一日志格式和追踪

**Cross-Component Dependencies:**

```
┌─────────────────────────────────────────────────────────────┐
│                         API Gateway                          │
│              (认证、限流、路由、协议转换、错误处理)            │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   host-svc   │    │   k8s-svc    │    │   obs-svc    │
│  ┌────────┐  │    │              │    │              │
│  │  GORM  │  │    │  ┌────────┐  │    │  ┌────────┐  │
│  │ pgx    │  │    │  │client-go│  │    │  │ OTel   │  │
│  └────────┘  │    │  └────────┘  │    │  └────────┘  │
└──────────────┘    └──────────────┘    └──────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              ▼
                    ┌──────────────┐
                    │ PostgreSQL   │
                    │ Redis        │
                    │ Kafka        │
                    └──────────────┘

共享依赖:
- pkg/auth (JWT 工具)
- pkg/db (数据库连接池)
- pkg/errors (错误定义)
- pkg/middleware (gRPC 中间件)
```

## Implementation Patterns & Consistency Rules

### Pattern Categories Defined

**Critical Conflict Points Identified:**
- 25+ areas where AI agents could make different choices
- 覆盖命名、结构、格式、通信和流程模式

### Naming Patterns

**Database Naming Conventions:**

| 类型 | 规范 | 示例 |
|-----|------|------|
| **表名** | `snake_case`, 复数 | `hosts`, `users`, `k8s_clusters` |
| **列名** | `snake_case` | `host_id`, `host_name`, `created_at` |
| **主键** | `id` | `id` |
| **外键** | `{ref_table}_id` | `user_id`, `cluster_id` |
| **索引** | `idx_{table}_{columns}` | `idx_hosts_status`, `idx_users_email` |
| **唯一约束** | `uniq_{table}_{columns}` | `uniq_hosts_hostname` |

```go
// Go Struct 定义示例
type Host struct {
    ID          string    `json:"id" gorm:"primaryKey;column:id"`
    HostName    string    `json:"hostName" gorm:"column:host_name;not null"`
    Status      string    `json:"status" gorm:"column:status;index:idx_hosts_status"`
    ClusterID   *string   `json:"clusterId" gorm:"column:cluster_id"`
    CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
    UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

// PostgreSQL 表: hosts
// Columns: id, host_name, status, cluster_id, created_at, updated_at
// Indexes: idx_hosts_status, uniq_hosts_hostname
```

**API Naming Conventions:**

| 类型 | 规范 | 示例 |
|-----|------|------|
| **REST 端点** | `/api/v1/{resource_plural}` | `/api/v1/hosts`, `/api/v1/pods` |
| **路由参数** | `:param_name` | `/api/v1/hosts/:host_id` |
| **查询参数** | `snake_case` | `?host_id=123&status=online` |
| **请求头** | `X-Custom-Header` | `X-Request-ID`, `X-Auth-Token` |

```typescript
// 前端 API 调用示例
const response = await axios.get('/api/v1/hosts/:host_id', {
  params: { host_id: '123' },
  headers: { 'X-Request-ID': generateRequestId() }
});
```

**Code Naming Conventions:**

**Go 代码：**
- 包名：`lowercase`（如 `hostsvc`, `k8ssvc`, `auth`）
- 文件名：`snake_case`（如 `host_service.go`, `repository.go`, `middleware.go`）
- 导出函数/变量：`PascalCase`（如 `GetHost`, `CreateHost`, `HostService`）
- 私有函数/变量：`camelCase`（如 `validateInput`, `parseConfig`）
- 常量：`PascalCase` 或 `UPPER_SNAKE_CASE`（如 `MaxRetries` 或 `MAX_RETRIES`）
- 接口：`PascalCase` + `er` 后缀（如 `HostRepository`, `Logger`）

**React/TypeScript 代码：**
- 组件：`PascalCase`（如 `HostList`, `HostDetail`, `PodLogs`）
- 文件名：与组件名一致（如 `HostList.tsx`, `PodLogs.tsx`）
- Hook：`camelCase`，`use` 前缀（如 `useHosts`, `useAuth`, `useWebSocket`）
- 工具函数：`camelCase`（如 `formatDate`, `truncate`, `parseError`）
- 类型/接口：`PascalCase`（如 `Host`, `ApiResponse`, `WebSocketMessage`）
- 枚举：`PascalCase`（如 `HostStatus`, `LogLevel`）

```typescript
// React 组件示例
// 文件: HostList.tsx
interface HostListProps {
  clusterId?: string;
}

export function HostList({ clusterId }: HostListProps) {
  const { data, isLoading } = useHosts({ clusterId });
  // ...
}

// 自定义 Hook
// 文件: useHosts.ts
export function useHosts(params: HostQueryParams) {
  return useQuery({
    queryKey: ['hosts', params],
    queryFn: () => fetchHosts(params),
  });
}
```

### Structure Patterns

**Project Organization:**

**Go 后端：**
```
backend/{service-name}/
├── cmd/
│   └── server/
│       └── main.go              # 服务入口
├── internal/
│   ├── config/                  # 配置管理
│   ├── grpc/                    # gRPC 服务实现
│   │   ├── host_service.go      # 服务实现
│   │   └── host_service_test.go # 测试（同目录）
│   ├── repository/              # 数据访问层
│   │   ├── host_repository.go
│   │   └── host_repository_test.go
│   ├── service/                 # 业务逻辑层
│   │   └── host_service.go
│   └── middleware/              # 中间件
│       └── auth.go
├── migrations/                  # 数据库迁移
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── go.mod
└── Dockerfile
```

**React 前端：**
```
frontend/src/
├── pages/                       # 页面组件（按功能组织）
│   ├── HostList/
│   │   ├── index.tsx            # 页面组件
│   │   ├── components/          # 页面私有组件
│   │   │   └── HostCard.tsx
│   │   ├── hooks/               # 页面私有 hooks
│   │   │   └── useHostFilters.ts
│   │   └── api.ts               # 页面 API 调用
│   └── Dashboard/
│       └── index.tsx
├── components/                  # 通用组件
│   ├── ui/                      # 基础 UI 组件
│   │   ├── Button.tsx
│   │   └── Input.tsx
│   └── business/                # 业务组件
│       └── StatusBadge.tsx
├── hooks/                       # 通用 hooks
│   ├── useAuth.ts
│   ├── useWebSocket.ts
│   └── useRequest.ts
├── api/                         # API 层
│   ├── hosts.ts
│   ├── pods.ts
│   └── client.ts                # Axios 实例配置
├── store/                       # 状态管理
│   ├── auth.ts
│   └── global.ts
├── types/                       # TypeScript 类型
│   ├── host.ts
│   ├── k8s.ts
│   └── api.ts
├── utils/                       # 工具函数
│   ├── format.ts
│   └── validation.ts
├── App.tsx
└── main.tsx
```

**测试文件位置：**
- **Go**: 与源文件同目录，`{filename}_test.go`
- **React**: 可选同目录 `*.test.tsx` 或 `tests/` 目录

### Format Patterns

**API Response Formats:**

**成功响应：**
```json
{
  "data": {
    "id": "123",
    "hostName": "web-01",
    "status": "online"
  },
  "requestId": "req-abc123"
}
```

**错误响应：**
```json
{
  "error": {
    "code": "HOST_NOT_FOUND",
    "message": "主机不存在",
    "details": {
      "hostId": "123"
    }
  },
  "requestId": "req-abc123"
}
```

**分页响应：**
```json
{
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "pageSize": 20
  },
  "requestId": "req-abc123"
}
```

**Data Exchange Formats:**

| 场景 | 格式 |
|-----|------|
| **JSON 字段名（前端）** | `camelCase` |
| **JSON 字段名（后端）** | `camelCase`（通过 json tag） |
| **数据库列名** | `snake_case` |
| **日期时间** | ISO 8601 字符串 |
| **布尔值** | `true` / `false` |
| **空值** | `null` |

### Communication Patterns

**Event System Patterns (Kafka):**

**事件命名：**
- 格式：`{service}.{entity}.{action}`
- 示例：`host.host.registered`, `host.host.updated`, `k8s.pod.scaled`

**事件结构：**
```json
{
  "eventType": "host.host.registered",
  "eventVersion": "1.0",
  "timestamp": "2026-02-02T10:00:00Z",
  "data": {
    "hostId": "123",
    "hostName": "web-01",
    "clusterId": "c1"
  }
}
```

**State Management Patterns (Zustand):**

```typescript
// store/auth.ts
interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (credentials: Credentials) => Promise<void>;
  logout: () => void;
}

// 使用不可变更新
const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  isAuthenticated: false,
  login: async (credentials) => {
    const { user, token } = await api.login(credentials);
    set({ user, token, isAuthenticated: true });
  },
  logout: () => set({ user: null, token: null, isAuthenticated: false }),
}));
```

### Process Patterns

**Error Handling Patterns:**

**后端 (Go)：**
```go
// pkg/errors/errors.go
type AppError struct {
    Code    string
    Message string
    Details map[string]interface{}
}

func (e *AppError) Error() string {
    return e.Message
}

// 使用示例
if host == nil {
    return &AppError{
        Code:    "HOST_NOT_FOUND",
        Message: "主机不存在",
        Details: map[string]interface{}{"host_id": id},
    }
}
```

**前端 (React)：**
```typescript
// 全局错误边界
// components/ErrorBoundary.tsx
class ErrorBoundary extends React.Component {
  // ...

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught:', error, errorInfo);
    // 发送到错误追踪服务
  }
}

// API 错误处理
// api/client.ts
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    const { error: err } = error.response.data;
    // 显示通知
    notification.error({
      message: err.code,
      description: err.message,
    });
    return Promise.reject(error);
  }
);
```

**Loading State Patterns:**

```typescript
// 使用 React Query 自动管理加载状态
const { data, isLoading, error } = useHosts();

// 或使用自定义 Hook
function useHosts(params: HostQueryParams) {
  const [isLoading, setIsLoading] = useState(false);
  const [data, setData] = useState<Host[]>([]);

  useEffect(() => {
    setIsLoading(true);
    fetchHosts(params).finally(() => setIsLoading(false));
  }, [params]);

  return { data, isLoading };
}
```

### Enforcement Guidelines

**所有 AI 代理必须遵守：**

1. **命名规范**
   - 数据库：`snake_case`
   - JSON API：`camelCase`
   - Go 代码：遵循 Go 官方规范
   - React 代码：遵循 React 社区规范

2. **文件组织**
   - Go 测试与源文件同目录
   - React 页面按功能组织
   - 共享代码放在共享目录

3. **API 格式**
   - 成功：`{ data: ..., requestId: ... }`
   - 错误：`{ error: { code, message, details }, requestId: ... }`

4. **日志格式**
   - 使用 zap 结构化日志
   - 必填：`level`, `service`, `message`, `timestamp`
   - 可选：`user_id`, `request_id`, `trace_id`

5. **错误处理**
   - 使用统一的错误类型（AppError）
   - 所有 API 响应包含 `requestId`
   - 前端使用错误边界捕获组件错误

### Pattern Examples

**Good Examples:**

```go
// ✅ 正确：Go 服务层
type HostService struct {
    repo HostRepository
    log  *zap.Logger
}

func (s *HostService) GetHost(ctx context.Context, id string) (*Host, error) {
    host, err := s.repo.FindByID(ctx, id)
    if err != nil {
        s.log.Error("failed to find host", zap.String("host_id", id), zap.Error(err))
        return nil, &AppError{
            Code:    "HOST_NOT_FOUND",
            Message: "主机不存在",
            Details: map[string]interface{}{"host_id": id},
        }
    }
    return host, nil
}
```

```typescript
// ✅ 正确：React 组件
export function HostList({ clusterId }: HostListProps) {
  const { data, isLoading, error } = useHosts({ clusterId });

  if (isLoading) return <Spin />;
  if (error) return <Alert type="error" message={error.message} />;

  return (
    <Table dataSource={data} columns={columns} />
  );
}
```

**Anti-Patterns:**

```go
// ❌ 错误：混合命名风格
type host_service struct {  // 应该是 HostService
    Host_ID string          // 应该是 HostID
}
```

```typescript
// ❌ 错误：不一致的 API 响应格式
// 有时返回 { data: ... }，有时直接返回数据
// 有时包含 requestId，有时不包含
```

### Enforcement Process

1. **代码审查检查清单**
   - [ ] 命名规范遵守
   - [ ] 文件组织正确
   - [ ] API 格式统一
   - [ ] 错误处理一致
   - [ ] 日志格式正确

2. **自动化检查**
   - Go: `golangci-lint`
   - React: `eslint`, `typescript`

3. **模式更新流程**
   - 在架构文档中记录变更
   - 通知所有开发人员
   - 更新现有代码（如必要）

## Project Structure & Boundaries

### Complete Project Directory Structure

```
myops-k8s-platform/
├── README.md
├── .gitignore
├── docker-compose.yml
├── .github/
│   └── workflows/
│       └── ci.yml
│
├── backend/                      # Go 后端 Monorepo
│   ├── go.work                   # Go workspace 文件
│   ├── proto/                    # 共享 Protobuf 定义
│   │   ├── host/
│   │   │   └── host.proto
│   │   ├── k8s/
│   │   │   └── k8s.proto
│   │   ├── obs/
│   │   │   └── obs.proto
│   │   ├── ai/
│   │   │   └── ai.proto
│   │   └── auth/
│   │       └── auth.proto
│   │
│   ├── pkg/                      # 共享代码包
│   │   ├── auth/
│   │   │   ├── jwt.go           # JWT 工具
│   │   │   ├── middleware.go    # gRPC 认证中间件
│   │   │   └── password.go      # 密码加密
│   │   ├── db/
│   │   │   ├── postgres.go      # PostgreSQL 连接池
│   │   │   ├── redis.go         # Redis 连接
│   │   │   └── migrations.go    # 迁移工具封装
│   │   ├── kafka/
│   │   │   ├── producer.go      # Kafka 生产者
│   │   │   ├── consumer.go      # Kafka 消费者
│   │   │   └── events.go        # 事件定义
│   │   ├── errors/
│   │   │   └── app_error.go     # 统一错误类型
│   │   ├── logger/
│   │   │   └── logger.go        # zap 日志配置
│   │   └── middleware/
│   │       ├── logging.go       # 日志中间件
│   │       ├── tracing.go       # 追踪中间件
│   │       └── recovery.go      # 恢复中间件
│   │
│   ├── api-gateway/             # API Gateway
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── gateway/
│   │   │   │   ├── gateway.go   # gRPC-Gateway 服务器
│   │   │   │   ├── middleware.go # 认证、限流中间件
│   │   │   │   └── proxy.go      # gRPC 代理
│   │   │   └── handler/
│   │   │       └── websocket.go # WebSocket 处理
│   │   ├── migrations/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── host-svc/               # 主机管理服务
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── grpc/
│   │   │   │   ├── host_service.go
│   │   │   │   └── host_service_test.go
│   │   │   ├── repository/
│   │   │   │   ├── host_repository.go
│   │   │   │   ├── host_repository_test.go
│   │   │   │   ├── agent_repository.go
│   │   │   │   └── task_repository.go
│   │   │   ├── service/
│   │   │   │   ├── host_service.go
│   │   │   │   ├── agent_service.go
│   │   │   │   └── task_service.go
│   │   │   ├── ssh/
│   │   │   │   ├── proxy.go       # SSH 代理
│   │   │   │   └── terminal.go    # 终端管理
│   │   │   └── model/
│   │   │       ├── host.go
│   │   │       ├── agent.go
│   │   │       └── task.go
│   │   ├── migrations/
│   │   │   ├── 000001_init_schema.up.sql
│   │   │   ├── 000001_init_schema.down.sql
│   │   │   ├── 000002_create_hosts_table.up.sql
│   │   │   └── 000002_create_hosts_table.down.sql
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── k8s-svc/                # K8s 管理服务
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── grpc/
│   │   │   │   ├── cluster_service.go
│   │   │   │   ├── pod_service.go
│   │   │   │   ├── helm_service.go
│   │   │   │   └── *_test.go
│   │   │   ├── repository/
│   │   │   │   ├── cluster_repository.go
│   │   │   │   └── pod_repository.go
│   │   │   ├── service/
│   │   │   │   ├── cluster_service.go
│   │   │   │   ├── k8s_client.go   # K8s 客户端封装
│   │   │   │   ├── helm_service.go
│   │   │   │   └── websocket.go    # Pod 日志/终端 WebSocket
│   │   │   └── model/
│   │   │       ├── cluster.go
│   │   │       ├── pod.go
│   │   │       └── helm_release.go
│   │   ├── migrations/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── obs-svc/                # 可观测性服务
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── grpc/
│   │   │   │   ├── query_service.go
│   │   │   │   └── collector_service.go
│   │   │   ├── service/
│   │   │   │   ├── prometheus.go   # Prometheus 查询
│   │   │   │   ├── loki.go         # Loki 查询
│   │   │   │   ├── tempo.go        # Tempo 查询
│   │   │   │   └── otel.go         # OpenTelemetry 接入
│   │   │   └── model/
│   │   │       ├── metric.go
│   │   │       ├── log.go
│   │   │       └── trace.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── ai-svc/                 # AI 分析服务
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── grpc/
│   │   │   │   ├── anomaly_service.go
│   │   │   │   ├── alert_service.go
│   │   │   │   └── llm_service.go
│   │   │   ├── service/
│   │   │   │   ├── anomaly.go      # 异常检测
│   │   │   │   ├── alert.go        # 告警聚合
│   │   │   │   ├── llm.go          # LLM 集成
│   │   │   │   └── rag.go          # RAG 检索
│   │   │   └── model/
│   │   │       ├── anomaly.go
│   │   │       ├── alert.go
│   │   │       └── llm_prompt.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   └── scripts/                # 部署脚本
│       ├── build.sh
│       ├── deploy.sh
│       └── migrate.sh
│
├── frontend/                   # React 前端
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── .env.local
│   ├── .env.example
│   ├── index.html
│   │
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       │
│       ├── pages/               # 页面组件
│       │   ├── Dashboard/
│       │   │   ├── index.tsx
│       │   │   ├── components/
│       │   │   └── hooks/
│       │   │
│       │   ├── HostList/
│       │   │   ├── index.tsx
│       │   │   ├── components/
│       │   │   │   ├── HostTable.tsx
│       │   │   │   ├── HostCard.tsx
│       │   │   │   └── HostFilter.tsx
│       │   │   ├── hooks/
│       │   │   │   └── useHostFilters.ts
│       │   │   └── api.ts
│       │   │
│       │   ├── HostDetail/
│       │   │   ├── index.tsx
│       │   │   ├── components/
│       │   │   │   ├── InfoCard.tsx
│       │   │   │   ├── MetricsChart.tsx
│       │   │   │   └── ProcessList.tsx
│       │   │   └── tabs/
│       │   │       ├── Terminal.tsx
│       │   │       └── FileManager.tsx
│       │   │
│       │   ├── ClusterList/
│       │   │   ├── index.tsx
│       │   │   ├── components/
│       │   │   └── api.ts
│       │   │
│       │   ├── PodList/
│       │   │   ├── index.tsx
│       │   │   ├── components/
│       │   │   │   ├── PodTable.tsx
│       │   │   │   └── PodDetail.tsx
│       │   │   ├── tabs/
│       │   │   │   ├── Logs.tsx
│       │   │   │   └── Terminal.tsx
│       │   │   └── api.ts
│       │   │
│       │   ├── Observability/
│       │   │   ├── Metrics/
│       │   │   ├── Logs/
│       │   │   └── Traces/
│       │   │
│       │   └── AIAssistant/
│       │       ├── index.tsx
│       │       ├── components/
│       │       │   └── ChatPanel.tsx
│       │       └── hooks/
│       │           └── useChat.ts
│       │
│       ├── components/          # 通用组件
│       │   ├── ui/
│       │   │   ├── Button.tsx
│       │   │   ├── Input.tsx
│       │   │   └── Modal.tsx
│       │   └── business/
│       │       ├── StatusBadge.tsx
│       │       └── ResourceCard.tsx
│       │
│       ├── hooks/               # 通用 hooks
│       │   ├── useAuth.ts
│       │   ├── useWebSocket.ts
│       │   └── useRequest.ts
│       │
│       ├── api/                 # API 层
│       │   ├── client.ts        # Axios 实例配置
│       │   ├── hosts.ts
│       │   ├── pods.ts
│       │   ├── clusters.ts
│       │   └── observability.ts
│       │
│       ├── store/               # 状态管理
│       │   ├── auth.ts
│       │   ├── global.ts
│       │   └── websocket.ts
│       │
│       ├── types/               # TypeScript 类型
│       │   ├── host.ts
│       │   ├── k8s.ts
│       │   ├── observability.ts
│       │   └── api.ts
│       │
│       ├── utils/               # 工具函数
│       │   ├── format.ts
│       │   ├── validation.ts
│       │   └── constants.ts
│       │
│       └── styles/
│           └── globals.css
│
├── agent/                      # 主机 Agent
│   ├── cmd/
│   │   └── agent/
│   │       └── main.go
│   ├── internal/
│   │   ├── collector/
│   │   │   ├── system.go       # 系统信息采集
│   │   │   ├── process.go      # 进程信息采集
│   │   │   └── network.go      # 网络信息采集
│   │   ├── sshd/
│   │   │   └── server.go       # SSH 服务器
│   │   └── upstream/
│   │       └── reporter.go     # 上报通信
│   ├── go.mod
│   └── Dockerfile
│
├── deploy/                     # K8s 部署配置
│   ├── helm/
│   │   ├── myops/
│   │   │   ├── Chart.yaml
│   │   │   ├── values.yaml
│   │   │   └── templates/
│   │   │       ├── gateway/
│   │   │       ├── host-svc/
│   │   │       ├── k8s-svc/
│   │   │       ├── obs-svc/
│   │   │       ├── ai-svc/
│   │   │       └── frontend/
│   │   └── dependency-charts/
│   │       ├── postgresql/
│   │       ├── redis/
│   │       ├── kafka/
│   │       └── prometheus/
│   └── k8s/
│       ├── base/
│       └── overlays/
│
└── docs/                       # 文档
    ├── plans/
    │   └── 2025-02-02-aiops-platform-design.md
    ├── api/
    │   ├── openapi.yaml
    │   └── grpc.md
    └── user-guide/
        └── getting-started.md
```

### Architectural Boundaries

**API Boundaries:**

| 边界 | 端点 | 协议 |
|-----|------|------|
| **外部 API** | `/api/v1/*` | REST (JSON) |
| **内部 gRPC** | `:{service_port}` | gRPC (Protobuf) |
| **WebSocket** | `/ws/*` | WebSocket |
| **Agent 上报** | `/agent/v1/report` | HTTP/gRPC |

**Component Boundaries:**

```
┌─────────────────────────────────────────────────────────────┐
│                       Frontend (React)                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Pages       │  │ Components  │  │ Hooks (useWebSocket) │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────────────┘ │
│         │                │                  │               │
│         └────────────────┴──────────────────┘               │
│                           │                                  │
└───────────────────────────┼──────────────────────────────────┘
                            │
                    ┌───────▼────────┐
                    │  API Gateway   │
                    │  (认证 + 路由)  │
                    └───────┬────────┘
                            │ gRPC
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼────────┐ ┌────────▼────────┐ ┌───────▼────────┐
│   host-svc     │ │    k8s-svc      │ │    obs-svc     │
│ (主机管理)     │ │  (K8s 管理)     │ │ (可观测性)     │
└────────────────┘ └─────────────────┘ └────────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼────────┐ ┌───────▼────────┐ ┌───────▼────────┐
│  PostgreSQL    │ │     Redis       │ │     Kafka       │
└────────────────┘ └─────────────────┘ └────────────────┘
```

**Service Boundaries:**

- 每个服务独立部署
- 服务间通过 gRPC 通信
- 服务共享 `pkg/` 中的代码
- 每个服务有自己的数据库 schema

**Data Boundaries:**

- 每个服务拥有独立的数据表
- 通过 Kafka 事件同步数据
- Redis 用于共享缓存
- PostgreSQL 按服务分 schema

### Requirements to Structure Mapping

**Feature/Epic Mapping:**

| 功能模块 | 后端位置 | 前端位置 | API |
|---------|---------|---------|-----|
| **主机列表** | `host-svc/internal/grpc/host_service.go` | `frontend/src/pages/HostList/` | `GET /api/v1/hosts` |
| **主机详情** | `host-svc/internal/grpc/host_service.go` | `frontend/src/pages/HostDetail/` | `GET /api/v1/hosts/:id` |
| **SSH 终端** | `host-svc/internal/ssh/terminal.go` | `frontend/src/pages/HostDetail/tabs/Terminal.tsx` | `WS /ws/ssh/:host_id` |
| **集群列表** | `k8s-svc/internal/grpc/cluster_service.go` | `frontend/src/pages/ClusterList/` | `GET /api/v1/clusters` |
| **Pod 列表** | `k8s-svc/internal/grpc/pod_service.go` | `frontend/src/pages/PodList/` | `GET /api/v1/pods` |
| **Pod 日志** | `k8s-svc/internal/service/websocket.go` | `frontend/src/pages/PodList/tabs/Logs.tsx` | `WS /ws/pods/:namespace/:name/logs` |
| **指标查询** | `obs-svc/internal/grpc/query_service.go` | `frontend/src/pages/Observability/Metrics/` | `GET /api/v1/metrics` |
| **AI 助手** | `ai-svc/internal/grpc/llm_service.go` | `frontend/src/pages/AIAssistant/` | `POST /api/v1/ai/chat` |

**Cross-Cutting Concerns:**

| 关注点 | 后端位置 | 前端位置 |
|-------|---------|---------|
| **认证** | `pkg/auth/` | `frontend/src/store/auth.ts` |
| **日志** | `pkg/logger/` | 前端集成 Sentry |
| **错误处理** | `pkg/errors/` | `frontend/src/api/client.ts` |
| **WebSocket** | `api-gateway/internal/handler/websocket.go` | `frontend/src/hooks/useWebSocket.ts` |
| **配置管理** | `{service}/internal/config/` | `frontend/.env.local` |

### Integration Points

**Internal Communication:**

```
Frontend ←→ API Gateway ←→ [host-svc, k8s-svc, obs-svc, ai-svc]
                                    ↓
                    [PostgreSQL, Redis, Kafka]
```

**External Integrations:**

| 系统 | 集成点 | 用途 |
|-----|-------|------|
| **Kubernetes API** | `k8s-svc/internal/service/k8s_client.go` | 管理 K8s 资源 |
| **Prometheus** | `obs-svc/internal/service/prometheus.go` | 查询指标 |
| **Loki** | `obs-svc/internal/service/loki.go` | 查询日志 |
| **Tempo** | `obs-svc/internal/service/tempo.go` | 查询链路 |
| **LLM API** | `ai-svc/internal/service/llm.go` | AI 分析 |

**Data Flow:**

```
Agent → 上报信息 → Kafka → host-svc → PostgreSQL
                        ↓
                   ai-svc → 异常检测 → 告警

用户 → 前端 → API Gateway → 各服务 → PostgreSQL/Redis/K8s API
```

### File Organization Patterns

**Configuration Files:**

- 根目录：`.env.example`, `docker-compose.yml`, `.github/workflows/ci.yml`
- 项目文档：`docs/` 目录

**Source Organization:**

- Go: 每个服务独立 `go.mod`
- React: 单体应用，按页面组织

**Test Organization:**

- Go: `*_test.go` 与源文件同目录
- React: 可选同目录或 `tests/` 目录

**Asset Organization:**

- 静态文件: `frontend/public/`
- 样式: `frontend/src/styles/`
- 图标: `frontend/public/icons/`

## Architecture Validation Results

### Coherence Validation ✅

**Decision Compatibility:**
- ✅ Go 1.23+ 与 gRPC、GORM、zap 版本兼容
- ✅ React 18 + TypeScript 5.x + Vite 6.x + Ant Design 5.x 协同工作
- ✅ JWT + Redis 认证架构支持所有服务
- ✅ gRPC-Gateway 支持 REST 到 gRPC 的协议转换
- ✅ 无矛盾决策，所有技术选择相互支持

**Pattern Consistency:**
- ✅ 命名规范覆盖所有层面（数据库、API、代码）
- ✅ 文件组织模式与技术栈对齐
- ✅ 通信模式（REST/gRPC/WebSocket/Kafka）清晰定义
- ✅ 错误处理模式统一（AppError + 前端 ErrorBoundary）

**Structure Alignment:**
- ✅ 项目结构支持微服务架构
- ✅ 组件边界清晰（4 个微服务 + API Gateway + 前端）
- ✅ 集成点明确（PostgreSQL、Redis、Kafka、K8s API）

### Requirements Coverage Validation ✅

**Epic/Feature Coverage:**

| 功能模块 | 架构支持 | 状态 |
|---------|---------|------|
| **主机管理** | host-svc + Agent | ✅ |
| **K8s 管理** | k8s-svc + client-go | ✅ |
| **可观测性** | obs-svc + OTel + Prometheus/Loki/Tempo | ✅ |
| **AI 分析** | ai-svc + LLM API | ✅ |
| **实时通信** | WebSocket + Gateway 代理 | ✅ |
| **认证授权** | JWT + Redis + RBAC | ✅ |

**Functional Requirements Coverage:**
- ✅ 所有 PRD 功能模块均有架构支持
- ✅ 跨领域关注点（安全、可观测性、AI）已处理

**Non-Functional Requirements Coverage:**

| 需求 | 架构支持 | 状态 |
|-----|---------|------|
| **性能** | 无状态服务 + 连接池 + HPA | ✅ |
| **并发** | 水平扩展 + Redis 缓存 | ✅ |
| **可用性** | 多副本 + 故障恢复 | ✅ |
| **安全** | RBAC + TLS/mTLS + 审计日志 | ✅ |
| **可扩展性** | 微服务 + Kafka 事件 | ✅ |

### Implementation Readiness Validation ✅

**Decision Completeness:**
- ✅ 所有关键决策已记录版本号
- ✅ 实施模式包含具体示例
- ✅ 一致性规则清晰可执行

**Structure Completeness:**
- ✅ 项目结构完整且具体
- ✅ 所有文件和目录已定义
- ✅ 集成点明确指定

**Pattern Completeness:**
- ✅ 25+ 潜在冲突点已处理
- ✅ 命名约定全面
- ✅ 通信模式完整定义

### Gap Analysis Results

**Critical Gaps:** 无

**Important Gaps:** 无

**Nice-to-Have Gaps:**
1. CI/CD 流程的详细配置（GitHub Actions workflow）
2. 监控告警的具体规则配置
3. 性能测试计划

### Architecture Completeness Checklist

**✅ 需求分析**
- [x] 项目上下文已彻底分析
- [x] 规模和复杂度已评估
- [x] 技术约束已识别
- [x] 跨领域关注点已映射

**✅ 架构决策**
- [x] 关键决策已记录版本
- [x] 技术栈已完全指定
- [x] 集成模式已定义
- [x] 性能考虑已处理

**✅ 实施模式**
- [x] 命名约定已建立
- [x] 结构模式已定义
- [x] 通信模式已指定
- [x] 流程模式已记录

**✅ 项目结构**
- [x] 完整目录结构已定义
- [x] 组件边界已建立
- [x] 集成点已映射
- [x] 需求到结构映射已完成

### Architecture Readiness Assessment

**Overall Status:** ✅ 准备实施

**Confidence Level:** 高 - 基于验证结果

**Key Strengths:**
1. 微服务架构支持独立开发和部署
2. 完整的技术栈选择成熟稳定
3. 清晰的模式确保 AI 代理一致性
4. 全面的项目结构便于导航
5. 设计文档与架构决策完全一致

**Areas for Future Enhancement:**
1. CI/CD 自动化流程配置
2. 高级监控告警规则定义
3. 性能测试与优化计划

### Implementation Handoff

**AI Agent Guidelines:**
- 严格按照文档中的架构决策实施
- 一致使用所有实施模式
- 遵循项目结构和边界定义
- 所有架构问题参考此文档

**First Implementation Priority:**

```bash
# 1. 初始化 Go Monorepo 和前端项目
mkdir -p backend/{pkg,proto,host-svc,k8s-svc,obs-svc,ai-svc,api-gateway}
npm create vite@latest frontend -- --template react-ts

# 2. 安装后端核心依赖
cd backend/pkg
go mod init github.com/wangjialin/myops-k8s-platform/backend/pkg
go get google.golang.org/grpc
go get google.golang.org/protobuf
go get go.uber.org/zap
go get github.com/prometheus/client_golang
go get gorm.io/gorm
go get github.com/redis/go-redis/v9
go get github.com/golang-jwt/jwt/v5

# 3. 安装前端依赖
cd frontend
npm install antd @ant-design/icons
npm install react-router-dom zustand @tanstack/react-query axios
```

---

## 🎉 架构工作流完成

**Wangjialin，架构文档已完成！**

**文档位置**: [`_bmad-output/planning-artifacts/architecture.md`](_bmad-output/planning-artifacts/architecture.md)

**包含内容：**
1. ✅ 项目上下文分析
2. ✅ 起始模板评估
3. ✅ 核心架构决策
4. ✅ 实施模式与一致性规则
5. ✅ 项目结构与边界
6. ✅ 架构验证结果

**下一步建议：**
- 开始 Phase 1 实施（基础框架）
- 或创建史诗和用户故事
- 或进行实施就绪性检查

**你可以：**
- 输入 `MH` 重新显示此菜单
- 输入 `CH` 与我讨论任何问题
- 输入 `DA` 解除架构师代理
