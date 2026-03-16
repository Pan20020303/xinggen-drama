# Doubao Seed Model Alias Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在后端完整支持 `doubao-seed1.8`、`seedream4.5`、`seedance1.5pro` 三个短模型名，并统一映射到可用的 canonical 模型 ID。

**Architecture:** 新增一个模型别名归一化模块，作为唯一映射入口；服务层在配置写入、模型匹配、计费模型解析处统一调用；图像/视频客户端在请求下发前再做一次归一化兜底，确保历史配置与新请求都可运行。全流程采用 TDD（先红后绿）。

**Tech Stack:** Go 1.23, Gin, GORM, SQLite(test), existing `pkg/image` + `pkg/video` clients.

---

### Task 1: 建立模型别名规范与归一化单元测试（RED）

**Files:**
- Create: `pkg/modelalias/model_alias_test.go`

**Step 1: 写失败测试（别名 -> canonical）**

```go
func TestNormalize_DoubaoAliasModels(t *testing.T) {
  // text: doubao-seed1.8 -> doubao-seed-1-8-251228
  // image: seedream4.5 -> doubao-seedream-4-5-251128
  // video: seedance1.5pro -> doubao-seedance-1-5-pro-251215
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./pkg/modelalias -v`  
Expected: FAIL，报 `undefined: Normalize` 或断言失败（因为还未实现）

**Step 3: 提交（仅测试）**

```bash
git add pkg/modelalias/model_alias_test.go
git commit -m "test: add failing tests for doubao model alias normalization"
```

---

### Task 2: 实现统一别名归一化模块（GREEN）

**Files:**
- Create: `pkg/modelalias/model_alias.go`
- Modify: `pkg/modelalias/model_alias_test.go`

**Step 1: 实现最小代码**

```go
func Normalize(serviceType, model string) string
func NormalizeAny(model string) string
```

最小要求：
- `text`: `doubao-seed1.8` 映射到 `doubao-seed-1-8-251228`
- `image`: `seedream4.5` 映射到 `doubao-seedream-4-5-251128`
- `video`: `seedance1.5pro` 映射到 `doubao-seedance-1-5-pro-251215`
- 未命中别名时原样返回

**Step 2: 运行测试确认通过**

Run: `go test ./pkg/modelalias -v`  
Expected: PASS

**Step 3: 格式化与提交**

```bash
gofmt -w pkg/modelalias/model_alias.go pkg/modelalias/model_alias_test.go
git add pkg/modelalias/model_alias.go pkg/modelalias/model_alias_test.go
git commit -m "feat: add canonical model alias normalizer for doubao seed models"
```

---

### Task 3: 服务层接入归一化（配置写入+模型匹配+计费解析）

**Files:**
- Modify: `application/services/ai_service.go`
- Modify: `application/services/ai_service_platform_test.go`

**Step 1: 先写失败测试（服务层别名匹配）**

在 `ai_service_platform_test.go` 增加：
- seed canonical 配置（text/image/video 各一条）
- 用短名调用 `GetConfigForModel`
- 断言能命中对应配置

**Step 2: 运行测试确认失败**

Run: `go test ./application/services -run TestGetConfigForModel_ResolvesDoubaoAliases -v`  
Expected: FAIL，提示找不到 alias 对应模型

**Step 3: 实现最小改动**

在 `ai_service.go` 接入：
- `CreateConfig` / `UpdateConfig` / `UpdatePlatformConfig`：写入前归一化 model
- `GetConfigForModel`：比较时对请求 model 与配置 model 都归一化
- `GetBillingConfig`：返回 `actualModel` 前归一化
- `GetAIClientForModel` / `GetAIClientForModelWithUser` / `TestConnection`：输入模型先归一化

**Step 4: 运行测试确认通过**

Run: `go test ./application/services -run TestGetConfigForModel_ResolvesDoubaoAliases -v`  
Expected: PASS

**Step 5: 回归关键测试**

Run: `go test ./application/services -run TestGetBillingConfig_FallbacksToPositivePlatformPrice -v`  
Expected: PASS（验证无回归）

**Step 6: 格式化与提交**

```bash
gofmt -w application/services/ai_service.go application/services/ai_service_platform_test.go
git add application/services/ai_service.go application/services/ai_service_platform_test.go
git commit -m "feat: normalize doubao model aliases in ai service resolution path"
```

---

### Task 4: 客户端层兜底（图像/视频）

**Files:**
- Modify: `pkg/image/volcengine_image_client.go`
- Modify: `pkg/video/volces_ark_client.go`
- Modify: `pkg/video/chatfire_client.go`（仅在需要时）

**Step 1: 先写失败测试**

建议新增测试：
- `VolcEngineImageClient` 在 `seedream4.5` 短名时，实际请求模型为 canonical，默认 size 按 Seedream 4.5 走 `2K`
- `VolcesArkClient` 在 `seedance1.5pro` 短名时，实际请求模型为 canonical，任务类型逻辑不变（`i2v/t2v`）

**Step 2: 运行失败测试**

Run: `go test ./pkg/image ./pkg/video -run "TestVolcEngine.*Alias|TestVolcesArk.*Alias" -v`  
Expected: FAIL

**Step 3: 实现最小代码**

- 图像客户端请求前调用 `modelalias.Normalize("image", model)`
- 视频客户端 `normalizeSeedance15ProModel` 统一基于 `modelalias.Normalize("video", model)`

**Step 4: 运行测试确认通过**

Run: `go test ./pkg/image ./pkg/video -run "TestVolcEngine.*Alias|TestVolcesArk.*Alias" -v`  
Expected: PASS

**Step 5: 格式化与提交**

```bash
gofmt -w pkg/image/volcengine_image_client.go pkg/video/volces_ark_client.go pkg/video/chatfire_client.go
git add pkg/image/volcengine_image_client.go pkg/video/volces_ark_client.go pkg/video/chatfire_client.go
git commit -m "feat: normalize seedream and seedance aliases in provider clients"
```

---

### Task 5: 全量验证与交付说明

**Files:**
- Modify: `docs/saas-backend-apis.md`（新增模型名说明，可选）

**Step 1: 跑核心测试集**

Run:
- `go test ./application/services -v`
- `go test ./pkg/... -v`

Expected: PASS

**Step 2: 输出兼容性说明**

需明确：
- 旧 canonical 名可继续使用
- 新短名与 canonical 行为一致
- billing/transaction 记录中的 `model` 统一为 canonical（避免多别名统计分裂）

**Step 3: 提交**

```bash
git add docs/saas-backend-apis.md
git commit -m "docs: document doubao seed model alias compatibility"
```

