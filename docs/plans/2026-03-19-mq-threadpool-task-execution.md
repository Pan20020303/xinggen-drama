# MQ + 线程池任务调度改造 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将当前基于进程内 `TaskRunner` 的异步任务执行模型，升级为“数据库任务表 + MQ 分发 + Worker 线程池消费”的可扩展任务调度体系，提升跨实例调度、失败恢复、并发控制和可观测性。

**Architecture:** 保留 `async_tasks` 作为任务事实源和用户可查询状态表，引入 MQ 只负责分发与唤醒，不让 MQ 直接承载业务状态。应用侧增加 `Dispatcher` 和 `WorkerPool` 两层：`Dispatcher` 负责把待执行任务投递到 MQ，`WorkerPool` 从 MQ 拉取任务并在进程内用受控线程池执行，执行结果统一回写 `async_tasks`。第一阶段采用 RabbitMQ 作为推荐方案，线程池继续进程内实现，但从“直接提交闭包”改为“按任务类型路由到注册 Handler”。

**Tech Stack:** Go, Gin, GORM, MySQL, RabbitMQ, goroutine worker pool

---

## 现状评估

当前项目已经有一套可用但局限明显的异步模型：

- `domain/models/task.go`
  - 维护 `AsyncTask` 表，包含 `type/status/progress/message/error/result/resource_id`
- `application/services/task_service.go`
  - 负责创建、复用、更新任务状态
- `application/services/task_runner.go`
  - 仅用 `chan struct{}` 做本地并发限制，然后 `go func()` 执行
- 多个服务直接依赖 `TaskRunner`
  - `ImageGenerationService`
  - `VideoGenerationService`
  - `StoryboardService`
  - `ScriptGenerationService`
  - `CharacterLibraryService`
  - `PropService`
  - `FramePromptService`
  - `VideoMergeService`

这套方案的问题不是“不能用”，而是它只能覆盖单实例、进程内执行：

- API 服务和任务执行强耦合，扩容后任务会分散在多个 Web 实例里
- 进程崩溃时，内存中的待执行任务直接丢失
- 当前 `TaskRunner` 无法做真正的重试、死信、延迟任务、优先级队列
- 各服务自己决定如何提交任务，任务入口分散，无法统一治理
- 恢复逻辑靠局部补丁，例如视频恢复 pending 任务，不是平台级机制

---

## 方案对比

### 方案 A：RabbitMQ + AsyncTask 表 + WorkerPool

**推荐方案**

做法：
- `async_tasks` 继续做状态源
- API 创建任务后，把任务 ID 投递到 RabbitMQ
- 独立 worker 进程或同进程 worker 订阅队列
- worker 收到消息后，从 DB 读取任务并执行
- 成功 `ack`，失败按规则 `nack/requeue` 或进死信队列

优点：
- 交付快，跟你当前模型最兼容
- 支持 ack、重试、死信、预取、延迟重投
- 很适合图片/视频/分镜这类中低吞吐、长耗时任务

缺点：
- 需要新增 RabbitMQ 运维组件
- 消息幂等需要自己保证

### 方案 B：Redis Streams + 消费组 + WorkerPool

做法与方案 A 类似，但消息层换成 Redis Streams。

优点：
- 如果项目已经重度依赖 Redis，部署成本低
- 支持消费组和 pending list

缺点：
- 死信、重试、延迟调度策略需要自己拼更多机制
- 语义上更灵活，但治理成本高于 RabbitMQ

### 方案 C：仅 DB 轮询 + WorkerPool

不引入 MQ，只靠 DB 扫描 `pending` 任务并竞争执行。

优点：
- 最省依赖

缺点：
- 数据库会承担分发职责，轮询成本高
- 跨实例竞争锁和抢占逻辑更复杂
- 最终大概率会自己手搓一个低配 MQ

**结论：**
如果你的目标是“真正把任务系统做成可扩展基础设施”，建议直接选 **方案 A：RabbitMQ + AsyncTask 表 + WorkerPool**。它最契合你当前代码结构，迁移成本最低，收益最大。

---

## 目标架构

```mermaid
flowchart LR
    API["API Handler / Service"] --> TS["TaskService<br/>Create task row"]
    TS --> DB[("async_tasks")]
    API --> D["TaskDispatcher"]
    D --> MQ["RabbitMQ"]
    MQ --> C["TaskConsumer"]
    C --> WP["WorkerPool"]
    WP --> H["Registered Task Handler"]
    H --> DB
    H --> EXT["AI / FFmpeg / Upload / Merge"]
    H --> ACK["Ack / Retry / DLQ"]
```

