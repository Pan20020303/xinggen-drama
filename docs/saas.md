# 星亘剧生成器 SaaS 化改造方案

## 1. 总体目标

将现有的单机版分镜生成工具改造成支持多用户、计费、云端存储的 SaaS 平台。

## 2. 核心架构变更

### 2.1 用户身份体系 (Identity & Access Management)

* **新增 User 模型** :
* `ID`: 主键
* `Email/Phone`: 登录账号
* `Password`: 加密存储
* `Role`: `admin` (管理员), `user` (普通用户), `vip` (会员)
* `Balance`: 账户余额/积分
* **认证机制** :
* 实现 JWT (JSON Web Token) 认证
* 添加 `AuthMiddleware`，解析 Token 并将 `UserID` 注入 Context

### 2.2 多租户数据隔离 (Multi-Tenancy)

* **数据模型改造** :
* 所有资源表 (`dramas`, `episodes`, `scenes`, `assets`, `ai_configs`) 添加 `user_id` 字段
* 建立索引 `idx_user_id`
* **业务逻辑改造** :
* 修改所有 CRUD 操作，强制带上 `Where("user_id = ?", uid)`
* 禁止跨用户访问资源

### 2.3 计费与配额系统 (Billing & Quota)

* **积分系统** :
* 新增 `CreditTransaction` 表记录消费/充值流水
* 定义计费规则 (例如: 分镜生成=10积分/次, 图片生成=5积分/张)
* **AI 服务模式调整** :
* **平台模式** : 管理员配置全局 AI Key，用户消耗积分使用
* **BYOK 模式** (Bring Your Own Key): 用户配置自己的 Key，免积分或低积分

### 2.4 存储云化 (Cloud Storage)

* **本地存储 -> 云对象存储 (OSS/S3)** :
* 实现

  Storage 接口的 S3/OSS 版本
* 图片/视频上传后直接传至云端，数据库只存 URL
* 修改视频合成逻辑，支持从 URL 下载素材进行合成

### 2.5 基础设施升级

* **数据库** : SQLite -> PostgreSQL (更好的并发和事务支持)
* **缓存/队列** : 引入 Redis (用于验证码、任务队列、限流)
* **部署** : Docker 容器化 + Docker Compose/K8s

---

## 3. 实施路径 (Roadmap)

### 第一阶段：基础 SaaS 化 (MVP)

0� [ ]  **用户系统** : 实现注册、登录、JWT 认证
0� [ ]  **数据隔离** : 给核心表添加 `user_id`，并迁移现有代码
0� [ ]  **数据库迁移** : 适配 PostgreSQL
0� [ ]  **简单计费** : 每个新用户赠送固定积分，操作扣减积分

### 第二阶段：云原生适配

0� [ ]  **OSS 集成** : 对接阿里云 OSS 或 AWS S3
0� [ ]  **异步任务队列** : 使用 Redis 或消息队列优化耗时任务

### 第三阶段：商业化功能

0� [ ]  **支付对接** : 集成支付网关
0� [ ]  **会员体系** : 订阅制 (Subscription) 权益
0� [ ]  **后台管理** : 用户管理、充值管理、系统配置

---

## 4. 数据库变更预览 (SQL)

