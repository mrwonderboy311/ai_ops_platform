# 开发环境设置指南

## 前置要求

- Go 1.23+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 16+ (或使用 Docker)

## 后端设置

```bash
cd backend

# 同步 Go 工作区依赖
go work sync

# 下载依赖
cd api-gateway && go mod download
cd ../host-svc && go mod download
# ... 其他服务类似

# 运行服务（每个终端一个）
cd api-gateway && go run cmd/server/main.go
```

## 前端设置

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

## 数据库设置

```bash
# 启动 Docker 服务
cd deploy
docker-compose up -d

# 运行数据库迁移
cd ../backend/api-gateway
migrate -path migrations -database "postgres://myops:myops_dev_pass@localhost:5432/myops?sslmode=disable" up
```

## 项目结构

详见 `docs/plans/2025-02-02-aiops-platform-design.md`
