---
stepsCompleted: [1, 2, 3, 4]
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - _bmad-output/planning-artifacts/architecture.md
  - docs/plans/2025-02-02-aiops-platform-design.md
workflowType: 'epics-and-stories'
lastStep: 4
status: 'complete'
completedAt: '2026-02-02'
project_name: 'myops-k8s-platform'
user_name: 'Wangjialin'
date: '2026-02-02'
---

# myops-k8s-platform - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for myops-k8s-platform, decomposing the requirements from the PRD, Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

**FR1**: 主机管理 - 主机发现与注册
- 支持 SSH 被动扫描和 Agent 主动上报两种发现方式
- 首次发现自动注册，支持审批流程
- 记录主机硬件信息、操作系统、网络配置
- 支持标签分组管理

**FR2**: 主机管理 - 运维操作
- Web SSH 终端（xterm.js + WebSocket）
- 文件传输和管理
- 进程管理（查看、启动、停止）
- 服务控制

**FR3**: 主机管理 - 批量任务
- 并发控制（限制同时执行数量）
- 任务编排（串行/并行执行）
- 执行历史记录

**FR4**: K8s 管理 - 集群接入
- 多集群管理（Kubeconfig 导入）
- Service Account 认证

**FR5**: K8s 管理 - 资源管理
- CRD 资源 CRUD（Workload、Service、Ingress、ConfigMap、Secret）
- Pod 交互（日志流、终端执行）

**FR6**: K8s 管理 - Helm 应用管理
- Chart 仓库管理
- 应用安装/升级/回滚

**FR7**: 可观测性 - 数据采集
- OpenTelemetry Collector 接入
- Agent 数据上报

**FR8**: 可观测性 - 统一查询
- 指标/日志/链路关联查询 API
- 对接 Prometheus/Loki/Tempo

**FR9**: AI 智能分析 - 异常检测
- 时序指标异常检测（STL + 统计分析）
- 日志异常模式识别（孤立森林/LSTM）
- 基线学习

**FR10**: AI 智能分析 - 智能告警
- 告警聚合（基于时间窗口和相关性）
- 根因关联（拓扑图追溯）
- 降噪规则

**FR11**: AI 智能分析 - LLM 集成
- 自然语言查询（"CPU 超过 80% 的 Pod" → PromQL）
- 智能诊断助手（结合知识库 + 实时数据）

**FR12**: AI 智能分析 - 预测分析
- 容量预测（资源耗尽时间）
- 故障预测（异常模式提前预警）

**FR13**: 安全与权限 - 认证
- 本地账号（用户名/密码，bcrypt 加密）
- LDAP/AD 集成
- OAuth 2.0 / SAML
- API Token（Agent 认证）

**FR14**: 安全与权限 - RBAC
- 用户 → 角色 → 权限 → 资源模型
- 预定义角色（超级管理员、运维人员、只读用户、审计员）
- 资源级权限（限定到特定集群、命名空间、主机组）

**FR15**: 安全与权限 - 审计日志
- 所有操作记录（谁、何时、做了什么、结果）
- 敏感操作二次确认
- 审计日志不可修改

**FR16**: 前端 - 主机管理界面
- 主机列表（搜索、过滤、排序）
- 主机详情（基本信息、监控图表、进程列表、文件管理）

**FR17**: 前端 - K8s 管理界面
- 集群列表
- Workload 管理（Deployment、StatefulSet、DaemonSet）
- Pod 管理（列表、详情、日志、终端）
- Service/Ingress、ConfigMap/Secret
- Helm 应用管理

**FR18**: 前端 - 可观测性界面
- 仪表盘
- 指标查询
- 日志查询（虚拟滚动）
- 链路追踪

**FR19**: 前端 - AI 智能助手
- 自然语言查询输入
- 告警分析面板
- 智能诊断对话界面

### NonFunctional Requirements

