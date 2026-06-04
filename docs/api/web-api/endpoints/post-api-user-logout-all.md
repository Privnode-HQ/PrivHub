---
method: POST
path: /api/user/logout_all
auth: admin
handler: controller.LogoutAllUsers
source: router/api-router.go:138
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/logout_all`

管理员递增全局 Web session 版本，使所有用户现有 session 失效。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为全局配置更新错误，或当前 session 保存错误。

