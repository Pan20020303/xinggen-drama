# Admin-Only AI Config + Usage Billing Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Move all AI model/provider configuration to platform admin only, remove user-side AI configuration, and enforce credit consumption for every user-triggered “generation/AI usage” action with a clear, configurable billing strategy.

**Architecture:** Treat AI service configs as platform-level records (`user_id=0`) managed only under `/api/v1/admin/*`. All AI calls for end users must resolve configs only from platform records. Billing is enforced at the service layer before creating async tasks, within DB transactions that also write credit transactions.

**Tech Stack:** Go (Gin, GORM, JWT), Vue3 (Element Plus, Pinia, axios), SQLite/MySQL, existing response/config/logger modules.

---

## Non-Goals

- No migration of existing user-owned AI configs to platform configs (user chose: **discard user configs**).
- No refunds for async AI failures after task creation (baseline policy).
- No complex idempotency keys in v1 (baseline is UI throttling + optional future enhancement).

---

## Billing Policy (Baseline)

**Principles**

- Charge credits for any operation that triggers AI inference or creates an async generation task.
- Charge **before** task creation, within the same DB transaction as:
  - user credit decrement
  - credit transaction insert
  - generation/task record creation (or whatever record the endpoint persists)
- If the request fails before task creation, the transaction rolls back, so no charge.
- If task is created and later AI fails, **no refund** in baseline.

**Chargeable actions (recommended minimum)**

- Script generation / extraction flows that call AI (script, character extraction, scene extraction).
- Storyboard generation.
- Frame prompt generation (already charged).
- Image generation (already charged).
- Video generation.
- Video merge/finalize (optional: can be 0 cost initially).

**Configuration**

Add to `billing` config:

- `script_generation_credits`
- `storyboard_generation_credits`
- `video_generation_credits`
- `video_merge_credits`
- Keep existing:
  - `frame_prompt_credits`
  - `image_generation_credits`

If a cost is `<= 0`, treat as “free”.

---

### Task 1: Lock AI Config Resolution To Platform Only

**Files:**

- Modify: `application/services/ai_service.go`
- Test: `application/services/ai_service_platform_test.go`

**Step 1: Write the failing test**

Create tests that assert:

- `GetDefaultConfig(serviceType)` returns only records where `user_id=0`
- `GetConfigForModel(serviceType, modelName)` returns only records where `user_id=0`
- User-owned configs (`user_id != 0`) are ignored even if they have higher `priority`.

**Step 2: Run test to verify it fails**

Run: `go test ./application/services -run TestPlatformAIConfig -v`
Expected: FAIL because current queries don’t constrain `user_id` when userID is 0.

**Step 3: Write minimal implementation**

- In `GetDefaultConfig` / `GetConfigForModel` / `GetAIClient` / `GetAIClientForModel`, enforce:
  - platform scope only: `user_id = 0`
- Ensure callers that used to omit user ID now still work (platform default exists).

**Step 4: Run test to verify it passes**

Run: `go test ./application/services -run TestPlatformAIConfig -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add application/services/ai_service.go application/services/ai_service_platform_test.go
git commit -m "feat(ai): resolve configs from platform scope only"
```

---

### Task 2: Add Admin-Only AI Config APIs (Masked Secrets)

**Files:**

- Create: `api/handlers/admin_ai_config.go`
- Modify: `api/routes/routes.go`
- Modify: `application/services/ai_service.go`
- Test: `api/handlers/admin_ai_config_test.go`
- Test: `application/services/ai_service_admin_mask_test.go`

**Step 1: Write failing tests**

1) Handler tests:

- `POST /api/v1/admin/ai-configs` creates a platform config (`user_id=0`)
- `GET /api/v1/admin/ai-configs?service_type=text` lists configs
- `PUT /api/v1/admin/ai-configs/:id` updates fields
- `DELETE /api/v1/admin/ai-configs/:id` deletes
- `POST /api/v1/admin/ai-configs/test` tests connectivity

2) Secret-masking tests:

- API responses do not include `api_key` plaintext (expect empty string or omit)
- Include `api_key_set: true|false` to indicate presence

**Step 2: Run tests to verify they fail**

Run:

- `go test ./api/handlers -run TestAdminAIConfig -v`
- `go test ./application/services -run TestAIConfigMask -v`

Expected: FAIL due missing handler/routes and missing masking behavior.

**Step 3: Minimal implementation**

- Add an admin handler similar to `AIConfigHandler`, but:
  - No `tenant.GetUserID` usage
  - Uses platform scope (`user_id=0`)
  - Uses `middlewares.AdminAuthMiddleware`
- Update `AIService` to support an internal “view model”:
  - Persist API key as-is in DB
  - Return masked config to clients (api_key empty; api_key_set boolean)

**Step 4: Run tests to verify they pass**

Run:

- `go test ./api/handlers -run TestAdminAIConfig -v`
- `go test ./application/services -run TestAIConfigMask -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add api/handlers/admin_ai_config.go api/routes/routes.go application/services/ai_service.go api/handlers/admin_ai_config_test.go application/services/ai_service_admin_mask_test.go
git commit -m "feat(admin): manage platform ai configs with masked secrets"
```

---

### Task 3: Disable User-Side AI Config (Backend)

**Files:**

- Modify: `api/routes/routes.go`
- Test: `api/handlers/ai_config_disabled_test.go`

**Step 1: Write failing test**

Add tests that assert user-facing endpoints are not accessible:

- `/api/v1/ai-configs` is not mounted (expect 404 API endpoint not found)
- `/api/v1/ai-configs/test` not accessible by normal user routes

