# 模型 Token 消耗埋点与管理后台统计页 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为平台新增模型调用 token 精确埋点，并在管理后台提供按模型聚合的 token 消耗统计页。

**Architecture:** 使用 `credit_transactions` 作为统一计费与消耗审计表，在其中追加 `prompt_tokens`、`completion_tokens`、`total_tokens` 字段。文本/图片/视频客户端暴露统一 usage 能力，业务服务在成功调用后按 `reference_id` 回写 token，管理后台再按模型和服务类型聚合展示。

**Tech Stack:** Go, GORM, Gin, Vue 3, Element Plus, TypeScript

---

### Task 1: 先锁定 token 字段与聚合行为

**Files:**
- Modify: `domain/models/admin_models_test.go`
- Create: `application/services/admin_token_stats_service_test.go`

**Step 1: Write the failing tests**

- `CreditTransaction` 必须包含 token 字段
- 管理端聚合必须按模型求和
- usage 回写必须只影响目标 `reference_id`

**Step 2: Run tests to verify they fail**

Run: `wsl bash -lic 'cd /mnt/d/code/github/huobao-drama && go test ./domain/models ./application/services -v'`
Expected: FAIL before implementation

### Task 2: 扩展流水模型与后端服务

**Files:**
- Modify: `domain/models/user.go`
- Modify: `application/services/billing_service.go`
- Create: `application/services/admin_token_stats_service.go`
- Create: `pkg/usage/usage.go`

**Step 1: Implement the minimal backend**

- 增加 token 字段
- 增加 usage 回写方法
- 增加按模型聚合服务

**Step 2: Re-run tests**

Run: `wsl bash -lic 'cd /mnt/d/code/github/huobao-drama && go test ./domain/models ./application/services -v'`
Expected: targeted token tests PASS

### Task 3: 给客户端和业务链路接入 usage 回写

**Files:**
- Modify: `pkg/ai/client.go`
- Modify: `pkg/ai/openai_client.go`
- Modify: `pkg/ai/gemini_client.go`
- Modify: `pkg/image/image_client.go`
- Modify: `pkg/image/openai_image_client.go`
- Modify: `pkg/image/gemini_image_client.go`
- Modify: `pkg/image/volcengine_image_client.go`
- Modify: `pkg/video/video_client.go`
- Modify: `pkg/video/chatfire_client.go`
- Modify: `pkg/video/minimax_client.go`
- Modify: `pkg/video/openai_sora_client.go`
- Modify: `pkg/video/volces_ark_client.go`
- Modify: `application/services/*.go` token 调用点

**Step 1: Implement**

- 文本/图片/视频客户端增加 `GetLastUsage`
- 成功调用后按 `reference_id` 回写 token

### Task 4: 新增管理端统计接口与页面

**Files:**
- Modify: `api/handlers/admin_billing.go`
- Modify: `api/routes/routes.go`
- Modify: `web/src/router/index.ts`
- Modify: `web/src/api/admin.ts`
- Modify: `web/src/types/admin.ts`
- Create: `web/src/views/admin/AdminTokenStats.vue`
- Modify: `web/src/views/admin/AdminBilling.vue`
- Modify: `web/src/views/admin/AdminAIConfig.vue`

**Step 1: Implement**

- 新增 `/api/v1/admin/billing/token-stats`
- 新增管理后台 Token 统计页

### Task 5: 最终验证

**Step 1: Run backend targeted tests**

Run: `wsl bash -lic 'cd /mnt/d/code/github/huobao-drama && go test ./domain/models ./application/services ./api/handlers -v'`
Expected: token-related tests PASS; unrelated historical failures must be called out separately

**Step 2: Run frontend build**

Run: `powershell.exe -Command "cd D:\\code\\github\\huobao-drama\\web; cmd /c npm run build"`
Expected: PASS