核心原则：

- **DB 是事实源**：任务状态、进度、结果、错误都只认 `async_tasks`
- **MQ 是分发总线**：只传 `task_id` + 最少元数据，不存业务结果
- **线程池是执行器**：限制单实例并发，避免一个实例吃满 CPU/网络
- **Handler 注册化**：不同任务类型统一注册，而不是在服务里到处写 `runner.Submit("xx", func(){...})`

---

## 需要新增的核心组件

### 1. TaskDispatcher

职责：
- 接收任务投递请求
- 将 `task_id`、`task_type`、`retry_count`、`resource_id` 投递到 MQ
- 对外暴露统一接口，例如：

```go
type TaskDispatcher interface {
    Dispatch(task *models.AsyncTask) error
}
```

说明：
- `CreateOrGetActiveTask` 之后不再直接 `runner.Submit`
- 改为 `dispatcher.Dispatch(task)`

### 2. TaskConsumer

职责：
- 长连接订阅 MQ
- 拿到消息后交给本地线程池
- 线程池执行完成后，根据结果 `ack/nack`

关键控制点：
- `prefetch` 限制单实例正在消费的未确认消息数量
- 不要消费后立刻开无限 goroutine，必须交给固定大小线程池

### 3. WorkerPool

职责：
- 用固定 worker 数量执行任务
- 支持优雅停机
- 支持任务上下文取消

和当前 `TaskRunner` 的差异：
- 现在 `TaskRunner` 是“提交一个闭包就起 goroutine”
- 新版 `WorkerPool` 应该是“投递一个标准 Job，由预启动 worker 执行”

### 4. TaskRegistry / TaskExecutor

职责：
- 按任务类型找到真正的执行器

建议接口：

```go
type TaskExecutor interface {
    Execute(ctx context.Context, task *models.AsyncTask) error
}
```

注册表：

```go
type TaskRegistry interface {
    Register(taskType string, executor TaskExecutor)
    Get(taskType string) (TaskExecutor, bool)
}
```

这样图片、视频、分镜、角色提取都能接进统一入口。

---

## 任务状态机设计

当前状态：
- `pending`
- `processing`
- `completed`
- `failed`

建议扩展为：
- `pending`
- `queued`
- `processing`
- `completed`
- `failed`
- `retry_wait`
- `canceled`

状态意义：

- `pending`
  - DB 已创建，尚未投递 MQ
- `queued`
  - 已成功投递 MQ，等待消费
- `processing`
  - worker 已领取并开始执行
- `retry_wait`
  - 本次执行失败但允许重试，等待再次投递
- `failed`
  - 超过最大重试次数或不可恢复错误
- `completed`
  - 成功
- `canceled`
  - 手动取消或资源已失效

不要让 MQ 自己决定任务最终状态。MQ 只负责“投递/重投/死信”，最终状态以 DB 更新为准。

---

## 幂等与重试设计

这是这套方案的关键。

### 幂等原则

同一个 `task_id` 被重复消费时，执行器必须能安全返回。

建议：
- 执行前先读取 `async_tasks.status`
- 如果已是 `completed/failed/canceled`，直接跳过并 `ack`
- 如果是 `processing`，需要基于“租约超时”判断是否被别的 worker 卡死

### 重试原则

区分两类错误：

- **可重试**
  - AI 上游超时
  - MQ 短暂断连
  - 临时网络错误
  - 对象存储临时失败
- **不可重试**
  - 参数错误
  - 数据不存在
  - 用户积分不足
  - 资源状态不合法

建议：
- 在 `async_tasks` 增加：
  - `retry_count`
  - `max_retries`
  - `next_retry_at`
  - `worker_id`
  - `leased_until`
- 可重试错误进入 `retry_wait`
- Dispatcher 或定时调度器负责将到期任务重新入队

---

## 建议的 MQ 主题与队列划分

不要一开始就做一堆复杂主题。按“资源成本”和“任务时长”分三类就够了。

### 建议队列

- `task.image`
  - 图片生成、背景提取、角色图生成、道具图生成
- `task.video`
  - 视频生成、视频轮询、视频合并
- `task.script`
  - 分镜生成、角色提取、道具提取、帧提示词生成

