---
method: POST
path: /api/admin/service-accounts/
auth: admin
handler: controller.CreateAdminServiceAccount
source: router/api-router.go:29
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/admin/service-accounts/`

创建 Admin Service Account，并一次性返回 Bearer JWT 凭据。

## 请求体字段

- `name`: 字符串，必填。Service Account 名称，去除空白后不能为空，最多 80 个字符。
- `description`: 字符串或 `null`，可选。描述，最多 255 个字符。
- `target`: 字符串，可选。目标管理员，可为用户 ID、CAH ID 或用户名；为空时默认当前管理员。
- `user_id`: 整数，可选。目标管理员用户 ID；优先级高于 `target`。
- `expires_at`: 整数，可选。过期时间 Unix 秒；必须晚于当前时间至少 60 秒，且不超过 365 天。
- `expires_in_days`: 整数，可选。有效天数；未给 `expires_at` 时使用，默认 90，范围 1 到 365。
- `allow_ips`: 字符串或 `null`，可选。允许来源 IP 列表，格式由模型规范化；为空表示不限制。
- `status`: 整数。创建时忽略，后端固定为启用。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.account`: 创建后的 Service Account 视图，字段同 `GET /api/admin/service-accounts/` 的 `data.items[]`。
- `data.credential`: 字符串。一次性显示的 JWT 凭据，后端只保存哈希，后续无法再次取回。
- `data.credential_type`: 字符串，固定为 `Bearer JWT`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`Service Account 名称不能为空`、`只能为管理员或超级管理员生成 Service Account`、`Service Account 有效期不能超过 365 天`、IP 白名单错误或数据库错误。

