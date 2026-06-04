---
method: POST
path: /api/user/2fa/disable
auth: user
handler: controller.Disable2FA
source: router/api-router.go:116
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/2fa/disable`

禁用当前用户 2FA。可用 TOTP 验证码或备用码验证。

## 请求体字段

- `code`: 字符串，必填。TOTP 验证码或备用码。

## 成功响应字段

- `success`: `true`。
- `message`: `两步验证已禁用`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、`用户未启用2FA`、`验证码或备用码错误，请重试`，或备用码/数据库错误。

