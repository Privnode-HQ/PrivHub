---
method: POST
path: /api/user/passkey/verify/finish
auth: user_verify_session
handler: controller.PasskeyVerifyFinish
source: router/api-router.go:97
request:
  content_type: application/json
response:
  success_http_status: 200
  unauthorized_http_status: 401
  envelope: custom_success
---

# POST `/api/user/passkey/verify/finish`

提交 Passkey 安全验证结果。成功后更新凭证最后使用时间，并设置通用安全验证 session。

## 请求体字段

请求体是浏览器 `navigator.credentials.get()` 返回的 PublicKeyCredential JSON：

- `id`: 字符串。凭证 ID。
- `rawId`: 字符串。Base64URL 编码的原始凭证 ID。
- `type`: 字符串，通常为 `public-key`。
- `response.clientDataJSON`: 字符串。客户端数据。
- `response.authenticatorData`: 字符串。认证器数据。
- `response.signature`: 字符串。签名。
- `response.userHandle`: 字符串，可选。用户句柄。
- `clientExtensionResults`: 对象。浏览器扩展结果。

## 成功响应字段

- `success`: `true`。
- `message`: `Passkey 验证成功`。

## 失败响应

- HTTP 401 且 `success=false`: 未登录或 session 无效。
- HTTP 200 且 `success=false`: `管理员未启用 Passkey 登录`、`该用户尚未绑定 Passkey`，或 WebAuthn/session/数据库错误。