<pre><div node="[object Object]" class="relative whitespace-pre-wrap word-break-all my-2 rounded-lg bg-list-hover-subtle border border-gray-500/20"><div class="min-h-7 relative box-border flex flex-row items-center justify-between rounded-t border-b border-gray-500/20 px-2 py-0.5"><div class="font-sans text-sm text-ide-text-color opacity-60">sql</div><div class="flex flex-row gap-2 justify-end"><div class="cursor-pointer opacity-70 hover:opacity-100"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" class="lucide lucide-copy h-3.5 w-3.5"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"></rect><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"></path></svg></div></div></div><div class="p-3"><div class="w-full h-full text-xs cursor-text"><div class="code-block"><div class="code-line" data-line-number="1" data-line-start="1" data-line-end="1"><div class="line-content"><span class="mtk4">-- 用户表</span></div></div><div class="code-line" data-line-number="2" data-line-start="2" data-line-end="2"><div class="line-content"><span class="mtk5">CREATE</span><span class="mtk1"></span><span class="mtk5">TABLE</span><span class="mtk1"></span><span class="mtk15">users</span><span class="mtk1"> (</span></div></div><div class="code-line" data-line-number="3" data-line-start="3" data-line-end="3"><div class="line-content"><span class="mtk1">    id </span><span class="mtk5">BIGINT</span><span class="mtk1"></span><span class="mtk5">PRIMARY KEY</span><span class="mtk1"> AUTO_INCREMENT,</span></div></div><div class="code-line" data-line-number="4" data-line-start="4" data-line-end="4"><div class="line-content"><span class="mtk1">    email </span><span class="mtk5">VARCHAR</span><span class="mtk1">(</span><span class="mtk6">255</span><span class="mtk1">) </span><span class="mtk5">UNIQUE</span><span class="mtk1"></span><span class="mtk5">NOT NULL</span><span class="mtk1">,</span></div></div><div class="code-line" data-line-number="5" data-line-start="5" data-line-end="5"><div class="line-content"><span class="mtk1">    password_hash </span><span class="mtk5">VARCHAR</span><span class="mtk1">(</span><span class="mtk6">255</span><span class="mtk1">) </span><span class="mtk5">NOT NULL</span><span class="mtk1">,</span></div></div><div class="code-line" data-line-number="6" data-line-start="6" data-line-end="6"><div class="line-content"><span class="mtk1"></span><span class="mtk5">role</span><span class="mtk1"></span><span class="mtk5">VARCHAR</span><span class="mtk1">(</span><span class="mtk6">20</span><span class="mtk1">) </span><span class="mtk5">DEFAULT</span><span class="mtk1"></span><span class="mtk11">'user'</span><span class="mtk1">,</span></div></div><div class="code-line" data-line-number="7" data-line-start="7" data-line-end="7"><div class="line-content"><span class="mtk1">    credits </span><span class="mtk5">INT</span><span class="mtk1"></span><span class="mtk5">DEFAULT</span><span class="mtk1"></span><span class="mtk6">0</span><span class="mtk1">,</span></div></div><div class="code-line" data-line-number="8" data-line-start="8" data-line-end="8"><div class="line-content"><span class="mtk1">    created_at </span><span class="mtk5">TIMESTAMP</span><span class="mtk1"></span><span class="mtk5">DEFAULT</span><span class="mtk1"> CURRENT_TIMESTAMP</span></div></div><div class="code-line" data-line-number="9" data-line-start="9" data-line-end="9"><div class="line-content"><span class="mtk1">);</span></div></div><div class="code-line" data-line-number="10" data-line-start="10" data-line-end="10"><div class="line-content"><span class="mtk1"></span></div></div><div class="code-line" data-line-number="11" data-line-start="11" data-line-end="11"><div class="line-content"><span class="mtk4">-- 为现有表添加 user_id</span></div></div><div class="code-line" data-line-number="12" data-line-start="12" data-line-end="12"><div class="line-content"><span class="mtk5">ALTER</span><span class="mtk1"></span><span class="mtk5">TABLE</span><span class="mtk1"> dramas </span><span class="mtk5">ADD</span><span class="mtk1"> COLUMN user_id </span><span class="mtk5">BIGINT</span><span class="mtk1"></span><span class="mtk5">NOT NULL</span><span class="mtk1"></span><span class="mtk5">DEFAULT</span><span class="mtk1"></span><span class="mtk6">0</span><span class="mtk1">;</span></div></div><div class="code-line" data-line-number="13" data-line-start="13" data-line-end="13"><div class="line-content"><span class="mtk5">CREATE</span><span class="mtk1"></span><span class="mtk5">INDEX</span><span class="mtk1"></span><span class="mtk15">idx_dramas_user_id</span><span class="mtk1"></span><span class="mtk5">ON</span><span class="mtk1"> dramas(user_id);</span></div></div><div class="code-line" data-line-number="14" data-line-start="14" data-line-end="14"><div class="line-content"><span class="mtk1"></span></div></div><div class="code-line" data-line-number="15" data-line-start="15" data-line-end="15"><div class="line-content"><span class="mtk4">-- 消费流水</span></div></div><div class="code-line" data-line-number="16" data-line-start="16" data-line-end="16"><div class="line-content"><span class="mtk5">CREATE</span><span class="mtk1"></span><span class="mtk5">TABLE</span><span class="mtk1"></span><span class="mtk15">credit_transactions</span><span class="mtk1"> (</span></div></div><div class="code-line" data-line-number="17" data-line-start="17" data-line-end="17"><div class="line-content"><span class="mtk1">    id </span><span class="mtk5">BIGINT</span><span class="mtk1"></span><span class="mtk5">PRIMARY KEY</span><span class="mtk1"> AUTO_INCREMENT,</span></div></div><div class="code-line" data-line-number="18" data-line-start="18" data-line-end="18"><div class="line-content"><span class="mtk1">    user_id </span><span class="mtk5">BIGINT</span><span class="mtk1"></span><span class="mtk5">NOT NULL</span><span class="mtk1">,</span></div></div><div class="code-line" data-line-number="19" data-line-start="19" data-line-end="19"><div class="line-content"><span class="mtk1">    amount </span><span class="mtk5">INT</span><span class="mtk1"></span><span class="mtk5">NOT NULL</span><span class="mtk1">, </span><span class="mtk4">-- 负数为消费，正数为充值</span></div></div><div class="code-line" data-line-number="20" data-line-start="20" data-line-end="20"><div class="line-content"><span class="mtk1"></span><span class="mtk5">type</span><span class="mtk1"></span><span class="mtk5">VARCHAR</span><span class="mtk1">(</span><span class="mtk6">50</span><span class="mtk1">) </span><span class="mtk5">NOT NULL</span><span class="mtk1">, </span><span class="mtk4">-- GENERATE_STORYBOARD, RECHARGE, etc.</span></div></div><div class="code-line" data-line-number="21" data-line-start="21" data-line-end="21"><div class="line-content"><span class="mtk1"></span><span class="mtk5">description</span><span class="mtk1"></span><span class="mtk5">VARCHAR</span><span class="mtk1">(</span><span class="mtk6">255</span><span class="mtk1">),</span></div></div><div class="code-line" data-line-number="22" data-line-start="22" data-line-end="22"><div class="line-content"><span class="mtk1">    created_at </span><span class="mtk5">TIMESTAMP</span><span class="mtk1"></span><span class="mtk5">DEFAULT</span><span class="mtk1"> CURRENT_TIMESTAMP</span></div></div><div class="code-line" data-line-number="23" data-line-start="23" data-line-end="23"><div class="line-content"><span class="mtk1">);</span></div></div></div></div></div></div></pre>

## 5. 待确认事项

* **部署目标** : 是部署在国内 (阿里云/腾讯云)
* **支付方式** : 优先支持微信/支付宝
* **存量数据** : 不需要保存
