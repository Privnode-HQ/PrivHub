---
method: POST
path: /api/user/passkey/login/finish
auth: passkey_login_session
handler: controller.PasskeyLoginFinish
source: router/api-router.go:69
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/passkey/login/finish`

提交浏览器 WebAuthn 登录结果并完成 Web 登录。

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

成功后响应与普通登录相同：

- `success`: `true`。
- `message`: 空字符串。
- `data.cah_id`: 用户 CAH ID。
- `data.username`: 用户名。
- `data.display_name`: 显示名。
- `data.role`: 用户角色。
- `data.status`: 用户状态。
- `data.group`: 用户分组。
- `data.email`: 绑定邮箱。
- `data.force_password_reset`: 是否需要强制重置密码。
- `data.force_email_bind`: 是否需要强制绑定邮箱。
- `data.required_actions`: 当前用户必须完成的动作列表。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `管理员未启用 Passkey 登录`、`未找到 Passkey 凭证`、`用户句柄与凭证不匹配`、`Passkey 登录状态异常`、`Passkey 凭证更新失败`、用户禁用提示或 WebAuthn 校验错误。

