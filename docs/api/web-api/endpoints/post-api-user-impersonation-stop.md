---
method: POST
path: /api/user/impersonation/stop
auth: user
handler: controller.StopUserImpersonation
source: router/api-router.go:86
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/impersonation/stop`

停止当前管理员模拟访问 session，并返回原管理员的当前用户对象。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

如果当前不在模拟访问中，响应等同 `GET /api/user/self`。

停止成功时：

- `success`: `true`。
- `message`: 空字符串。
- `data.user`: 原管理员用户对象，字段同 `GET /api/user/self`。

## 失败响应

- `success`: `false`。
- `message`: 停止模拟访问或用户查询错误。

