---
method: GET
path: /api/user/2fa/stats
auth: admin
handler: controller.Admin2FAStats
source: router/api-router.go:142
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/user/2fa/stats`

管理员获取全站 2FA 统计信息。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 2FA 统计对象，来自 `model.GetTwoFAStats()`。
- `data.total_users`: 若模型返回该字段，表示统计用户总数。
- `data.enabled_users`: 若模型返回该字段，表示已启用 2FA 用户数。
- `data.disabled_users`: 若模型返回该字段，表示未启用 2FA 用户数。

## 失败响应

- `success`: `false`。
- `message`: 统计查询错误。

