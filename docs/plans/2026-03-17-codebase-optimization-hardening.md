# 前后端稳定性与架构优化计划书 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 先修复当前前端构建失稳和类型漂移问题，再逐步降低前后端耦合、收敛异步任务模型，并提升媒体处理链路的可维护性与稳定性。

**Architecture:** 采用“先止血、再解耦、后提吞吐”的顺序推进。第一阶段先恢复 `web` 的类型与构建可信度，第二阶段把请求鉴权、路由装配、服务依赖和超大组件拆成清晰边界，第三阶段再把后台 goroutine、FFmpeg 和外部 HTTP 调用收敛到可取消、可限流、可观测的执行模型。

**Tech Stack:** Go, Gin, GORM, Vue 3, TypeScript, Vite, Pinia, Element Plus, FFmpeg

---

## 问题总览与优化建议

| 优先级 | 问题 | 现状证据 | 优化建议 |
| --- | --- | --- | --- |
| P0 | 前端类型契约漂移 | `npm run build:check` 失败；`StoryboardEditor.vue` 使用 `composed_url/background_id`，而 `types/drama.ts` 中未定义对应字段 | 先统一 API 响应模型、编辑器组件和 `types/*`，把类型系统重新变成可信边界 |
| P0 | 请求层与鉴权状态重复实现 | `web/src/utils/request.ts` 和 `web/src/stores/auth.ts` 各自维护刷新逻辑 | 将 token 刷新、登出和重定向统一收口到 store/session service，请求层只保留 transport 逻辑 |
| P1 | 超大单文件组件难维护 | `ProfessionalEditor.vue` 6404 行，`EpisodeWorkflow.vue` 3643 行，`VideoTimelineEditor.vue` 2493 行 | 按“状态管理 / API 映射 / 视图组件 / 命令操作”拆成 composable + 子组件 |
| P1 | 前端启动包存在冗余 | `main.ts` 全量注册 `@element-plus/icons-vue`，但多数页面已经本地按需 import | 移除全局图标注册，保持本地按需引入，顺便补齐 `vite-env.d.ts` |
| P1 | 后端异步任务模型无生命周期控制 | 多处直接 `go func` 或 `go s.Process...`，缺少取消、限流和统一错误处理 | 抽统一任务执行器，使用 `context`、并发上限和显式状态机管理后台任务 |
| P1 | 媒体处理链路缺少 timeout 和 context | `ffmpeg.go`、`gemini_image_client.go`、`image_utils.go`、`openai_sora_client.go` 使用裸 `http.Get`；FFmpeg 使用裸 `exec.Command` | 统一到带超时的 `http.Client` 和 `exec.CommandContext`，为下载、转码、轮询增加取消能力 |
| P2 | 路由装配层耦合过重 | `api/routes/routes.go` 同时负责构造服务、构造 handler、注册路由，并对 `localStorage` 做强转 | 引入依赖容器/装配层，把“创建依赖”和“注册路由”拆开，避免配置切换时 panic |
| P2 | 服务职责过粗 | `image_generation_service.go` 1223 行，`storyboard_service.go` 1182 行，`ai_service.go` 733 行 | 将 prompt 构建、账单扣费、provider 适配、任务编排、结果落库拆成独立协作对象 |

## 约束与执行顺序

- 必须先做 Task 1 和 Task 2，再做任何前端功能修改，否则构建结果不可用，改动反馈会持续失真。
- Task 5 和 Task 6 涉及后台执行模型，开始前必须先完成 Task 7 的依赖装配收敛，否则任务执行器仍会被隐式依赖污染。
- 所有任务结束前，必须执行 `@verification-before-completion` 风格的完整验证，不能用“应该可以”替代真实命令输出。
- Go 验证命令统一从 WSL 执行；前端验证命令统一从 Windows PowerShell 执行。

### Task 1: 修复前端类型契约与构建基线

