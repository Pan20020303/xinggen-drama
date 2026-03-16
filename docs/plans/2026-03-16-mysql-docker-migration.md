# MySQL 默认部署与 SQLite 历史数据迁移 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将项目默认部署数据库切换为 MySQL，保留 SQLite 兼容能力，并支持把现有 SQLite 历史数据一次性迁移到 Docker 内的 MySQL。

**Architecture:** 保持应用代码同时支持 `sqlite/mysql` 两种数据库类型，但把 `config.example.yaml`、Docker Compose 和镜像启动流程切换到 MySQL 默认值。新增一个独立的 SQLite 到 MySQL 迁移命令和 Docker 启动脚本，在检测到旧 SQLite 数据文件且 MySQL 目标库为空时执行一次导入。

**Tech Stack:** Go, GORM, MySQL 8, Docker Compose, Bash/sh, Testify

---

### Task 1: 为迁移工具抽取可测试的纯函数并先写失败测试

**Files:**
- Create: `infrastructure/database/sqlite_to_mysql_test.go`
- Modify: `infrastructure/database/sqlite_to_mysql.go`

**Step 1: Write the failing test**

- 为迁移表顺序写测试，确保用户、剧本、章节、分镜等按依赖顺序导入
- 为扫描值归一化写测试，确保 `[]byte` 会转为字符串、`nil` 会保留
- 为“目标库非空时拒绝导入”判定逻辑写测试

**Step 2: Run test to verify it fails**

Run: `go test ./infrastructure/database -run TestSQLiteToMySQL -v`
Expected: FAIL because helper functions do not exist yet

### Task 2: 实现 SQLite 到 MySQL 迁移工具

**Files:**
- Create: `infrastructure/database/sqlite_to_mysql.go`
- Create: `cmd/sqlite_to_mysql/main.go`
- Test: `infrastructure/database/sqlite_to_mysql_test.go`

**Step 1: Write minimal implementation**

- 定义迁移表顺序
- 连接 SQLite 源库和 MySQL 目标库
- 对目标库执行 `AutoMigrate`
- 检查目标库是否为空，不为空则直接报错避免重复导入
- 按表扫描 SQLite 数据并写入 MySQL，保留主键与时间字段

**Step 2: Run the tests**

Run: `go test ./infrastructure/database -run TestSQLiteToMySQL -v`
Expected: PASS

### Task 3: 调整配置默认值到 MySQL，并补配置测试

**Files:**
- Modify: `configs/config.example.yaml`
- Modify: `pkg/config/config.go`
- Create: `pkg/config/config_test.go`

**Step 1: Write the failing test**

- 断言 MySQL DSN 生成正确
- 断言 SQLite 兼容逻辑仍保留

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/config -run TestDatabaseConfig -v`
Expected: FAIL before implementation is complete

**Step 3: Write minimal implementation**

- 保持 `sqlite/mysql` 双模式
- 将示例配置默认值改为 MySQL

**Step 4: Run tests**

Run: `go test ./pkg/config -run TestDatabaseConfig -v`
Expected: PASS

### Task 4: 改造 Docker Compose、Dockerfile 和启动脚本

**Files:**
- Modify: `docker-compose.yml`
- Modify: `Dockerfile`
- Create: `docker/entrypoint.sh`

**Step 1: Implement Docker changes**

- 新增 `mysql` 服务、healthcheck、持久化卷
- 应用服务依赖 MySQL 健康状态
- 构建并复制 `sqlite_to_mysql` 可执行文件
- 启动脚本在检测到旧 SQLite 文件且迁移未执行时，自动触发导入

**Step 2: Verify Docker config syntax**

Run: `docker compose config`
Expected: PASS with valid merged configuration

### Task 5: 更新文档和运维说明

**Files:**
- Modify: `README-CN.md`
- Modify: `README.md`
- Modify: `docs/PROJECT_ARCHITECTURE.md`

**Step 1: Update docs**

- 默认数据库从 SQLite 改为 MySQL
- 补 Docker 启动与历史数据迁移说明
- 标明 SQLite 仍可用于本地兼容调试

### Task 6: 最终验证

**Files:**
- Review: `docker-compose.yml`
- Review: `configs/config.example.yaml`
- Review: `cmd/sqlite_to_mysql/main.go`

**Step 1: Run targeted Go tests**

Run: `go test ./infrastructure/database ./pkg/config -v`
Expected: PASS

**Step 2: Run compose config validation**

Run: `docker compose config`
Expected: PASS

**Step 3: Run full Go test subset covering changed packages**

Run: `go test ./infrastructure/database ./pkg/config ./cmd/sqlite_to_mysql/... -v`
Expected: PASS