**NFR1**: 性能 - API 响应时间 P95 < 200ms
**NFR2**: 性能 - 日志查询延迟 < 2s
**NFR3**: 性能 - 并发用户支持 500+
**NFR4**: 性能 - 数据采集延迟 < 10s
**NFR5**: 可扩展性 - 服务水平扩展
**NFR6**: 可扩展性 - 数据库读写分离
**NFR7**: 可扩展性 - Kafka 分区扩展
**NFR8**: 可靠性 - 服务可用性 > 99.5%
**NFR9**: 可靠性 - 数据持久化保证
**NFR10**: 可靠性 - 自动故障恢复
**NFR11**: 安全 - 全站 HTTPS (TLS 1.3)
**NFR12**: 安全 - 内部 mTLS 通信
**NFR13**: 安全 - Agent 双向认证、Token 轮换
**NFR14**: 安全 - 敏感数据加密
**NFR15**: 安全 - API Rate Limiting、防注入、CSRF/XSS 防护

### Additional Requirements

**技术栈需求:**
- 后端: Go 1.23+、gRPC、Protobuf v3、GORM、zap
- 前端: React 18、TypeScript 5.x、Vite 6.x、Ant Design 5.x、Zustand、React Query
- 数据库: PostgreSQL、Redis、Kafka
- 监控: Prometheus、Loki、Tempo、OpenTelemetry
- 部署: Kubernetes + Helm

**架构需求:**
- 微服务架构（host-svc、k8s-svc、obs-svc、ai-svc、api-gateway）
- API Gateway 统一认证和协议转换（gRPC-Gateway）
- OpenTelemetry 数据采集
- Agent 轻量级部署（DaemonSet / Systemd）

**Starter Template:**
- 无第三方 starter template，需要从零开始搭建
- 架构文档中已定义完整项目结构

### FR Coverage Map

| FR | 描述 | 史诗 |
|----|------|------|
| FR1 | 主机发现与注册 | Epic 2 |
| FR2 | Web SSH 终端 | Epic 3 |
| FR3 | 批量任务执行 | Epic 3 |
| FR4 | K8s 集群接入 | Epic 4 |
| FR5 | CRD 资源管理 | Epic 4 |
| FR5 | Pod 日志/终端 | Epic 5 |
| FR6 | Helm 应用管理 | Epic 5 |
| FR7 | OTel 数据采集 | Epic 6 |
| FR8 | 统一查询 API | Epic 7 |
| FR9 | 异常检测算法 | Epic 8 |
| FR10 | 告警聚合与根因 | Epic 8 |
| FR11 | LLM 集成 | Epic 9 |
| FR12 | 容量与故障预测 | Epic 8 |
| FR13 | 认证方式 | Epic 1 |
| FR14 | RBAC 权限模型 | Epic 10 |
| FR15 | 审计日志 | Epic 10 |
| FR16 | 主机界面/详情 | Epic 2 / Epic 3 |
| FR17 | K8s 管理界面 | Epic 4 / Epic 5 |
| FR18 | 可观测性界面 | Epic 7 |
| FR19 | AI 智能助手界面 | Epic 9 |

## Epic List

### Epic 1: 平台基础与用户认证
**用户成果**: 运维人员可以安全地登录平台并开始使用
**FRs 覆盖**: FR13
**说明**: 用户注册、登录（本地账号/LDAP）、API Gateway 和认证中间件、平台基础架构搭建

### Epic 2: 主机发现与管理
**用户成果**: 运维人员可以发现、注册并查看主机资产台账
**FRs 覆盖**: FR1, FR16
**说明**: SSH 扫描和 Agent 主动上报、主机注册与审批、主机列表与详情展示

### Epic 3: 主机远程运维
**用户成果**: 运维人员可以通过 Web SSH 远程操作主机并执行批量任务
**FRs 覆盖**: FR2, FR3
**说明**: Web SSH 终端、文件传输和管理、进程/服务管理、批量任务执行

### Epic 4: Kubernetes 集群接入
**用户成果**: 运维人员可以连接多个 K8s 集群并查看资源
**FRs 覆盖**: FR4, FR5, FR17
**说明**: 多集群管理（Kubeconfig 导入）、Workload、Service、Ingress 等 CRD 资源查看