**Files:**
- Create: `web/src/vite-env.d.ts`
- Modify: `web/src/types/drama.ts`
- Modify: `web/src/types/image.ts`
- Modify: `web/src/types/timeline.ts`
- Modify: `web/src/types/video.ts`
- Modify: `web/src/api/drama.ts`
- Modify: `web/src/api/character-library.ts`
- Modify: `web/src/components/editor/StoryboardEditor.vue`
- Modify: `web/src/components/editor/VideoTimelineEditor.vue`
- Modify: `web/src/views/drama/DramaManagement.vue`
- Modify: `web/src/views/drama/DramaWorkflow.vue`
- Modify: `web/src/views/drama/EpisodeWorkflow.vue`
- Modify: `web/src/views/drama/ProfessionalEditor.vue`
- Modify: `web/src/views/drama/components/UploadScriptDialog.vue`
- Modify: `web/src/views/editor/TimelineEditor.vue`

**Step 1: 固化失败基线**

Run: `powershell.exe -Command "cd D:\\code\\github\\huobao-drama\\web; npm run build:check"`
Expected: FAIL，且输出包含 `composed_url`、`background_id`、`import.meta.env`、`TimelineClip` 等类型错误

**Step 2: 先补 Vite 环境类型**

- 新增 `web/src/vite-env.d.ts`
- 写入 `/// <reference types="vite/client" />`

**Step 3: 对齐核心领域类型**

- 统一 `Storyboard`、`Scene`、`ImageGeneration`、`TimelineClip` 的字段命名和 ID 类型
- 明确哪些字段是后端真实返回值，哪些字段仅是前端派生字段
- 删除只能靠 `[key: string]: any` 勉强兜底的字段访问

**Step 4: 对齐 API 映射层**

- 把 `web/src/api/*.ts` 中返回值映射为显式模型
- 避免页面直接猜测接口字段
- 对 `parseScript`、`generateShots` 一类已经漂移的 API 调用做显式修正或删除

**Step 5: 修复编辑器页面的类型误用**

- 用真实字段替换 `composed_url` / `background_id` / 错误的 `scene_id` 推断
- 清理字符串与数字 ID 混用
- 让 `vue-tsc` 在这些文件上不再依赖隐式 `any`

**Step 6: 重新验证前端基线**

Run: `powershell.exe -Command "cd D:\\code\\github\\huobao-drama\\web; npm run build:check"`
Expected: 类型错误显著减少；如果仍失败，错误应集中到尚未进入本任务处理范围的文件

### Task 2: 收敛请求层与鉴权状态管理

**Files:**
- Create: `web/src/services/session.ts`
- Modify: `web/src/utils/request.ts`
- Modify: `web/src/stores/auth.ts`
- Modify: `web/src/stores/adminAuth.ts`
- Modify: `web/src/router/index.ts`

**Step 1: 为会话状态定义单一职责入口**

- 在 `web/src/services/session.ts` 中定义用户和管理员 token 的读取、刷新、清理和重定向策略
- 避免 `request.ts`、`router.ts`、store 三处重复拼接登录跳转

**Step 2: 缩减请求层职责**

- `request.ts` 只处理 `Authorization` 注入、响应解包和 401 钩子
- 401 之后调用 session service，不直接操作 `window.location`
- 移除导致 `CustomAxiosInstance` 失真的强制断言，让实例保留可调用签名

**Step 3: 缩减 store 职责**

- `auth.ts` 和 `adminAuth.ts` 只保留状态变更与业务动作
- 删除与请求层重复的 refreshPromise 管理
- 统一 `logout` 后的本地存储清理规则

**Step 4: 验证请求层重构**

Run: `powershell.exe -Command "cd D:\\code\\github\\huobao-drama\\web; npx vue-tsc --noEmit --skipLibCheck"`
Expected: 不再出现 `request.ts` 中的 `AxiosRequestHeaders` 和“instance 不可调用”错误

### Task 3: 拆分超大编辑器与工作流组件

**Files:**
- Create: `web/src/composables/editor/useProfessionalEditor.ts`
- Create: `web/src/composables/editor/useEpisodeWorkflow.ts`
- Create: `web/src/composables/editor/useTimelineMerge.ts`
- Create: `web/src/components/editor/professional/`
- Create: `web/src/components/editor/episode/`
- Modify: `web/src/views/drama/ProfessionalEditor.vue`
- Modify: `web/src/views/drama/EpisodeWorkflow.vue`
- Modify: `web/src/components/editor/VideoTimelineEditor.vue`
- Modify: `web/src/components/editor/StoryboardEditor.vue`

