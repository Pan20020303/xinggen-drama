# Platform Admin A/B Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement platform operations backend capabilities for User Management (A) and Billing (B), including dedicated admin auth, admin-only APIs, and audit logging.

**Architecture:** Extend the existing Go monolith with a dedicated `/api/v1/admin/*` API surface. Keep business APIs unchanged while adding a strict admin auth boundary (`aud=admin` + admin role check), admin services for user and billing operations, and transactional audit logs for every write action.

**Tech Stack:** Go, Gin, GORM, SQLite/MySQL, JWT v5, existing logger/response/config modules.

---

### Task 1: Add Domain Models For Admin Ops

**Files:**
- Modify: `domain/models/user.go`
- Create: `domain/models/admin_audit_log.go`
- Modify: `infrastructure/database/database.go`
- Test: `domain/models/admin_models_test.go`

**Step 1: Write the failing test**

Write tests that assert:
- user role supports platform admin
- user has status field defaulting to active behavior
- admin audit log table name is `admin_audit_logs`

**Step 2: Run test to verify it fails**

Run: `go test ./domain/models -run TestAdmin -v`
Expected: FAIL due missing symbols/fields.

**Step 3: Write minimal implementation**

Add:
- platform admin role constant
- user status type/constants + status field
- admin audit log model
- include new model in AutoMigrate.

**Step 4: Run test to verify it passes**

Run: `go test ./domain/models -run TestAdmin -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add domain/models/user.go domain/models/admin_audit_log.go infrastructure/database/database.go domain/models/admin_models_test.go
git commit -m "feat: add admin domain models and audit log schema"
```

### Task 2: Add Admin Auth Token Path

**Files:**
- Modify: `application/services/auth_service.go`
- Create: `application/services/auth_service_admin_test.go`

**Step 1: Write the failing test**

Add tests for:
- `AdminLogin` rejects non-admin user
- `AdminLogin` accepts admin user
- admin token contains `aud=admin`
- normal token contains `aud=user`

**Step 2: Run test to verify it fails**

Run: `go test ./application/services -run TestAdminLogin -v`
Expected: FAIL due missing admin login/token audience behavior.

**Step 3: Write minimal implementation**

Add:
- `AdminLogin` method
- helper role checks for admin
- token generation with explicit audience (`user` vs `admin`).

**Step 4: Run test to verify it passes**

Run: `go test ./application/services -run TestAdminLogin -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add application/services/auth_service.go application/services/auth_service_admin_test.go
git commit -m "feat: add dedicated admin login and audience-scoped JWT"
```

### Task 3: Add Admin Middleware Guardrails

**Files:**
- Modify: `api/middlewares/auth.go`
- Create: `api/middlewares/admin_auth_test.go`

**Step 1: Write the failing test**

Test:
- user token cannot access admin-protected route
- admin token can access admin-protected route
- admin token cannot access user-protected route

**Step 2: Run test to verify it fails**

Run: `go test ./api/middlewares -run TestAdmin -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Add:
- `AdminAuthMiddleware`
- role helper for platform admin
- audience checks in both auth middlewares.

**Step 4: Run test to verify it passes**

Run: `go test ./api/middlewares -run TestAdmin -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add api/middlewares/auth.go api/middlewares/admin_auth_test.go
git commit -m "feat: enforce admin-only middleware and token audience checks"
```

### Task 4: Implement Admin User + Billing Services

**Files:**
- Create: `application/services/admin_audit_service.go`
- Create: `application/services/admin_user_service.go`
- Create: `application/services/admin_billing_service.go`
- Create: `application/services/admin_services_test.go`

**Step 1: Write the failing test**

Add tests for:
- list users pagination
- update user status
- update user role
- recharge updates credits + writes credit txn + audit log

**Step 2: Run test to verify it fails**

Run: `go test ./application/services -run TestAdminService -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Implement services with DB transactions and audit writes for mutations.

**Step 4: Run test to verify it passes**

Run: `go test ./application/services -run TestAdminService -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add application/services/admin_*.go application/services/admin_services_test.go
git commit -m "feat: add admin user and billing services with audit logging"
```

### Task 5: Implement Admin Handlers + Routes

**Files:**
- Create: `api/handlers/admin_auth.go`
- Create: `api/handlers/admin_user.go`
- Create: `api/handlers/admin_billing.go`
- Modify: `api/routes/routes.go`
- Create: `api/handlers/admin_handlers_test.go`

**Step 1: Write the failing test**

Add handler tests for:
- admin login endpoint
- users list endpoint
- status/role patch endpoint
- recharge endpoint
- transactions list endpoint

**Step 2: Run test to verify it fails**

Run: `go test ./api/handlers -run TestAdminHandler -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Implement handlers and wire `/api/v1/admin/*` groups under `AdminAuthMiddleware`.

**Step 4: Run test to verify it passes**

Run: `go test ./api/handlers -run TestAdminHandler -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add api/handlers/admin_*.go api/routes/routes.go api/handlers/admin_handlers_test.go
git commit -m "feat: expose admin auth users and billing APIs"
```

### Task 6: Verification And Documentation

**Files:**
- Modify: `docs/saas-backend-apis.md`

**Step 1: Run full verification**

Run:
- `go test ./...`

Expected: all pass.

**Step 2: Update docs**

Document:
- admin login
- admin user management APIs
- admin billing APIs
- audit behavior and constraints.

**Step 3: Commit**

```bash
git add docs/saas-backend-apis.md
git commit -m "docs: add platform admin A/B API documentation"
```
