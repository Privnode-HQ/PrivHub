---
method: POST
path: /api/user/passkey/verify/begin
auth: user
handler: controller.PasskeyVerifyBegin
source: router/api-router.go:96
request:
  body: none
response:
  success_http_status: 200
  unauthorized_http_status: 401
  envelope: custom_success
---

# POST `/api/user/passkey/verify/begin`

开始当前已登录用户的 Passkey 安全验证流程。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.options`: WebAuthn `PublicKeyCredentialRequestOptions` 对象。
- `data.options.challenge`: 验证挑战值。
- `data.options.timeout`: 建议超时时间。
- `data.options.rpId`: Relying Party ID。
- `data.options.allowCredentials`: 当前用户已绑定凭证描述。
- `data.options.userVerification`: 用户验证偏好。

## 失败响应

- HTTP 401 且 `success=false`: 未登录或 session 无效。
- HTTP 200 且 `success=false`: `管理员未启用 Passkey 登录`、`该用户尚未绑定 Passkey`，或 WebAuthn/session 错误。