### 原因

- 图片/视频任务对 GPU / 外部 API / 轮询时长的占用差异大
- 分开后可独立配置 worker 数量和 prefetch
- 避免一个视频任务把整个线程池拖死

### 单队列什么时候才做

只有在第一阶段验证系统链路时，才可以先用一个通用队列 `task.default`。  
一旦准备上线正式生产，建议至少拆成上面三类。

---

## 线程池设计建议

### 不建议继续沿用当前 TaskRunner

当前实现：
- `Submit -> go func -> semaphore`

问题：
- 任务生命周期不可观测
- 无关闭机制
- 无上下文
- 无队列长度指标

### 建议实现

```go
type Job struct {
    TaskID   string
    TaskType string
}

type WorkerPool struct {
    jobs chan Job
    wg   sync.WaitGroup
}
```

能力要求：
- 固定 worker 数
- `Start() / Stop(ctx)` 生命周期
- 停机时停止继续领取新消息
- 正在执行的任务允许超时收尾

### 并发建议

首期默认值：
- `script`: 4
- `image`: 4
- `video`: 2

原因：
- 文本类主要吃外部 API 和解析
- 图片类任务有轮询和 I/O
- 视频类任务最重，先保守

并发数必须配置化，不要写死在服务里。

---

## 与当前代码的映射改造

### 当前最需要改的文件

**Create:**
- `application/services/task_dispatcher.go`
- `application/services/task_consumer.go`
- `application/services/worker_pool.go`
- `application/services/task_registry.go`
- `application/services/task_executor_image.go`
- `application/services/task_executor_video.go`
- `application/services/task_executor_storyboard.go`
- `infrastructure/mq/rabbitmq.go`

**Modify:**
- `application/services/task_runner.go`
- `application/services/task_service.go`
- `application/services/image_generation_service.go`
- `application/services/video_generation_service.go`
- `application/services/storyboard_service.go`
- `application/services/script_generation_service.go`
- `application/services/character_library_service.go`
- `application/services/prop_service.go`
- `application/services/frame_prompt_service.go`
- `api/routes/dependencies.go`
- `main.go`
- `domain/models/task.go`
- `infrastructure/database/database.go`

### 改造原则

原来：

```go
task, created, err := s.taskService.CreateOrGetActiveTask(...)
if created {
    s.runner.Submit("image.process_generation", func() {
        s.ProcessImageGeneration(imageGen.ID)
    })
}
```

改成：

```go
task, created, err := s.taskService.CreateOrGetActiveTask(...)
if created {
    _ = s.taskService.MarkQueued(task.ID)
    err = s.dispatcher.Dispatch(task)
}
```

真正执行逻辑从 Service 的 API 入口挪到 `TaskExecutor`。

---

## 推荐实施顺序

### Phase 1：先抽象，不切 MQ

目标：
- 把 `runner.Submit(fn)` 迁移成 `dispatcher.Dispatch(task)`
- 但 Dispatcher 先做本地内存实现，方便低风险收口调用点

这样做的好处：
- 先把任务入口统一
- 不把“架构重构”和“基础设施切换”绑死在一起

### Phase 2：接 RabbitMQ

目标：
- 增加 RabbitMQ Dispatcher 和 Consumer
- 将本地 Dispatcher 切换成 MQ Dispatcher
- 保留一个开关，允许回退到本地模式

### Phase 3：引入独立 worker 进程

目标：
- Web API 与 Worker 进程分离部署
- API 只负责写任务和查状态
- Worker 专门负责消费任务

这一步完成后，扩容逻辑才真正清晰。

---

## 配置设计

建议新增：

```yaml
task:
  mode: "mq" # local | mq
  worker_id: "auto-generated"
  queues:
    script:
      concurrency: 4
      prefetch: 4
    image:
      concurrency: 4
      prefetch: 4
    video:
      concurrency: 2
      prefetch: 2

mq:
  provider: "rabbitmq"
  url: "amqp://user:pass@rabbitmq:5672/"
  exchange: "xinggen.tasks"
  dlx_exchange: "xinggen.tasks.dlx"
```

注意：
- `task.mode` 要允许本地回退
- `worker_id` 要写入日志和 DB，便于定位卡死任务

---

## 监控与运维建议

必须至少暴露这些指标：