**Step 2: Run test to verify it fails**

Run: `go test ./api/handlers -run TestAIConfigDisabled -v`
Expected: FAIL because routes still exist.

**Step 3: Minimal implementation**

- Remove the `/api/v1/ai-configs` route group mounting from `api/routes/routes.go`.
- Keep internal service logic intact for platform operations.

**Step 4: Run test to verify it passes**

Run: `go test ./api/handlers -run TestAIConfigDisabled -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add api/routes/routes.go api/handlers/ai_config_disabled_test.go
git commit -m "chore(ai): disable user-side ai config APIs"
```

---

### Task 4: Disable User-Side AI Config (Frontend) + Add Admin AI Config Page

**Files:**

- Modify: `web/src/router/index.ts`
- Modify: `web/src/components/common/AppHeader.vue`
- Delete or orphan: `web/src/views/settings/AIConfig.vue` route usage
- Create: `web/src/views/admin/AdminAIConfig.vue`
- Modify: `web/src/views/admin/AdminUsers.vue` (add nav button to AdminAIConfig)
- Modify: `web/src/api/admin.ts` (add admin ai-config endpoints) OR create `web/src/api/adminAiConfig.ts`
- Create/Modify types: `web/src/types/admin.ts`

**Step 1: Write failing build check**

Run: `cd web && npm run build`
Expected: FAIL after removing route/button but before adding admin page and API.

**Step 2: Minimal implementation**

- Remove user route `/settings/ai-config` and hide/remove “AI 配置” button in header.
- Add admin route `/admin/ai-config` and page.
- Reuse UI patterns from `AIConfig.vue` but:
  - call admin endpoints
  - do not display API key after save; show “已设置/未设置”

**Step 3: Verify**

Run: `cd web && npm run build`
Expected: PASS.

**Step 4: Commit**

```bash
git add web/src/router/index.ts web/src/components/common/AppHeader.vue web/src/views/admin/AdminAIConfig.vue web/src/views/admin/AdminUsers.vue web/src/api/admin.ts web/src/types/admin.ts
git commit -m "feat(web-admin): move ai config management to admin"
```

---

### Task 5: Expand Billing Config + Service Costs

**Files:**

- Modify: `pkg/config/config.go`
- Modify: `configs/config.yaml`
- Modify: `application/services/billing_service.go`
- Modify: `domain/models/user.go` (add txn types)
- Test: `application/services/billing_service_test.go`

**Step 1: Write failing test**

Add tests for new consumption methods:

- consumes configured credits and writes transaction
- returns `insufficient credits` when balance low
- no-op when cost <= 0

**Step 2: Run test to verify it fails**

Run: `go test ./application/services -run TestBillingPolicy -v`
Expected: FAIL because methods/config/txn types don’t exist.

**Step 3: Minimal implementation**

- Extend config struct to include new billing fields.
- Extend BillingService with:
  - `ConsumeForScriptGeneration`
  - `ConsumeForStoryboardGeneration`
  - `ConsumeForVideoGeneration`
  - `ConsumeForVideoMerge` (optional, can be cost=0)
- Add txn types in `domain/models/user.go`.

**Step 4: Run test to verify it passes**

Run: `go test ./application/services -run TestBillingPolicy -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/config/config.go configs/config.yaml application/services/billing_service.go domain/models/user.go application/services/billing_service_test.go
git commit -m "feat(billing): add configurable costs for AI usage"
```

---

### Task 6: Enforce Billing At All User AI Usage Entry Points

**Files (expected candidates; verify with search):**

- Modify: `application/services/script_generation_service.go`
- Modify: `application/services/storyboard_service.go`
- Modify: `application/services/video_generation_service.go`
- Modify: `application/services/video_merge_service.go` (if charge)
- Test: `application/services/usage_billing_integration_test.go`

**Step 1: Identify call sites**

Run: `rg "GenerateText|GenerateTextStream|GetAIClient|aiService\\." -n application/services`
Output: list all endpoints that trigger AI calls.

**Step 2: Write failing integration tests**

Test that:

- a user with 0 credits cannot trigger each operation (expects 400/500 with `insufficient credits` depending on handler mapping)
- a user with enough credits can trigger it and credits decrement + txn written

**Step 3: Run tests to verify it fails**

Run: `go test ./application/services -run TestUsageBilling -v`
Expected: FAIL because other flows don’t charge yet.

**Step 4: Minimal implementation**

In each service method that triggers AI/task creation:

- call appropriate `BillingService.ConsumeForXxx(userID, detail)` before creating tasks/records.
- ensure it’s inside the same transaction scope when there are multiple DB writes.

**Step 5: Run tests to verify it passes**

Run: `go test ./application/services -run TestUsageBilling -v`
Expected: PASS.

**Step 6: Commit**

```bash
git add application/services/script_generation_service.go application/services/storyboard_service.go application/services/video_generation_service.go application/services/video_merge_service.go application/services/usage_billing_integration_test.go
git commit -m "feat(billing): enforce credit consumption across AI usage flows"
```

---

### Task 7: Documentation + Final Verification

**Files:**

- Modify: `docs/saas-backend-apis.md`

**Step 1: Verify backend**

Run: `go test ./...`
Expected: PASS.

**Step 2: Verify frontend**

Run: `cd web && npm run build`
Expected: PASS.

**Step 3: Update docs**

- State that user-side AI config is removed.
- Document admin AI config endpoints and security rules.
- Document billing costs and which actions consume credits.

**Step 4: Commit**

```bash
git add docs/saas-backend-apis.md
git commit -m "docs: admin-only ai config and usage billing policy"
```

