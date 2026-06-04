---
method: POST
path: /api/admin/service-accounts/:id/rotate
auth: admin
handler: controller.RotateAdminServiceAccountCredential
source: router/api-router.go:31
request:
  path_params:
    - id
  content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/admin/service-accounts/:id/rotate`

轮换 Admin Service Account 的 JWT 凭据。旧凭据立即失效，新凭据只在本次响应中返回一次。

## 路径参数字段

- `id`: 整数，必填。Service Account 数据库 ID。

## 请求体字段

请求体可为空。传入 JSON 时支持：

- `expires_at`: 整数，可选。新凭据过期时间 Unix 秒；必须晚于当前时间至少 60 秒，且不超过 365 天。
- `expires_in_days`: 整数，可选。有效天数；未给 `expires_at` 时使用，默认 90，范围 1 到 365。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.account`: 轮换后的 Service Account 视图，字段同 `GET /api/admin/service-accounts/` 的 `data.items[]`。
- `data.credential`: 字符串。新的 JWT 凭据。
- `data.credential_type`: 字符串，固定为 `Bearer JWT`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `Service Account ID 无效`、`无效的参数`、`Service Account 有效天数必须在 1 到 365 天之间`、目标管理员不可管理或数据库错误。

