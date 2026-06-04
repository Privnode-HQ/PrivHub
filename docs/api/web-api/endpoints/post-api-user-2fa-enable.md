---
method: POST
path: /api/user/2fa/enable
auth: user
handler: controller.Enable2FA
source: router/api-router.go:115
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/2fa/enable`

用 TOTP 验证码启用已初始化的 2FA。

## 请求体字段

- `code`: 字符串，必填。认证器中的 TOTP 数字验证码。

## 成功响应字段

- `success`: `true`。
- `message`: `两步验证启用成功`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、`请先完成2FA初始化设置`、`2FA已经启用`、`验证码或备用码错误，请重试`，或验证码格式/数据库错误。

