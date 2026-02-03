# MyOps AIOps Platform

面向中型企业（50-500 用户，管理 1000-10000 台主机）的 AIOps 平台，提供主机管理、K8s 管理、可观测性和 AI 智能分析四大核心能力。

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
