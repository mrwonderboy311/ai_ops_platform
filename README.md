# MyOps AIOps Platform

> 智能运维平台 - AI驱动的Kubernetes主机与集群管理平台

面向中型企业（50-500 用户，管理 1000-10000 台主机）的 AIOps 平台，提供主机管理、K8s 管理、可观测性和 AI 智能分析四大核心能力。

## 🎯 项目完成度

**整体完成度：95%**

| 阶段 | 状态 | 完成度 |
|------|------|--------|
| Phase 1: 基础设施管理 | ✅ 完成 | 100% |
| Phase 2: Kubernetes管理 | ✅ 完成 | 100% |
| Phase 3: 应用管理 | ✅ 完成 | 100% |
| Phase 4: 可扩展性框架 | ✅ 完成 | 100% |
| Phase 5: 可观测性数据采集 | ✅ 完成 | 100% |
| Phase 6: AI智能分析 | ✅ 完成 | 100% |
| Phase 7: 安全与权限 (RBAC) | ✅ 完成 | 100% |
| Phase 8: 测试与优化 | 🚧 进行中 | 30% |

## 🚀 功能特性

### 基础设施管理
- **主机管理**
  - 主机列表、详情查看
  - SSH终端连接
  - 文件管理（上传、下载、删除、目录操作）
  - 进程管理（查看、终止、执行命令）
  - 批量任务执行

### Kubernetes集群管理
- **集群管理**
  - 集群连接、配置管理
  - 节点信息查看
  - 命名空间管理
  - 集群指标监控
- **工作负载管理**
  - Deployment、StatefulSet、DaemonSet管理
  - Pod详情、日志流式查看
  - Pod终端连接
  - Service管理

### 应用管理
- **Helm仓库管理**
  - 添加、同步、删除Helm仓库
  - Chart搜索与查看
  - 仓库连接测试
- **Helm应用管理**
  - 应用安装、升级、回滚
  - Release历史查看
  - 应用状态监控

### 可观测性数据采集
- **OpenTelemetry Collector**
  - Collector部署与管理
  - 启动、停止、重启操作
  - 实时状态监控
- **Prometheus集成**
  - 数据源管理（添加、编辑、删除、测试连接）
  - 告警规则配置（CRUD操作）
  - Dashboard管理（CRUD操作）
  - PromQL查询执行
- **Grafana集成**
  - 实例管理（添加、编辑、删除、测试连接）
  - Dashboard同步
  - 数据源同步
  - 文件夹管理

### AI智能分析
- **异常检测**
  - 多种算法支持：
    - STL（季节性分解）
    - Isolation Forest（孤立森林）
    - LSTM（深度学习）
    - Baseline Learning（基线学习）
  - 规则配置与管理
  - 实时检测与告警
- **LLM对话助手**
  - 自然语言交互
  - 智能问题诊断
  - 运维建议生成
  - 对话历史管理
- **自然语言查询**
  - 自然语言转PromQL
  - 智能查询解析
- **知识库**
  - RAG增强的知识管理
  - 最佳实践沉淀

### 安全与权限 (RBAC)
- **权限模型**
  - 细粒度资源-操作-范围权限
  - 40+预定义权限覆盖所有资源
  - 多级权限范围（global/cluster/namespace/host）
- **角色管理**
  - 角色层级继承
  - 系统角色保护
  - 默认角色分配
  - 权限批量分配
- **用户角色**
  - 资源级权限分配
  - 角色过期时间
  - 用户权限聚合
- **访问策略**
  - Allow/Deny策略
  - 标签选择器支持
- **预定义角色**
  - `super_admin` - 超级管理员（全部权限）
  - `admin` - 管理员（除用户管理外的全部权限）
  - `operator` - 运维人员（主机、集群、应用管理）
  - `viewer` - 查看者（只读权限）

## 项目结构

```
myops-k8s-platform/
├── backend/                 # Go monorepo 后端
│   ├── pkg/                  # 共享代码包
│   │   ├── auth/             # 认证工具
│   │   ├── db/               # 数据库连接
│   │   ├── kafka/            # Kafka 客户端
│   │   ├── redis/            # Redis 客户端
│   │   └── proto/            # Protobuf 定义
│   ├── api-gateway/          # API 网关
│   ├── host-svc/             # 主机管理服务
│   ├── k8s-svc/              # K8s 管理服务
│   ├── obs-svc/              # 可观测性服务
│   ├── ai-svc/               # AI 分析服务
│   └── go.work               # Go 工作区配置
├── frontend/                # React 前端
│   └── src/
│       ├── pages/           # 页面组件
│       ├── components/      # 通用组件
│       ├── hooks/           # 自定义 Hooks
│       ├── api/             # API 调用
│       ├── store/           # 状态管理
│       └── types/           # TypeScript 类型
├── agent/                   # 主机 Agent
├── deploy/                  # K8s 部署配置
│   └── docker-compose.yml   # 本地开发环境
├── docs/                    # 文档
└── README.md
```

## 技术栈

### 后端
- Go 1.23+
- gRPC + Protobuf
- GORM
- PostgreSQL
- Redis
- Kafka

### 前端
- React 18
- TypeScript 5.x
- Vite 6.x
- Ant Design 5.x
- Zustand
- React Query
- React Router v6

## 快速开始

### 开发环境要求

- Go 1.23+
- Node.js 20+
- Docker & Docker Compose

### 后端设置

```bash
cd backend
go work sync
cd api-gateway && go mod download
```

### 前端设置

```bash
cd frontend
npm install
npm run dev
```

### 数据库设置

```bash
cd deploy
docker-compose up -d
```

## 开发指南

详见 [开发环境设置指南](docs/development-setup.md)

## Epic 和 Story 状态

查看 [_bmad-output/implementation-artifacts/sprint-status.yaml](_bmad-output/implementation-artifacts/sprint-status.yaml)

## 许可证

MIT
