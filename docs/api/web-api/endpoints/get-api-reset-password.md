---
method: GET
path: /api/reset_password
auth: public_turnstile_rate_limited
handler: controller.SendPasswordResetEmail
source: router/api-router.go:42
request:
  query_params:
    - email
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/reset_password`

向已注册邮箱发送密码重置链接。

## 查询参数字段

- `email`: 字符串，必填。目标邮箱地址，必须满足邮箱格式且已注册。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`该邮箱地址未注册`，或邮件发送错误。

## 业务规则

成功时后端生成一次性 token，并把链接写入邮件：`<server_address>/user/reset?email=<email>&token=<token>`。

