---
method: POST
path: /api/user/impersonation/request/:token/reject
auth: user
handler: controller.RejectImpersonationRequest
source: router/api-router.go:89
request:
  path_params:
    - token
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/impersonation/request/:token/reject`

目标用户拒绝管理员模拟访问请求。

## 路径参数字段

- `token`: 字符串，必填。模拟访问批准 token。

## 成功响应字段

- `success`: `true`。
- `message`: `访问请求已拒绝`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无权拒绝该访问请求`，或服务层/数据库错误。

