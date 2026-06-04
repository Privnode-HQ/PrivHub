---
method: DELETE
path: /api/user/impersonation/open_access
auth: user
handler: controller.CloseSelfServiceSupportAccess
source: router/api-router.go:92
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# DELETE `/api/user/impersonation/open_access`

关闭当前用户未使用的客服访问授权。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: `未使用的客服访问已关闭`。

## 失败响应

- `success`: `false`。
- `message`: 服务层错误，例如没有可关闭授权或数据库错误。

