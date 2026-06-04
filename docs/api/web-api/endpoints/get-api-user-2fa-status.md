---
method: GET
path: /api/user/2fa/status
auth: user
handler: controller.Get2FAStatus
source: router/api-router.go:113
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/user/2fa/status`

查询当前用户两步验证状态。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.enabled`: 布尔值。是否已启用 2FA。
- `data.locked`: 布尔值。2FA 是否处于锁定状态。
- `data.backup_codes_remaining`: 整数，可选。已启用 2FA 时返回，表示未使用备用码数量。

## 失败响应

- `success`: `false`。
- `message`: 2FA 查询错误。

