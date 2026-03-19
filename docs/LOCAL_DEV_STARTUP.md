---
title: 本地开发启动说明
owner: engineering
last_updated: 2026-03-19
review_schedule: as-needed
---

# 本地开发启动说明

> **TL;DR:** 日常开发默认使用 Docker 提供 MySQL 和 RabbitMQ，在宿主机直接运行 Go 后端和 Vite 前端。

## Definition of Done

本地开发环境准备完成时，必须满足：
- [ ] `xinggen-mysql` 和 `xinggen-rabbitmq` 容器为 `healthy`
- [ ] `go run main.go` 启动后无数据库和 MQ 连接错误
- [ ] `http://localhost:5678/health` 返回 `200`
- [ ] `cd web && npm run dev` 启动成功
- [ ] 浏览器可访问 `http://localhost:3012`

## 当前默认开发方式

宿主机运行：
- 后端：`go run main.go`
- 前端：`cd web && npm run dev`

Docker 提供依赖：
- MySQL：`localhost:3306`
- RabbitMQ：`localhost:5672`
- RabbitMQ 控制台：`http://localhost:15672`

不要把宿主机配置写成容器服务名：
- 宿主机运行用 `localhost`
- 容器内运行才用 `mysql` / `rabbitmq`

## 前置条件

需要本机具备：
- Go 1.23+
- Node.js 18+
- npm 9+
- Docker Desktop
- FFmpeg 可执行文件已加入 `PATH`

## 第 1 步：安装依赖

后端依赖：

```powershell
go mod download
```

前端依赖：

```powershell
cd web
npm install
```

## 第 2 步：确认本地配置

检查 [config.yaml](/D:/code/github/huobao-drama/configs/config.yaml)。

日常开发应至少满足：

```yaml
database:
  type: mysql
  host: localhost
  port: 3306
  user: xinggen
  password: xinggen123
  database: xinggen_drama

mq:
  enabled: true
  url: amqp://xinggen_rmq:XinggenRmq_2026_StrongPass@localhost:5672/

server:
  port: 5678
  cors_origins:
    - http://localhost:3012
```

如果你看到 `database.type: sqlite`，那当前不是 MySQL 开发模式。

## 第 3 步：启动基础依赖

执行：

```powershell
docker compose up -d mysql rabbitmq
docker ps --format "table {{.Names}}`t{{.Status}}`t{{.Ports}}"
```

预期：
- `xinggen-mysql` 为 `healthy`
- `xinggen-rabbitmq` 为 `healthy`

RabbitMQ 默认账号来自 [docker-compose.yml](/D:/code/github/huobao-drama/docker-compose.yml)：
- 用户名：`xinggen_rmq`
- 密码：`XinggenRmq_2026_StrongPass`

## 第 4 步：启动后端

执行：

```powershell
go run main.go
```

常用地址：
- 健康检查：`http://localhost:5678/health`
- API：`http://localhost:5678/api/v1`
- 静态文件：`http://localhost:5678/static`

如果只验证服务是否起来，可以单独访问：

```powershell
Invoke-WebRequest http://localhost:5678/health
```

## 第 5 步：启动前端

执行：

```powershell
cd web
npm run dev
```

前端默认地址：
- `http://localhost:3012`

Vite 代理配置在 [vite.config.ts](/D:/code/github/huobao-drama/web/vite.config.ts)，开发时会把：
- `/api` 代理到 `http://localhost:5678`
- `/static` 代理到 `http://localhost:5678`

## 日常验证命令

看容器状态：

```powershell
docker ps --format "table {{.Names}}`t{{.Status}}"
```

看后端健康：

```powershell
Invoke-WebRequest http://localhost:5678/health | Select-Object StatusCode, Content
```

看前端类型检查和构建：

```powershell
cd web
npm run build:check
```

看前端单测：

```powershell
cd web
npm test
```

## 常见场景

### 只想启动依赖，不跑应用容器

这是默认开发方式，执行：

```powershell
docker compose up -d mysql rabbitmq
```

不要启动 `xinggen-drama` 容器，否则你可能会同时有一份容器内后端和一份本机后端。

### 想看 RabbitMQ 控制台

打开：
- [http://localhost:15672](http://localhost:15672)

### 想让后端直接提供前端静态文件

执行：

```powershell
cd web
npm run build
cd ..
go run main.go
```

这适合快速验收，不适合日常前端开发。

## 常见问题

### MySQL 没启动，但应用还能访问数据库

原因：
- 你大概率又切回了 SQLite

检查 [config.yaml](/D:/code/github/huobao-drama/configs/config.yaml) 的 `database.type`。

### 前端能打开，但接口报错

优先检查：
- 后端是否已经启动
- `http://localhost:5678/health` 是否正常
- `web/vite.config.ts` 的代理目标是否还是 `http://localhost:5678`

### 后端启动报 MQ 连接错误

优先检查：
- `xinggen-rabbitmq` 是否健康
- `configs/config.yaml` 里的 `mq.url` 是否使用 `localhost:5672`

### 后端启动报数据库连接错误

优先检查：
- `xinggen-mysql` 是否健康
- `configs/config.yaml` 里的数据库主机是否是 `localhost`
- 账号是否还是 `xinggen / xinggen123`

## 相关文档

- [README-CN.md](/D:/code/github/huobao-drama/README-CN.md)
- [docker-compose.yml](/D:/code/github/huobao-drama/docker-compose.yml)
- [config.yaml](/D:/code/github/huobao-drama/configs/config.yaml)
- [MYSQL_SWITCH_AND_MIGRATION_RUNBOOK.md](/D:/code/github/huobao-drama/docs/MYSQL_SWITCH_AND_MIGRATION_RUNBOOK.md)
