---
title: SQLite 切换 MySQL 运行手册
owner: engineering
last_updated: 2026-03-19
review_schedule: as-needed
---

# SQLite 切换 MySQL 运行手册

> **TL;DR:** 这个手册用于把本地 `sqlite` 数据迁移到 Docker MySQL，并把服务切到 MySQL 运行。

## Definition of Done

这次切换完成时，必须同时满足：
- [ ] `configs/config.yaml` 中 `database.type=mysql`
- [ ] `xinggen-mysql`、`xinggen-rabbitmq`、`xinggen-drama` 三个容器都处于 `healthy`
- [ ] `http://localhost:5678/health` 返回 `200`
- [ ] MySQL 关键业务表数据已导入
- [ ] 启动日志中的 `Storyboard user ownership check` 没有残留 `remaining_zero_rows`

## 什么时候使用

在以下场景使用：
- 本地开发准备从 SQLite 切到 MySQL
- Docker MySQL 是空库，需要从 `data/drama_generator.db` 导入历史数据
- 历史数据中存在 `user_id = 0`，需要在切换后修复归属

## 当前仓库约定

宿主机直接运行后端时，使用：
- MySQL: `localhost:3306`
- RabbitMQ: `localhost:5672`
- 配置文件: `configs/config.yaml`

Docker 容器内部运行后端时，使用：
- MySQL: `mysql:3306`
- RabbitMQ: `rabbitmq:5672`
- 由 [docker-compose.yml](/D:/code/github/huobao-drama/docker-compose.yml) 注入环境变量

不要混用 `localhost` 和容器服务名。

## 前置条件

开始前先确认：
- Docker Desktop 已启动
- 本地 SQLite 文件存在：`data/drama_generator.db`
- 你能接受强制迁移会清空 MySQL 目标库
- 你知道匿名历史数据最终要归属给哪个用户

## 第 1 步：备份 SQLite

执行：

```powershell
Copy-Item .\data\drama_generator.db .\data\drama_generator.db.bak-$(Get-Date -Format 'yyyyMMdd_HHmmss')
```

验证：
- `data` 目录下出现新的 `.bak-*` 文件

## 第 2 步：把宿主机配置切到 MySQL

编辑 [config.yaml](/D:/code/github/huobao-drama/configs/config.yaml)，确认至少是下面这些值：

```yaml
database:
  type: mysql
  host: localhost
  port: 3306
  user: xinggen
  password: xinggen123
  database: xinggen_drama
  charset: utf8mb4

mq:
  enabled: true
  url: amqp://xinggen_rmq:XinggenRmq_2026_StrongPass@localhost:5672/
```

验证：
- `database.type` 不是 `sqlite`
- `mq.url` 指向 `localhost:5672`

## 第 3 步：启动基础容器

执行：

```powershell
docker compose up -d mysql rabbitmq
docker ps --format "table {{.Names}}`t{{.Status}}`t{{.Ports}}"
```

预期：
- `xinggen-mysql` 为 `healthy`
- `xinggen-rabbitmq` 为 `healthy`

## 第 4 步：如果要强制迁移，先重建目标库

只有在 MySQL 已经有旧数据，且你明确要覆盖时执行。

执行：

```powershell
docker exec xinggen-mysql mysql -uroot -proot123 -e "DROP DATABASE IF EXISTS xinggen_drama; CREATE DATABASE xinggen_drama CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

警告：
- 这一步会永久删除 MySQL 里的现有数据

## 第 5 步：执行 SQLite -> MySQL 迁移

仓库已经提供迁移入口 [main.go](/D:/code/github/huobao-drama/cmd/sqlite_to_mysql/main.go) 和实现 [sqlite_to_mysql.go](/D:/code/github/huobao-drama/infrastructure/database/sqlite_to_mysql.go)。

如果宿主机没有安装 Go，可以直接用 Go 容器执行：

```powershell
docker run --rm `
  -v "${PWD}:/app" `
  -w /app `
  --add-host host.docker.internal:host-gateway `
  -e DATABASE_HOST=host.docker.internal `
  golang:1.24 `
  bash -lc "export PATH=/usr/local/go/bin:$PATH && go run ./cmd/sqlite_to_mysql -source ./data/drama_generator.db -marker ./data/.sqlite_to_mysql_done"