- 每类队列积压数
- 每类任务处理耗时
- 每类任务失败率
- 重试次数分布
- worker 活跃数
- 线程池队列长度
- `async_tasks` 中 `processing` 超时任务数量

日志里至少要带：
- `task_id`
- `task_type`
- `resource_id`
- `worker_id`
- `retry_count`

否则这套系统一旦线上卡住，很难查。

---

## 风险与规避

### 风险 1：消息重复消费

规避：
- DB 状态检查 + 幂等执行
- 以 `task_id` 为唯一执行单元

### 风险 2：任务处理中实例宕机

规避：
- `leased_until`
- 启动时扫描超时 `processing` 任务重新入队

### 风险 3：任务状态和 MQ 状态不一致

规避：
- 先写 DB，再投 MQ
- 投递失败时任务保持 `pending` 或回写错误
- 提供补偿扫描器，把 `pending/retry_wait` 任务重新投递

### 风险 4：长任务占满线程池

规避：
- 按队列拆并发
- 视频任务与文本/图片分池
- 轮询类任务尽量改为“短执行 + 延迟重投”，不要长期占用 worker

---

## 实施任务拆分

### Task 1: 收口任务提交入口

**Files:**
- Create: `application/services/task_dispatcher.go`
- Modify: `application/services/image_generation_service.go`
- Modify: `application/services/video_generation_service.go`
- Modify: `application/services/storyboard_service.go`
- Modify: `application/services/script_generation_service.go`
- Modify: `application/services/character_library_service.go`
- Modify: `application/services/prop_service.go`
- Modify: `application/services/frame_prompt_service.go`

**目标：**
- 不再在业务服务里直接依赖 `TaskRunner.Submit`
- 所有任务提交统一走 `TaskDispatcher`

### Task 2: 扩展任务表与状态机

**Files:**
- Modify: `domain/models/task.go`
- Modify: `application/services/task_service.go`
- Modify: `application/services/task_service_test.go`
- Modify: `infrastructure/database/database.go`

**目标：**
- 新增 `queued/retry_wait/canceled`
- 增加 `retry_count/max_retries/next_retry_at/worker_id/leased_until`

### Task 3: 引入注册式执行器

**Files:**
- Create: `application/services/task_registry.go`
- Create: `application/services/task_executor_*.go`
- Modify: `api/routes/dependencies.go`

**目标：**
- 各任务类型统一注册
- Worker 只认 `task_type`

### Task 4: 落地本地 WorkerPool

**Files:**
- Create: `application/services/worker_pool.go`
- Modify: `application/services/task_runner.go`
- Create: `application/services/worker_pool_test.go`

**目标：**
- 用标准 worker pool 替换当前轻量 semaphore runner

### Task 5: 接 RabbitMQ Dispatcher / Consumer

**Files:**
- Create: `infrastructure/mq/rabbitmq.go`
- Create: `application/services/task_consumer.go`
- Create: `application/services/task_dispatcher_rabbitmq.go`
- Modify: `main.go`

**目标：**
- 打通 MQ 投递和消费

### Task 6: 增加失败恢复与补偿调度

**Files:**
- Create: `application/services/task_recovery_service.go`
- Modify: `application/services/task_service.go`
- Modify: `main.go`

**目标：**
- 自动扫描卡死任务
- 自动重新投递 `pending/retry_wait`

### Task 7: API 与运维回归验证

**Files:**
- Modify: `docs/PROJECT_ARCHITECTURE.md`
- Modify: `docs/saas-backend-apis.md`
- Create: `docs/plans/2026-03-19-mq-threadpool-task-execution.md`

**目标：**
- 文档化
- 验证队列、重试、并发和恢复链路

---

## 最终建议

如果你现在就要开始做，我建议按这个顺序落地：

1. **先统一提交入口，不立刻接 MQ**
   - 先把所有 `runner.Submit` 收口到 `dispatcher.Dispatch`
2. **再扩展 `AsyncTask` 状态机**
   - 把幂等、重试、租约字段补齐
3. **然后接 RabbitMQ**
   - 先一个默认队列跑通，再拆 `script/image/video`
4. **最后再拆独立 worker 进程**
   - 这一步才做真正的横向扩容

这是最稳的路线。  
如果你一上来就“RabbitMQ、线程池、状态机、独立 worker”一起改，失败面会非常大，而且很难定位。

---

Plan complete and saved to `docs/plans/2026-03-19-mq-threadpool-task-execution.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
