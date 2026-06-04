---
method: POST
path: /api/user/passkey/register/finish
auth: user_registration_session
handler: controller.PasskeyRegisterFinish
source: router/api-router.go:95
request:
  content_type: application/json
response:
  success_http_status: 200
  unauthorized_http_status: 401
  envelope: custom_success
---

# POST `/api/user/passkey/register/finish`

提交浏览器创建的 Passkey 凭证并完成绑定。

## 请求体字段

请求体是浏览器 `navigator.credentials.create()` 返回的 PublicKeyCredential JSON：

- `id`: 字符串。凭证 ID。
- `rawId`: 字符串。Base64URL 编码的原始凭证 ID。
- `type`: 字符串，通常为 `public-key`。
- `response.clientDataJSON`: 字符串。客户端数据。
- `response.attestationObject`: 字符串。attestation 对象。
- `clientExtensionResults`: 对象。浏览器扩展结果。

## 成功响应字段

- `success`: `true`。
- `message`: `Passkey 注册成功`。

## 失败响应

- HTTP 401 且 `success=false`: 未登录或 session 无效。
- HTTP 200 且 `success=false`: `管理员未启用 Passkey 登录`、`无法创建 Passkey 凭证`，或 WebAuthn/session/数据库错误。