### Epic 5: Kubernetes Pod 交互与 Helm
**用户成果**: 运维人员可以查看 Pod 日志/终端，并管理 Helm 应用
**FRs 覆盖**: FR5, FR6
**说明**: Pod 日志流与终端执行、Helm Chart 仓库与应用管理

### Epic 6: 可观测性数据采集
**用户成果**: 系统开始采集和存储监控数据
**FRs 覆盖**: FR7
**说明**: OpenTelemetry Collector 部署、对接 Prometheus/Loki/Tempo

### Epic 7: 可观测性查询与展示
**用户成果**: 运维人员可以查询指标、日志和链路，查看仪表盘
**FRs 覆盖**: FR8, FR18
**说明**: 统一查询 API、仪表盘、指标/日志查询界面、链路追踪

### Epic 8: AI 异常检测与智能告警
**用户成果**: 系统自动检测异常并聚合告警
**FRs 覆盖**: FR9, FR10, FR12
**说明**: 时序异常检测、告警聚合、降噪、根因关联、容量与故障预测

### Epic 9: AI 智能助手
**用户成果**: 运维人员可以用自然语言查询并获得智能诊断建议
**FRs 覆盖**: FR11, FR19
**说明**: 自然语言查询转换、LLM 智能诊断助手

### Epic 10: RBAC 与审计日志
**用户成果**: 管理员可以控制用户权限并追踪所有操作
**FRs 覆盖**: FR14, FR15
**说明**: RBAC 权限模型、审计日志记录与查询

---

## Epic 1: 平台基础与用户认证

运维人员可以安全地登录平台并开始使用。

### Story 1.1: 项目初始化与数据库设置

作为一名开发者，
我想要初始化 Go monorepo 项目结构和 PostgreSQL 数据库，
以便为后续开发提供基础架构。

**验收标准:**

**Given** 开发环境已安装 Go 1.23+ 和 Docker
**When** 执行项目初始化脚本
**Then** 创建以下目录结构：
- `backend/` 目录
- `backend/pkg/` 共享代码包
- `backend/api-gateway/`、`backend/host-svc/` 等服务目录
- `frontend/` 前端目录
- `deploy/` 部署配置目录

**And** 每个 Go 服务目录包含：
- `cmd/server/main.go` 入口文件
- `internal/config/` 配置目录
- `go.mod` 模块文件

**And** PostgreSQL 数据库已创建：
- 数据库名称：`myops`
- 执行迁移工具：`golang-migrate`
- 创建 `users` 表：`id (UUID), username, email, password_hash, created_at, updated_at`
- 创建 `migrations` 目录并初始化版本控制

**And** 前端项目已初始化：
- 使用 Vite 6.x 创建 React + TypeScript 项目
- 安装 Ant Design 5.x、Zustand、React Query、Axios

---

### Story 1.2: API Gateway 与认证中间件

作为一名开发者，
我想要搭建 API Gateway 和 JWT 认证中间件，
以便所有 API 请求都经过统一的认证和协议转换。

**验收标准:**

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
- Protobuf 文件放在 `backend/proto/` 目录
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

---

### Story 1.3: 用户注册 API

作为一名新用户，
我想要注册一个账号，
以便登录平台使用功能。

**验收标准:**

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

---

### Story 1.4: 用户登录与 Token 生成

作为一名已注册用户，
我想要登录并获取访问令牌，
以便使用平台功能。

**验收标准:**

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

---

### Story 1.5: 前端登录页面

作为一名用户，
我想要通过网页登录平台，
以便访问管理界面。

**验收标准:**

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

---

### Story 1.6: LDAP 认证集成

作为一名企业用户，
我想要使用公司 LDAP/AD 账号登录，
以便无需单独注册账号。

**验收标准:**

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

---

## Epic 2: 主机发现与管理

运维人员可以发现、注册并查看主机资产台账。

### Story 2.1: 主机数据模型与 API

作为一名开发者，
我想要创建主机管理的数据模型和基础 API，
以便支持主机信息的存储和查询。

**验收标准:**

