# 用户迁移 API 文档

本文档描述 PrivHub 用户迁移功能的 HTTP API。除明确标注为公开迁移流程的接口外，所有接口都要求普通管理员权限。

## 通用约定

- 所有迁移相关 URI 都以 `/migrate` 开头。
- 管理员接口前缀：`/migrate/api/admin`
- 用户迁移流程接口前缀：`/migrate/api`
- 管理员接口使用现有 Web 管理员会话认证，并沿用 `New-API-User` 请求头校验。
- 响应格式沿用项目标准：

```json
{
  "success": true,
  "message": "",
  "data": {}
}
```

## 状态枚举

迁移任务 `status`：

- `draft`：草稿
- `active`：启用，允许追加目标和用户确认
- `closed`：关闭，不再允许用户确认
- `cancelled`：取消

迁移目标 `status`：

- `pending`：已创建，未发送或待处理
- `email_sent`：迁移指引邮件已发送
- `email_failed`：邮件发送失败
- `opened`：用户已打开链接并验证 token
- `captured`：用户已登录指定账户，迁移数据已记录
- `migrated`：接收方确认导入成功，原账户已禁用

导入记录 `status`：

- `pending_setup`：已导入，等待用户设置密码
- `active`：用户已设置密码并启用账户
- `email_failed`：密码设置邮件发送失败

## 筛选表达式

### GET `/migrate/api/admin/expression-docs`

获取管理员填写界面使用的表达式字段、运算符和示例。

支持字段：

- `username`
- `email`
- `cah` / `cah_id`
- `provider`
- `status`
- `role`
- `group`

支持运算符：

- `==`
- `!=`
- `eq`
- `ne`
- `contains`
- `matches`
- `in`

示例：

```text
status == "enabled" and email matches ".*@example\\.com$"
provider in {"github", "oidc"} and group == "paid"
(role == "common" or role == "support") and not email contains "+test"
```

### POST `/migrate/api/admin/preview`

预览表达式命中的用户。

表达式筛选默认只包含未软删除且启用的用户；如果表达式显式使用 `status` 字段，则按表达式状态匹配。超级管理员始终不会进入迁移目标。

请求：

```json
{
  "expression": "status == \"enabled\" and group == \"paid\"",
  "limit": 10
}
```

响应 `data`：

```json
{
  "count": 2,
  "users": [
    {
      "user_id": 123,
      "cah_id": "ABCDE1",
      "username": "alice",
      "display_name": "Alice",
      "email": "alice@example.com",
      "group": "paid",
      "role": 1,
      "status": 1,
      "providers": ["github", "password"]
    }
  ]
}
```

## 迁移任务

### POST `/migrate/api/admin/migrations`

创建并发起迁移任务。表达式非空时会立即生成目标用户；`send_email=true` 时会按当前 Postmark 批量发送实践发送迁移指引。

请求：

```json
{
  "name": "2026 Q2 paid users migration",
  "description": "迁移 paid 分组启用用户",
  "expression": "status == \"enabled\" and group == \"paid\"",
  "send_email": true
}
```

响应 `data`：

```json
{
  "migration": {
    "migrate_id": "mig_xxxxxxxxxxxxxxxx",
    "name": "2026 Q2 paid users migration",
    "status": "active",
    "target_count": 10,
    "email_sent_count": 10
  },
  "add_targets": {
    "created": 10,
    "duplicated": 0,
    "email_sent": 10,
    "email_failed": 0,
    "invalid": 0
  }
}
```

### GET `/migrate/api/admin/migrations`

分页列出迁移任务。

查询参数：

- `p`：页码
- `page_size`：每页数量

### GET `/migrate/api/admin/migrations/{migrate_id}`

获取单个迁移任务。

### PUT `/migrate/api/admin/migrations/{migrate_id}/status`

更新迁移任务状态。

请求：

```json
{
  "status": "closed"
}
```

## 迁移目标

### GET `/migrate/api/admin/migrations/{migrate_id}/targets`

分页列出迁移任务的目标用户。

### POST `/migrate/api/admin/migrations/{migrate_id}/targets`

向迁移任务追加目标用户。可通过表达式追加，也可通过 `user_id` 或 `cah_id` 追加单个或多个用户。

请求：按表达式追加

```json
{
  "expression": "status == \"enabled\" and email matches \".*@example\\.com$\"",
  "send_email": true
}
```

请求：按用户追加

```json
{
  "send_email": true,
  "targets": [
    {
      "user_id": 123,
      "email": "target@example.com"
    },
    {
      "cah_id": "ABCDE1",
      "email": "other@example.com"
    }
  ]
}
```

说明：

- `email` 是迁移通知发送邮箱，优先使用请求传入值。
- 如果未传 `email`，表达式追加会使用用户当前邮箱。
- 同一迁移任务中同一用户只能存在一次。
- 超级管理员用户不会被加入迁移目标。
- token 只在生成链接时返回或发送，不以明文持久化。

## 用户迁移流程接口

迁移链接格式：

```text
/migrate/<migrate_id>/#~/t/<migrate_token>/<access_token>/~/u/<user_token>
```

前端页面读取 fragment 后调用以下接口。

`user_token` 是必填安全参数，缺失时验证、登录和捕获都会失败。

### POST `/migrate/api/migrations/{migrate_id}/verify`

验证迁移链接并返回目标账户信息。

请求：

```json
{
  "migration_token": "token",
  "access_token": "token",
  "user_token": "token"
}
```

响应 `data`：

