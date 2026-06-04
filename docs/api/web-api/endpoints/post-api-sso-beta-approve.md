---
method: POST
path: /api/sso-beta/approve
auth: user + session
handler: controller.SSOApprove
source: router/api-router.go:369
request:
  body:
    - client_id
    - nonce
    - metadata
    - postauth
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/sso-beta/approve`

当前登录用户批准 SSO 授权，并生成回调 URL。

## 请求体字段

- `client_id`: SSO 客户端 ID，必填，必须在允许列表中。
- `nonce`: SSO nonce，必填，原样带入回调参数。
- `metadata`: 可选元数据，非空时带入回调参数。
- `postauth`: 回调域名，必填；后端拼接为 `https://{postauth}/sso/callback`。

## 成功响应字段

- `success`: `true`。
- `data.redirect_url`: SSO 回调 URL，包含 `nonce`、`token`，以及可选 `metadata`。

## 失败响应

- HTTP `401`: `success=false`，`message=未登录`。
- HTTP `400`: `success=false`，`message=无效的请求参数` 或 `无效的客户端ID`。
- HTTP `500`: `success=false`，`message` 可能为 `获取用户信息失败`、`生成访问令牌失败`、`保存访问令牌失败`、`生成 token 失败`。