**Given** 数据库已运行
**When** 执行主机相关的数据库迁移
**Then** 创建 `hosts` 表：
```sql
CREATE TABLE hosts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  hostname VARCHAR(255) NOT NULL,
  ip_address INET NOT NULL,
  port INTEGER DEFAULT 22,
  status VARCHAR(50) DEFAULT 'pending',
  os_type VARCHAR(100),
  os_version VARCHAR(100),
  cpu_cores INTEGER,
  memory_gb INTEGER,
  disk_gb BIGINT,
  labels JSONB,
  tags TEXT[],
  cluster_id UUID REFERENCES clusters(id),
  registered_by UUID REFERENCES users(id),
  approved_by UUID REFERENCES users(id),
  approved_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(ip_address, port)
);
```

**And** 创建索引：
- `idx_hosts_status` on `status`
- `idx_hosts_hostname` on `hostname`
- `idx_hosts_labels` on `labels` USING GIN

**And** 实现以下 gRPC 服务：
```protobuf
service HostService {
  rpc GetHost(GetHostRequest) returns (Host);
  rpc ListHosts(ListHostsRequest) returns (ListHostsResponse);
  rpc CreateHost(CreateHostRequest) returns (Host);
  rpc UpdateHost(UpdateHostRequest) returns (Host);
  rpc DeleteHost(DeleteHostRequest) returns (Empty);
}
```

**And** API Gateway 暴露 REST 端点：
- `GET /api/v1/hosts` - 列表查询（支持分页、过滤）
- `GET /api/v1/hosts/:id` - 获取详情
- `POST /api/v1/hosts` - 创建主机
- `PUT /api/v1/hosts/:id` - 更新主机
- `DELETE /api/v1/hosts/:id` - 删除主机

---

### Story 2.2: SSH 被动扫描

作为一名运维人员，
我想要通过扫描 IP 段发现可访问的主机，
以便快速添加大量主机到平台。

**验收标准:**

**Given** 用户已登录且有主机创建权限
**When** 发送 POST 请求到 `/api/v1/hosts/scan`
```json
{
  "ipRange": "192.168.1.0/24",
  "ports": [22, 2222],
  "timeout": 5
}
```
**Then** 返回 202 异步任务已创建：
```json
{
  "data": {
    "taskId": "scan-task-uuid",
    "status": "running",
    "ipRange": "192.168.1.0/24",
    "estimatedHosts": 254
  }
}
```

**And** 扫描任务在后台执行：
- 并发扫描（最多 50 个并发连接）
- 超时 5 秒未响应则跳过
- 成功连接的主机收集基本信息（hostname、OS 类型）

**And** 发现的主机自动保存到数据库：
- 状态设置为 `pending`
- `registered_by` 设置为当前用户

**And** 扫描完成后可通过 `GET /api/v1/hosts/scan-tasks/:taskId` 查询结果：
```json
{
  "data": {
    "taskId": "scan-task-uuid",
    "status": "completed",
    "discoveredHosts": 45,
    "hosts": [
      {"ipAddress": "192.168.1.10", "hostname": "web-01", "status": "pending"}
    ]
  }
}
```

---

### Story 2.3: Agent 开发与主动上报

作为一名开发者，
我想要开发主机 Agent 程序，
以便主机能主动上报实时信息。

**验收标准:**

**Given** Go 开发环境已配置
**When** 构建 Agent 程序
**Then** Agent 可执行文件生成：
- Linux amd64/arm64
- 支持 systemd 服务安装
- 支持 K8s DaemonSet 部署

**And** Agent 实现信息采集模块：
- 系统信息（hostname、OS 类型/版本、内核版本）
- 硬件信息（CPU 型号/核心数、内存总量、磁盘信息）
- 网络信息（IP 地址、MAC 地址、网关）
- 进程列表（可选，按需采集）

**And** Agent 实现上报功能：
- HTTP POST 到 `/api/v1/agent/report`
- 请求体包含采集的主机信息
- 携带 API Token 认证（从配置文件读取）