**Step 1: 先划分模块边界**

- `ProfessionalEditor.vue` 至少拆为：数据装载、镜头列表、素材面板、生成控制、预览区
- `EpisodeWorkflow.vue` 至少拆为：步骤导航、角色步骤、场景步骤、分镜步骤
- `VideoTimelineEditor.vue` 至少拆为：轨道区、属性面板、导出区

**Step 2: 先抽 composable，再抽视图**

- 优先把副作用、watch、轮询、复杂计算迁移到 composable
- 页面层保留 props、emit 和模板拼装
- 避免一开始就把模板和状态同时大规模重排

**Step 3: 为拆分后的状态建立显式输入输出**

- 每个 composable 返回明确的 state、actions、derived values
- 不允许子组件直接写全局 store 的深层状态

**Step 4: 验证拆分没有破坏编译**

Run: `powershell.exe -Command "cd D:\\code\\github\\huobao-drama\\web; npm run build:check"`
Expected: 构建继续通过；若失败，应仅剩新拆分文件的局部错误

### Task 4: 收紧前端启动与 UI 依赖注册

**Files:**
- Modify: `web/src/main.ts`
- Modify: `web/src/components/common/*.vue`
- Modify: `web/src/components/editor/*.vue`
- Modify: `web/src/views/**/*.vue`

**Step 1: 删除冗余全局图标注册**

- 移除 `main.ts` 中对 `@element-plus/icons-vue` 的全量 `Object.entries` 注册
- 保持各页面按需 import 本地已使用的图标组件

**Step 2: 扫描遗漏的模板图标**

- 对所有仍依赖全局注册的组件补本地 import
- 确保模板中不再依赖运行时全局图标兜底

**Step 3: 验证启动包改动**

Run: `powershell.exe -Command "cd D:\\code\\github\\huobao-drama\\web; npm run build"`
Expected: PASS，且不再依赖全局图标注册

### Task 5: 为后台任务引入统一执行器

**Files:**
- Create: `application/services/task_runner.go`
- Modify: `application/services/task_service.go`
- Modify: `application/services/image_generation_service.go`
- Modify: `application/services/video_generation_service.go`
- Modify: `application/services/character_library_service.go`
- Modify: `application/services/storyboard_service.go`
- Modify: `application/services/script_generation_service.go`

**Step 1: 先写任务执行器测试**

- 测试并发上限
- 测试任务取消
- 测试任务 panic/错误会正确回写状态

**Step 2: 固化当前任务状态机**

- 明确 `pending -> processing -> completed/failed`
- 统一 `progress`、`message`、`result` 的更新入口

**Step 3: 用执行器替换裸 goroutine**

- 把 `go s.ProcessImageGeneration(...)`、批量角色图生成、分镜并发生成迁移到统一执行器
- 为长任务显式传递 `context.Context`
- 允许后续接入更正规的持久化队列而不重写业务层

**Step 4: 验证后台任务重构**

Run: `wsl bash -lic 'cd /mnt/d/code/github/huobao-drama && go test ./application/services -run TestTask -v'`
Expected: PASS；新增任务执行器相关测试通过

### Task 6: 加固媒体处理与外部 HTTP 调用

**Files:**
- Create: `pkg/httpclient/client.go`
- Modify: `infrastructure/external/ffmpeg/ffmpeg.go`
- Modify: `pkg/image/gemini_image_client.go`
- Modify: `pkg/utils/image_utils.go`
- Modify: `pkg/video/openai_sora_client.go`
- Modify: `pkg/video/volces_ark_client.go`
- Modify: `pkg/video/chatfire_client.go`

**Step 1: 抽统一 HTTP client**

- 提供默认 timeout、User-Agent、重试策略和上下文取消能力
- 禁止业务包继续直接使用裸 `http.Get`

**Step 2: 给 FFmpeg 命令接入 context**

- 将 `exec.Command` 改为 `exec.CommandContext`
- 为下载、裁剪、合成、探测增加超时和错误分类
- 为临时文件清理增加 `defer`/失败路径一致性

**Step 3: 限制媒体链路的串行瓶颈**