```

预期输出：
- `migrated table=users rows=...`
- `migrated table=storyboards rows=...`
- `migrated table=image_generations rows=...`

说明：
- 迁移器会先自动建表
- 如果目标库非空，会直接拒绝导入，避免重复写入
- 成功后会写入标记文件 `data/.sqlite_to_mysql_done`

## 第 6 步：启动应用容器

执行：

```powershell
docker compose up -d xinggen-drama
docker ps --format "table {{.Names}}`t{{.Status}}"
```

预期：
- `xinggen-drama` 为 `healthy`

## 第 7 步：跑迁移后验证

执行：

```powershell
Invoke-WebRequest http://localhost:5678/health | Select-Object StatusCode, Content
docker logs --tail 120 xinggen-drama
docker exec xinggen-mysql mysql -uxinggen -pxinggen123 -D xinggen_drama -e "SELECT COUNT(*) AS users FROM users; SELECT COUNT(*) AS dramas FROM dramas; SELECT COUNT(*) AS episodes FROM episodes; SELECT COUNT(*) AS storyboards FROM storyboards; SELECT COUNT(*) AS image_generations FROM image_generations; SELECT COUNT(*) AS assets FROM assets; SELECT COUNT(*) AS async_tasks FROM async_tasks;"
```

重点检查：
- `/health` 返回 `200`
- 应用日志没有数据库连接错误
- 关键表行数与 SQLite 一致

## 第 8 步：修复 `user_id = 0` 的历史数据

如果启动日志或 SQL 检查发现匿名数据，需要先确认归属用户，再统一修复。

先查目标用户：

```powershell
docker exec xinggen-mysql mysql -uxinggen -pxinggen123 -D xinggen_drama -e "SELECT id,email,status,role FROM users;"
```

确认要归属的 `user_id` 后，执行：

```powershell
docker exec xinggen-mysql mysql -uxinggen -pxinggen123 -D xinggen_drama -e "
START TRANSACTION;
UPDATE dramas SET user_id = 3 WHERE user_id = 0;
UPDATE episodes SET user_id = 3 WHERE user_id = 0;
UPDATE storyboards SET user_id = 3 WHERE user_id = 0;
UPDATE scenes SET user_id = 3 WHERE user_id = 0;
UPDATE characters SET user_id = 3 WHERE user_id = 0;
UPDATE props SET user_id = 3 WHERE user_id = 0;
UPDATE frame_prompts SET user_id = 3 WHERE user_id = 0;
UPDATE image_generations SET user_id = 3 WHERE user_id = 0;
UPDATE video_generations SET user_id = 3 WHERE user_id = 0;
UPDATE assets SET user_id = 3 WHERE user_id = 0;
UPDATE character_libraries SET user_id = 3 WHERE user_id = 0;
UPDATE ai_service_configs SET user_id = 3 WHERE user_id = 0;
COMMIT;
"
```

然后重启应用：

```powershell
docker restart xinggen-drama
docker logs --tail 120 xinggen-drama
```

预期：
- `Storyboard user ownership check` 中 `remaining_zero_rows: 0`
- 同一条日志中的 `mismatch_rows: 0`
- 同一条日志中的 `orphan_rows: 0`

## 常用排障

### 现象：MySQL 没启动，但服务还能访问数据库

原因：
- 本地实际还在用 SQLite

检查：

```powershell
Get-Content .\configs\config.yaml
```

如果你看到：
- `database.type: sqlite`
- `database.path: ./data/drama_generator.db`

那当前就不是 MySQL。

### 现象：迁移命令提示目标库非空

原因：
- MySQL 已有历史数据

处理：
- 想保留旧数据：停止迁移，先人工比对
- 想覆盖旧数据：执行“第 4 步”重建数据库后再迁移

### 现象：应用启动后日志里仍然有 `remaining_zero_rows`

原因：
- [data_fixes.go](/D:/code/github/huobao-drama/infrastructure/database/data_fixes.go) 只会从 `episodes.user_id` 回填 `storyboards.user_id`
- 如果 `episodes` 或 `dramas` 自己也是 `user_id = 0`，就需要人工指定归属用户

## 本仓库本次实际切换结果

2026-03-19 的本地切换中，已经完成：
- SQLite 备份：`data/drama_generator.db.bak-20260319_152540`
- 宿主机配置切到 MySQL 和 RabbitMQ 本地端口
- Docker 容器启动：`xinggen-mysql`、`xinggen-rabbitmq`、`xinggen-drama`
- SQLite 数据导入 MySQL
- 历史匿名数据统一归属到 `akpp91299@gmail.com` 对应的 `user_id = 3`

## 相关文件

- [docker-compose.yml](/D:/code/github/huobao-drama/docker-compose.yml)
- [config.yaml](/D:/code/github/huobao-drama/configs/config.yaml)
- [main.go](/D:/code/github/huobao-drama/cmd/sqlite_to_mysql/main.go)
- [sqlite_to_mysql.go](/D:/code/github/huobao-drama/infrastructure/database/sqlite_to_mysql.go)
- [data_fixes.go](/D:/code/github/huobao-drama/infrastructure/database/data_fixes.go)