**And** Agent 配置文件 `/etc/myops-agent/config.yaml`：
```yaml
server:
  endpoint: "https://myops.example.com"
  token: "agent-token-xxx"
report:
  interval: 60
collector:
  collect_processes: false
  collect_network: true
```

**And** Agent 安装脚本：
- 自动检测系统类型
- 安装 systemd 服务文件
- 生成并注册 API Token

**And** 后端实现 Agent 上报 API `POST /api/v1/agent/report`：
- 验证 API Token
- 主机不存在则自动创建
- 已存在则更新主机信息
- 更新 `last_seen_at` 时间戳

---

### Story 2.4: 主机注册与审批

作为一名管理员，
我想要审批新注册的主机，
以便控制哪些主机可以纳入管理。

**验收标准:**

**Given** 存在待审批的主机
**When** 管理员访问主机列表
**Then** 显示状态过滤器：`全部`、`待审批`、`已批准`、`已拒绝`

**And** 主机列表显示待审批主机：
- 主机名、IP 地址、状态、注册时间
- 批准按钮、拒绝按钮

**When** 管理员点击批准按钮
**Then** 显示确认对话框："批准此主机？"

**And** 确认后调用 `PUT /api/v1/hosts/:id/approve`

**Then** 主机状态更新为 `approved`
- `approved_by` 设置为当前用户
- `approved_at` 设置为当前时间

**And** 批准后 Agent 可以正常上报数据

**When** 管理员点击拒绝按钮
**Then** 显示拒绝原因输入框

**And** 提交后调用 `PUT /api/v1/hosts/:id/reject`
```json
{
  "reason": "未授权的主机"
}
```

**Then** 主机状态更新为 `rejected`
- Agent 的后续上报请求返回 403 错误

**And** 自动审批规则（可配置）：
- 特定 IP 段自动批准
- 特定标签自动批准

---

### Story 2.5: 前端主机列表页

作为一名运维人员，
我想要查看和管理所有主机，
以便快速了解主机资产情况。

**验收标准:**

**Given** 用户已登录
**When** 访问 `/hosts` 路由
**Then** 显示主机列表页面，包含：

**And** 页面顶部搜索栏：
- 主机名/ IP 地址搜索框
- 状态下拉筛选（全部/在线/离线/待审批）
- 标签筛选（多选）

**And** Ant Design Table 显示主机列表：
- 列：主机名、IP 地址、状态、OS 类型、CPU、内存、标签、操作
- 支持分页（每页 20 条，可调整）
- 支持列排序（主机名、IP、创建时间）
- 状态徽章：在线(绿色)、离线(灰色)、待审批(橙色)

**And** 操作列：
- "详情" 按钮 - 跳转到主机详情页
- "编辑" 按钮 - 弹出编辑对话框（权限检查）
- "删除" 按钮 - 确认后删除主机（权限检查）

**And** 列表支持实时更新：
- 使用 WebSocket 接收主机状态变化
- 状态变更时自动刷新行数据

**And** 空状态：
- 无主机时显示空状态插图
- 显示"开始扫描"和"安装 Agent"引导按钮

**And** 加载状态：
- Table 显示 skeleton 加载动画
- 数据加载失败显示错误提示

---

### Story 2.6: 前端主机详情页

作为一名运维人员，
我想要查看主机的详细信息，
以便了解主机的完整配置和状态。

**验收标准:**

**Given** 用户已登录
**When** 从主机列表点击"详情"或访问 `/hosts/:id`
**Then** 显示主机详情页面，包含：

**And** 页面顶部：
- 面包屑导航：主机管理 > 主机详情
- 主机名和状态徽章
- 操作按钮：编辑、删除、刷新

**And** 基本信息卡片：
- 主机名、IP 地址、端口
- 操作系统（类型、版本、内核版本）
- 硬件信息（CPU 型号/核心数、内存总量、磁盘信息）
- 网络信息（IP、MAC、网关、DNS）
- 注册时间、最后心跳时间

**And** 标签管理：
- 显示已分配的标签（可删除）
- 添加标签输入框（回车添加）

