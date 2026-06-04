---
method: POST
path: /api/user/passkey/register/begin
auth: user
handler: controller.PasskeyRegisterBegin
source: router/api-router.go:94
request:
  body: none
response:
  success_http_status: 200
  unauthorized_http_status: 401
  envelope: custom_success
---

# POST `/api/user/passkey/register/begin`

开始当前用户 Passkey 注册流程，返回浏览器创建凭证所需的 WebAuthn options。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.options`: WebAuthn `PublicKeyCredentialCreationOptions` 对象。
- `data.options.challenge`: 注册挑战值。
- `data.options.rp`: Relying Party 信息。
- `data.options.user`: WebAuthn 用户信息。
- `data.options.pubKeyCredParams`: 支持的公钥算法列表。
- `data.options.timeout`: 建议超时时间。
- `data.options.excludeCredentials`: 已绑定凭证排除列表；存在旧凭证时返回。
- `data.options.authenticatorSelection`: 认证器偏好。
- `data.options.attestation`: attestation 偏好。

## 失败响应

- HTTP 401 且 `success=false`: 未登录或 session 无效。
- HTTP 200 且 `success=false`: `管理员未启用 Passkey 登录`，或 WebAuthn/session 保存错误。