- 对下载与转码做有限并发
- 保留顺序合成，但把 I/O 密集步骤从完全串行改为可控并发

**Step 4: 验证媒体链路改造**

Run: `wsl bash -lic 'cd /mnt/d/code/github/huobao-drama && go test ./infrastructure/external/ffmpeg ./pkg/image ./pkg/video ./pkg/utils -v'`
Expected: PASS；如存在集成测试依赖外部环境，需单独记录并隔离

### Task 7: 拆分路由装配与依赖构造

**Files:**
- Create: `api/routes/dependencies.go`
- Modify: `api/routes/routes.go`
- Modify: `main.go`
- Modify: `api/handlers/*.go`

**Step 1: 定义依赖容器**

- 在 `api/routes/dependencies.go` 中集中构造 service、repository、handler
- 将 `SetupRouter` 限制为“接收依赖并注册路由”

**Step 2: 移除危险类型断言**

- 替换 `localStorage.(*storage.LocalStorage)` 为显式接口或可选依赖
- 在非 local storage 配置下给出明确错误，而不是 panic

**Step 3: 缩减 handler 构造时的隐藏依赖**

- 避免 handler 在内部再次 `NewService`
- 让依赖关系能从入口一眼看清

**Step 4: 验证启动装配**

Run: `wsl bash -lic 'cd /mnt/d/code/github/huobao-drama && go test ./api/... ./application/services/... -v'`
Expected: PASS；至少覆盖被重构的 handler 和 middleware 相关测试

### Task 8: 拆解超大后端服务边界

**Files:**
- Create: `application/services/image_generation/`
- Create: `application/services/storyboard/`
- Create: `application/services/video_generation/`
- Modify: `application/services/image_generation_service.go`
- Modify: `application/services/storyboard_service.go`
- Modify: `application/services/video_generation_service.go`
- Modify: `application/services/ai_service.go`

**Step 1: 先按职责切包，不先追求目录完美**

- `prompt builder`
- `provider selector`
- `billing usage recorder`
- `task orchestration`
- `result persistence`

**Step 2: 先抽纯函数和协作者**

- 把易测的文本拼装、参数归一化、状态决策先抽出来
- 保留原 service 作为 facade，减少一次性大改风险

**Step 3: 为抽出的纯函数补测试**

- prompt 构建
- 参数标准化
- 状态机分支
- provider 选择优先级

**Step 4: 验证服务拆分**

Run: `wsl bash -lic 'cd /mnt/d/code/github/huobao-drama && go test ./application/services -v'`
Expected: PASS；新增测试覆盖率应高于拆分前

### Task 9: 更新文档并完成最终验收

**Files:**
- Modify: `docs/PROJECT_ARCHITECTURE.md`
- Modify: `README-CN.md`
- Modify: `README.md`
- Review: `docs/plans/2026-03-17-codebase-optimization-hardening.md`

**Step 1: 更新架构文档**

- 补充前端 session/request 分层
- 补充后台任务执行器
- 补充媒体链路 timeout/context 规则

**Step 2: 更新开发文档**

- 记录前端必须运行的构建校验命令
- 记录 Go 在 Windows + WSL 下的推荐验证方式
- 记录异步任务新增约束：禁止直接起裸 goroutine

**Step 3: 执行最终验证**

Run: `powershell.exe -Command "cd D:\\code\\github\\huobao-drama\\web; npm run build:check"`
Expected: PASS

Run: `wsl bash -lic 'cd /mnt/d/code/github/huobao-drama && go test ./... -v'`
Expected: PASS；若环境缺少依赖，必须明确列出阻塞项和未验证范围

Run: `powershell.exe -Command "cd D:\\code\\github\\huobao-drama; git diff --stat"`
Expected: 变更范围与计划一致，无意外文件漂移

## 实施建议

- 建议先只执行 Task 1 到 Task 2，恢复前端构建可信度后再继续后面的结构优化。
- Task 5 到 Task 8 都属于“收益高但改动面大”的任务，适合分多个短分支推进，不建议一次性混在同一个提交里。
- 每完成一个 Task，都单独做一次最小验证并提交，避免把“修类型”和“改架构”搅在一起。

Plan complete and saved to `docs/plans/2026-03-17-codebase-optimization-hardening.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