**And** 监控图表卡片（占位，Epic 7 实现）：
- CPU 使用率趋势图
- 内存使用率趋势图
- 磁盘使用率趋势图
- 网络流量图

**And** 操作记录表格：
- 时间、操作人、操作类型、详情
- 显示最近 20 条记录

**And** 使用 React Query 获取数据：
- 自动刷新（每 30 秒）
- 手动刷新按钮

**And** 加载和错误状态：
- Card 显示 skeleton
- 错误显示 Alert 提示

---

## Epic 3: 主机远程运维

运维人员可以通过 Web SSH 远程操作主机并执行批量任务。

### Story 3.1: SSH 代理服务

作为一名开发者，
我想要实现 SSH 代理服务，
以便前端可以通过 WebSocket 连接到远程主机。

**验收标准:**

**Given** host-svc 已运行
**When** 前端发起 WebSocket 连接到 `/ws/ssh/:hostId`
**Then** 后端建立到目标主机的 SSH 连接

**And** SSH 连接配置：
- 使用主机 `ip_address` 和 `port`
- 支持密码认证和密钥认证
- 连接超时 30 秒

**And** WebSocket 双向通信：
- 前端 → 后端 → SSH：终端输入
- SSH → 后端 → 前端：终端输出（包括 ANSI 转义序列）
- 支持终端 resize（`rows`、`cols`）

**And** 连接管理：
- 断开时清理 SSH 会话
- 支持心跳检测（30 秒超时）
- 记录连接日志（用户、主机、连接时长）

**And** 错误处理：
- 主机不可达返回错误消息
- 认证失败返回错误消息
- 连接超时返回错误消息

---

### Story 3.2: Web SSH 终端组件

作为一名运维人员，
我想要在网页中使用 SSH 终端，
以便直接操作远程主机。

**验收标准:**

**Given** 用户已登录并有主机权限
**When** 访问 `/hosts/:id/terminal` 或从主机详情点击"终端"
**Then** 显示全屏终端组件

**And** 使用 xterm.js 渲染终端：
- 终端尺寸自适应窗口大小
- 支持常见的 ANSI 颜色和样式
- 支持特殊键（Ctrl+C、Ctrl+D 等）

**And** WebSocket 连接管理：
- 组件挂载时连接 `/ws/ssh/:hostId`
- 组件卸载时断开连接
- 连接失败显示错误提示并支持重连

**And** 终端工具栏：
- 显示主机名和 IP
- "断开连接"按钮
- "清屏"按钮
- "下载日志"按钮

**And** 支持多标签页：
- 可同时打开多个终端
- 标签显示主机名
- 支持切换和关闭标签

**And** 终端状态指示：
- 连接中：显示 spinner
- 已连接：显示绿点
- 已断开：显示红点

---

### Story 3.3: 文件传输 API

作为一名开发者，
我想要实现文件传输 API，
以便支持文件的上传、下载和管理。

**验收标准:**

**Given** host-svc 已运行
**When** 发送文件上传请求 `POST /api/v1/hosts/:id/files/upload`
**Then** 支持 multipart/form-data 上传

**And** 上传请求处理：
- 验证文件大小（最大 100MB）
- 验证文件名（防止路径穿越）
- 支持指定目标路径（默认 `/tmp`）

**And** 上传成功返回：
```json
{
  "data": {
    "path": "/tmp/uploaded-file.txt",
    "size": 1024,
    "uploadedAt": "2026-02-02T10:00:00Z"
  }
}
```

**And** 实现文件下载 `GET /api/v1/hosts/:id/files/download`：
- 查询参数 `?path=/path/to/file`
- 流式响应文件内容
- 设置正确的 Content-Type 和 Content-Disposition

**And** 实现文件列表 `GET /api/v1/hosts/:id/files`：
- 查询参数 `?path=/path/to/dir`
- 返回目录内容

**And** 实现文件删除 `DELETE /api/v1/hosts/:id/files`：
- 请求体 `{"path": "/path/to/file"}`
- 支持递归删除目录（`recursive: true`）

