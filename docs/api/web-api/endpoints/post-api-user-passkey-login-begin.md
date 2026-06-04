---
method: POST
path: /api/user/passkey/login/begin
auth: public_rate_limited
handler: controller.PasskeyLoginBegin
source: router/api-router.go:68
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/passkey/login/begin`

开始 Passkey Discoverable Login 流程，返回浏览器调用 WebAuthn 所需的 assertion options。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.options`: WebAuthn `PublicKeyCredentialRequestOptions` 对象，由 `go-webauthn` 生成并透传给前端。
- `data.options.challenge`: 登录挑战值。
- `data.options.timeout`: 建议超时时间。
- `data.options.rpId`: Relying Party ID。
- `data.options.allowCredentials`: 允许的凭证列表；Discoverable Login 下可能为空。
- `data.options.userVerification`: 用户验证偏好。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `管理员未启用 Passkey 登录`，或 WebAuthn/session 保存错误。

