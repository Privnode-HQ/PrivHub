---
method: GET
path: /api/user/passkey
auth: user
handler: controller.PasskeyStatus
source: router/api-router.go:93
request:
  body: none
response:
  success_http_status: 200
  unauthorized_http_status: 401
  envelope: custom_success
---

# GET `/api/user/passkey`

查询当前用户是否已绑定 Passkey。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.enabled`: 布尔值。是否已绑定 Passkey。
- `data.last_used_at`: 可选。Passkey 最近使用时间；仅已绑定时返回。

## 失败响应

- HTTP 401 且 `success=false`: 未登录或 session 无效。
- HTTP 200 且 `success=false`: Passkey 查询错误。

