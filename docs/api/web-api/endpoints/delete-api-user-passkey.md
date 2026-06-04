---
method: DELETE
path: /api/user/passkey
auth: user
handler: controller.PasskeyDelete
source: router/api-router.go:98
request:
  body: none
response:
  success_http_status: 200
  unauthorized_http_status: 401
  envelope: custom_success
---

# DELETE `/api/user/passkey`

解绑当前用户的 Passkey。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: `Passkey 已解绑`。

## 失败响应

- HTTP 401 且 `success=false`: 未登录或 session 无效。
- HTTP 200 且 `success=false`: 删除错误。