---

### Story 3.4: 文件管理界面

作为一名运维人员，
我想要通过网页管理远程主机上的文件，
以便方便地上传、下载和查看文件。

**验收标准:**

**Given** 用户已登录
**When** 从主机详情点击"文件管理"标签
**Then** 显示文件浏览器组件

**And** 左侧面包屑导航：
- 显示当前路径
- 点击路径段快速跳转

**And** 文件列表显示：
- 图标区分文件和目录
- 文件名、大小、修改时间
- 点击目录进入，点击文件预览

**And** 操作按钮：
- "上传文件"按钮
- "新建目录"按钮
- "刷新"按钮

**And** 文件操作菜单：
- 下载文件
- 删除文件/目录
- 重命名

**And** 拖拽上传：
- 支持拖拽文件到列表区域
- 显示上传进度条

**And** 文件预览：
- 点击文本文件显示预览对话框
- 支持语法高亮

---

### Story 3.5: 进程管理 API

作为一名开发者，
我想要实现进程管理 API，
以便支持查看和管理主机进程。

**验收标准:**

**Given** host-svc 已运行
**When** 获取进程列表 `GET /api/v1/hosts/:id/processes`
**Then** 返回进程列表

**And** 支持过滤和排序：
- 查询参数 `?name=nginx&status=running`
- 排序参数 `?sort=cpu&order=desc`

**And** 实现进程详情 `GET /api/v1/hosts/:id/processes/:pid`

**And** 实现进程操作：
- `POST /api/v1/hosts/:id/processes/:pid/stop` - 停止进程
- `POST /api/v1/hosts/:id/processes/:pid/kill` - 强制停止
- `POST /api/v1/hosts/:id/processes/:id/restart` - 重启进程

---

### Story 3.6: 批量任务引擎

作为一名开发者，
我想要实现批量任务引擎，
以便支持在多台主机上并发执行命令。

**验收标准:**

**Given** host-svc 已运行
**When** 创建数据库迁移
**Then** 创建 `tasks` 和 `task_executions` 表

**And** 实现任务调度器：
- 从 `task_executions` 表获取待执行任务
- 并发控制（最多同时执行 10 个）
- 任务状态机：pending → running → success/failed

**And** 实现任务 API

**And** 任务类型支持：
- `command` - 执行 shell 命令
- `script` - 执行脚本文件
- `upload` - 上传文件

**And** 执行策略：
- 串行执行：按主机顺序依次执行
- 并行执行：同时执行（默认，限制并发数）

---

### Story 3.7: 批量任务执行界面

作为一名运维人员，
我想要创建和监控批量任务，
以便在多台主机上批量执行操作。

**验收标准:**

**Given** 用户已登录
**When** 访问 `/tasks` 或从主机列表选择多台主机后点击"批量操作"
**Then** 显示批量任务创建页面

**And** 任务创建表单：
- 任务名称（必填）
- 任务类型：命令 / 脚本 / 文件上传
- 主机选择器
- 执行策略：串行 / 并行
- 并发数限制

**And** 命令类型输入：
- 命令输入框（多行文本）
- 环境变量编辑器

**And** 提交后跳转到任务详情页：
- 显示任务状态
- 总体进度条

**And** 执行结果展示：
- 表格显示每台主机的执行状态
- 点击行查看详细输出

**And** 操作按钮：
- "停止任务"按钮
- "重新执行"按钮
- "下载报告"按钮

---

## Epic 4: Kubernetes 集群接入

运维人员可以连接多个 K8s 集群并查看资源。

### Story 4.1: K8s 集群数据模型与 API

作为一名开发者，
我想要创建 K8s 集群管理的数据模型和基础 API，
以便支持多集群管理。

**验收标准:**

**Given** 数据库已运行
**When** 执行 K8s 相关的数据库迁移
**Then** 创建 `k8s_clusters` 表

**And** 使用 client-go 封装 K8s API 客户端

**And** 实现集群管理的 gRPC 和 REST API

---

### Story 4.2: 集群接入（Kubeconfig 导入）

