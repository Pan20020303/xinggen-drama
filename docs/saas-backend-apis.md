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

---

## 8. 前端联调用接口示例

> 以下示例默认后端地址为：`http://localhost:5678`  
> 生产环境请替换为实际网关域名。

### 8.1 通用请求封装（前端 `fetch`）

```ts
const API_BASE = "http://localhost:5678/api/v1";

function getToken() {
  return localStorage.getItem("token") || "";
}

async function apiFetch(path: string, options: RequestInit = {}) {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) headers.Authorization = `Bearer ${token}`;

  const resp = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  const data = await resp.json();
  if (!resp.ok || data.success === false) {
    throw new Error(data?.error?.message || data?.message || "Request failed");
  }
  return data;
}
```

### 8.2 注册 / 登录

#### 注册

```ts
await apiFetch("/auth/register", {
  method: "POST",
  body: JSON.stringify({
    email: "user@example.com",
    password: "Passw0rd123",
  }),
});
```

#### 登录（保存 token）

```ts
const loginRes = await apiFetch("/auth/login", {
  method: "POST",
  body: JSON.stringify({
    email: "user@example.com",
    password: "Passw0rd123",
  }),
});

localStorage.setItem("token", loginRes.data.token);
localStorage.setItem("user", JSON.stringify(loginRes.data.user));
```

---

### 8.3 剧本（`dramas`）联调

#### 创建剧本

```ts
const createDramaRes = await apiFetch("/dramas", {
  method: "POST",
  body: JSON.stringify({
    title: "我的SaaS剧本",
    description: "测试数据",
    genre: "科幻",
  }),
});

const dramaId = createDramaRes.data.id;
```

#### 剧本列表（分页）

```ts
const dramaListRes = await apiFetch("/dramas?page=1&page_size=20");
console.log(dramaListRes.data.items, dramaListRes.data.pagination);
```

#### 更新剧本章节（用于后续创建分镜）

```ts
await apiFetch(`/dramas/${dramaId}/episodes`, {
  method: "PUT",
  body: JSON.stringify({
    episodes: [{ episode_number: 1, title: "第1集" }],
  }),
});
```

---

### 8.4 AI 配置联调（`ai-configs`）

#### 创建 AI 文本配置

```ts
await apiFetch("/ai-configs", {
  method: "POST",
  body: JSON.stringify({
    service_type: "text",
    name: "我的文本模型",
    provider: "openai",
    base_url: "https://api.openai.com/v1",
    api_key: "sk-xxxx",
    model: ["gpt-4o-mini"],
  }),
});
```

#### 查询 AI 配置列表

```ts
const cfgRes = await apiFetch("/ai-configs?service_type=text");
console.log(cfgRes.data);
```

---

### 8.5 图片生成联调（`images`）

#### 生成图片（会扣积分）

```ts
await apiFetch("/images", {
  method: "POST",
  body: JSON.stringify({
    drama_id: String(dramaId),
    prompt: "一个赛博朋克夜景城市，电影感，细节丰富",
  }),
});
```

#### 查询图片记录

```ts
const imgList = await apiFetch(`/images?drama_id=${dramaId}&page=1&page_size=20`);
console.log(imgList.data.items);
```

---

### 8.6 帧提示词联调（`frame-prompt`）

> 先确保有 storyboard（可通过 `/storyboards` 创建）。

#### 生成帧提示词（会扣积分）

```ts
await apiFetch(`/storyboards/${storyboardId}/frame-prompt`, {
  method: "POST",
  body: JSON.stringify({
    frame_type: "first", // first | key | last | panel | action
  }),
});
```

#### 查询帧提示词列表

```ts
const fpRes = await apiFetch(`/storyboards/${storyboardId}/frame-prompts`);
console.log(fpRes.data.frame_prompts);
```

---

### 8.7 典型错误处理（前端）

- `401 UNAUTHORIZED`：未登录或 token 失效，前端应跳转登录页并清理本地 token。
- `404 NOT_FOUND`：跨用户访问资源时常见（多租户隔离生效）。
- `500 + insufficient credits`：积分不足（建议前端提示“余额不足，请充值”）。

---

### 8.8 一组最小 `curl` 联调命令

```bash
# 1) 注册
curl -X POST http://localhost:5678/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Passw0rd123"}'

# 2) 登录（取 token）
curl -X POST http://localhost:5678/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Passw0rd123"}'

# 3) 创建剧本
curl -X POST http://localhost:5678/api/v1/dramas \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Demo Drama","description":"demo"}'

# 4) 创建 AI 配置
curl -X POST http://localhost:5678/api/v1/ai-configs \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"service_type":"text","name":"cfg1","provider":"openai","base_url":"https://api.openai.com/v1","api_key":"sk-xxx","model":["gpt-4o-mini"]}'

# 5) 图片生成（扣积分）
curl -X POST http://localhost:5678/api/v1/images \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"drama_id":"1","prompt":"test prompt"}'
```