```json
{
  "migration_id": "mig_xxxxxxxxxxxxxxxx",
  "target_id": 1,
  "cah_id": "ABCDE1",
  "email": "target@example.com",
  "status": "opened",
  "login_ok": false,
  "captured": false
}
```

### POST `/migrate/api/migrations/{migrate_id}/login`

在 `/migrate` 流程内登录指定原 PrivHub 账户，并记录迁移数据。

请求：

```json
{
  "migration_token": "token",
  "access_token": "token",
  "user_token": "token",
  "username": "alice",
  "password": "password"
}
```

成功后系统会：

- 校验登录账户必须等于迁移目标用户。
- 记录 CAH、迁移邮箱和可迁移个人数据。
- 将目标状态更新为 `captured`。
- 建立普通 Web 会话。

### POST `/migrate/api/migrations/{migrate_id}/capture`

用户已在当前浏览器登录指定账户时，可直接确认并记录迁移数据。

请求：

```json
{
  "migration_token": "token",
  "access_token": "token",
  "user_token": "token"
}
```

## 导出与确认

### GET `/migrate/api/admin/exports`

获取按记录创建时间排序的最新待导出元数据。列表接口不会返回完整迁移 JSON，也不会更新 `last_exported_at`。

查询参数：

- `migrate_id`：可选，只取指定迁移任务
- `limit`：可选，默认 10，最大 100

响应 `data`：

```json
[
  {
    "target_id": 1,
    "migrate_id": "mig_xxxxxxxxxxxxxxxx",
    "cah_id": "ABCDE1",
    "email": "target@example.com",
    "captured_at": 1779870000,
    "last_exported_at": 0,
    "data_size": 65536,
    "download_url": "/migrate/api/admin/exports/1"
  }
]
```

### GET `/migrate/api/admin/exports/{target_id}`

下载单个目标的完整迁移 JSON 文件。下载尝试不会被当作接收方已成功取得文件的强信号，也不会自动更新 `last_exported_at`；只有接收方确认导入成功后，才调用确认接口。

### POST `/migrate/api/admin/exports/confirm`

接收方确认导入成功后调用。确认后：

- 迁移目标状态更新为 `migrated`。
- 原用户账户被禁用。
- 禁用原因写为 `已迁移`。
- 原账户 Web 会话版本递增，已有 Web 会话失效。

请求：

```json
{
  "target_ids": [1, 2, 3]
}
```

响应：

```json
{
  "updated": 3
}
```

## 导入用户

### POST `/migrate/api/admin/imports/users`

上传 CAH、邮箱和迁移 JSON，新增用户并导入可安全落库的数据。邮箱和 CAH 必须不存在，否则拒绝。

请求：

```json
{
  "cah_id": "ABCDE1",
  "email": "target@example.com",
  "data": {
    "version": 1,
    "account": {
      "display_name": "Alice",
      "group": "paid",
      "quota": 10000,
      "setting": "{}"
    },
    "records": {
      "topups": [],
      "topup_coupons": [],
      "usage_windows": []
    }
  }
}
```

成功后：

- 创建用户，用户名由系统生成。
- 用户初始为禁用状态，原因是等待迁移密码设置。
- 不导入密码、第三方登录、会话、passkey、2FA 或 API key 明文。
- 发送密码设置邮件，链接位于 `/migrate/import/#~/i/<setup_token>/<access_token>`。
- 如果密码设置邮件发送失败，接口仍返回成功结果，`status` 为 `email_failed`，并返回 `link` 供管理员恢复处理；重复导入仍会因 CAH 或邮箱已存在而被拒绝。

响应：

```json
{
  "import_id": "imp_xxxxxxxxxxxxxxxx",
  "user_id": 456,
  "cah_id": "ABCDE1",
  "email": "target@example.com",
  "status": "pending_setup",
  "link": "https://example.com/migrate/import/#~/i/..."
}
```

## 导入后设置密码

### POST `/migrate/api/imports/setup/verify`

验证导入后的密码设置链接。

请求：

```json
{
  "setup_token": "token",
  "access_token": "token"
}
```

### POST `/migrate/api/imports/setup/password`

设置密码并启用导入用户。

请求：

```json
{
  "setup_token": "token",
  "access_token": "token",
  "password": "new-password"
}
```

成功后：

- 用户状态改为启用。
- 清空禁用原因。
- 清除强制密码设置标记。
- 递增 Web 会话版本。

## 数据范围

导出包含：

- 顶层 `cah_id`
- 迁移通知邮箱 `email`
- 账户可迁移字段：展示名、分组、额度、用量计数、设置、备注、Stripe customer、订阅数据等
- 充值记录
- 充值优惠券
- 兑换码元数据，不含兑换码密钥
- API token 元数据，不含 token key
- 使用限制窗口计数
- 用户站内消息快照

导出明确排除：

- 用户 ID
- 用户名
- 密码
- 第三方登录 ID
- access token
- API key 明文
- session
- passkey
- 2FA
- 使用日志
- 任务/绘图日志
- 管理员审计日志

## Postmark 批量发送实践

迁移邮件使用现有 Postmark 封装：

- 使用 `common.SendBatchEmailsWithIdempotencyKey` 或同等单封发送路径。
- Message Stream 使用 transactional `outbound`。
- 批量发送保留每个目标用户的发送结果。
- 每个迁移任务和目标用户生成稳定幂等键，降低重复提交造成重复投递的风险。
- 部分失败不会被视为全量成功，失败原因会写入目标记录 `email_error`。
