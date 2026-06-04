---
method: POST
path: /api/user/reset
auth: public_rate_limited
handler: controller.ResetPassword
source: router/api-router.go:43
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: common
  data: string
---

# POST `/api/user/reset`

使用密码重置邮件中的 token 重置密码。后端会生成一个新密码并在响应中返回。

## 请求体字段

- `email`: 字符串，必填。接收重置邮件的邮箱。
- `token`: 字符串，必填。重置邮件中的一次性 token。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 字符串。新生成的 12 位密码。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`重置链接非法或已过期`，或数据库错误。

