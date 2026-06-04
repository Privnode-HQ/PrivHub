---
method: POST
path: /api/verify
auth: user
handler: controller.UniversalVerify
source: router/api-router.go:60
request:
  content_type: application/json
response:
  success_http_status: 200
  unauthorized_http_status: 401
  envelope: custom_success
---

# POST `/api/verify`

执行通用安全验证，用于需要二次确认的敏感操作。支持 2FA 和 Passkey，验证成功后在 Web session 中写入 5 分钟有效状态。

## 请求体字段

- `method`: 字符串，必填。验证方式，可能值为 `2fa`、`passkey`。
- `code`: 字符串，`method=2fa` 时必填。TOTP 验证码。

## 成功响应字段

- `success`: `true`。
- `message`: `验证成功`。
- `data.verified`: 布尔值，固定为 `true`。
- `data.expires_at`: 验证状态过期时间 Unix 秒，当前为验证成功后 300 秒。

## 失败响应

- HTTP 401 且 `success=false`: 未登录。
- HTTP 200 且 `success=false`: 可能为参数错误、用户被禁用、用户未启用 2FA 或 Passkey、验证码为空、不支持的验证方式、验证失败、保存 session 失败。

## 业务规则

- `method=passkey` 依赖前端先完成 Passkey begin/finish 流程；本接口只设置安全验证 session。
- 用户必须处于启用状态。

