# SaaS 改造已完成后端接口清单

> 更新时间：2026-02-09  
> 范围：当前已完成并联调通过的 SaaS MVP 后端接口能力（认证、多租户隔离、基础积分扣费）

---

## 1. 认证接口（新增）

### 1.1 用户注册

- **Method**: `POST`
- **Path**: `/api/v1/auth/register`
- **Body**:

```json
{
  "email": "user@example.com",
  "password": "Passw0rd123"
}
```

- **说明**:
  - 新建用户（默认 `role=user`）
  - 自动发放初始积分（`auth.initial_credits`）
  - 返回 JWT Token 和用户信息

### 1.2 用户登录

- **Method**: `POST`
- **Path**: `/api/v1/auth/login`
- **Body**:

```json
{
  "email": "user@example.com",
  "password": "Passw0rd123"
}
```

- **说明**:
  - 登录成功返回 JWT Token 和用户信息

---

## 2. 鉴权机制

- 除 `/api/v1/auth/*` 外，`/api/v1` 下业务接口已接入 JWT 鉴权中间件。
- 请求头必须包含：

```http
Authorization: Bearer <token>
```

---

## 3. 已接入多租户隔离的接口分组

以下分组已按 `user_id` 做资源隔离（查询/更新/删除均限制当前用户）：

- `/api/v1/dramas`（剧本相关）
- `/api/v1/ai-configs`（AI 配置）
- `/api/v1/character-library`（角色库）
- `/api/v1/characters`（角色相关）
- `/api/v1/images`（图片生成）
- `/api/v1/storyboards/:id/frame-prompt`（帧提示词生成）
- `/api/v1/storyboards/:id/frame-prompts`（帧提示词查询）

---

## 4. 已完成积分扣费规则（基础版）

### 4.1 图片生成扣费

- **接口**: `POST /api/v1/images`
- **扣费**: `billing.image_generation_credits`（默认 5）
- **行为**: 调用成功创建任务前扣费并记流水

### 4.2 帧提示词生成扣费

- **接口**: `POST /api/v1/storyboards/:id/frame-prompt`
- **扣费**: `billing.frame_prompt_credits`（默认 10）
- **行为**: 创建帧提示词任务前扣费并记流水

---

## 5. 关键模型改造（已完成）

已新增并启用：

- `users`
- `credit_transactions`

已增加 `user_id` 字段（并用于隔离）：

- `dramas`
- `characters`
- `episodes`
- `storyboards`
- `scenes`
- `props`
- `assets`
- `ai_service_configs`
- `image_generations`
- `frame_prompts`
- `character_libraries`

---

## 6. 已完成联调验证结果（摘要）

- 注册/登录成功（JWT 生效）
- 用户 A 创建的剧本，用户 B 访问返回 `404`
- 用户 A 的 AI 配置，用户 B 不可见
- 图片生成接口调用成功并扣费
- 帧提示词接口调用成功并扣费
- 积分总消耗符合规则：`5 + 10 = 15`

---

## 7. 当前配置项

已支持以下配置：

- `auth.jwt_secret`
- `auth.token_expire_hours`
- `auth.initial_credits`
- `billing.image_generation_credits`
- `billing.frame_prompt_credits`