作为一名运维人员，
我想要导入 Kubeconfig 连接集群，
以便在平台中管理 K8s 资源。

**验收标准:**

**Given** 用户已登录
**When** 访问 `/clusters/new` 页面
**Then** 显示集群接入表单，支持上传或粘贴 Kubeconfig

**And** 提交后验证连接，解析集群信息并保存

---

### Story 4.3: Workload 管理界面

作为一名运维人员，
我想要查看 Deployment、StatefulSet、DaemonSet，
以便了解应用负载状态。

**验收标准:**

**Given** 集群已接入
**When** 访问 `/clusters/:id/workloads`
**Then** 显示 Workload 列表，支持 Tabs 切换，显示副本数、镜像、年龄

---

### Story 4.4: Service/Ingress/ConfigMap/Secret 管理

作为一名运维人员，
我想要查看和管理 K8s 资源，
以便配置应用服务。

**验收标准:**

**Given** 集群已接入
**When** 访问对应资源页面
**Then** 显示资源列表，支持查看详情和编辑

---

## Epic 5: Kubernetes Pod 交互与 Helm

运维人员可以查看 Pod 日志/终端，并管理 Helm 应用。

### Story 5.1: Pod 列表与详情界面

**目标**: 查看 Pod 列表和详情

### Story 5.2: Pod 日志流（WebSocket）

**目标**: 实时查看 Pod 日志

### Story 5.3: Pod 终端（WebSocket）

**目标**: 进入 Pod 容器执行命令

### Story 5.4: Helm 仓库管理

**目标**: 管理 Helm Chart 仓库

### Story 5.5: Helm 应用管理界面

**目标**: 管理 Helm Release

---

## Epic 6: 可观测性数据采集

系统开始采集和存储监控数据。

### Story 6.1: OpenTelemetry Collector 部署

**目标**: 部署 OTel Collector

### Story 6.2: Prometheus 对接

**目标**: 配置 Prometheus 接收指标

### Story 6.3: Loki 对接

**目标**: 配置 Loki 接收日志

### Story 6.4: Tempo 对接

**目标**: 配置 Tempo 接收链路

---

## Epic 7: 可观测性查询与展示

运维人员可以查询指标、日志和链路，查看仪表盘。

### Story 7.1: 统一查询 API

**目标**: 实现统一查询 API

### Story 7.2: 仪表盘页面

**目标**: 显示概览仪表盘

### Story 7.3: 指标查询界面

**目标**: 查询和展示 Prometheus 指标

### Story 7.4: 日志查询界面

**目标**: 查询 Loki 日志

### Story 7.5: 链路追踪界面

**目标**: 查看 Tempo 链路

---

## Epic 8: AI 异常检测与智能告警

系统自动检测异常并聚合告警。

### Story 8.1: 时序异常检测模型

**目标**: 实现指标异常检测

### Story 8.2: 告警聚合与降噪

**目标**: 聚合和降噪告警

### Story 8.3: 根因关联分析

**目标**: 分析告警根因

### Story 8.4: 预测分析

**目标**: 预测容量和故障

### Story 8.5: 告警管理界面

**目标**: 查看和管理告警

---

## Epic 9: AI 智能助手

运维人员可以用自然语言查询并获得智能诊断建议。

### Story 9.1: 自然语言查询转换

**目标**: 将自然语言转换为查询

### Story 9.2: LLM 集成 API

**目标**: 集成 LLM API

### Story 9.3: RAG 知识库

**目标**: 构建诊断知识库

### Story 9.4: 智能助手界面

**目标**: 聊天式交互

---

## Epic 10: RBAC 与审计日志

管理员可以控制用户权限并追踪所有操作。

### Story 10.1: RBAC 数据模型

**目标**: 创建权限模型表

### Story 10.2: 权限管理 API

**目标**: 实现权限管理 API

### Story 10.3: 角色管理界面

**目标**: 管理角色和权限

### Story 10.4: 审计日志记录

**目标**: 记录所有操作

### Story 10.5: 审计日志查询界面

**目标**: 查询审计日志
