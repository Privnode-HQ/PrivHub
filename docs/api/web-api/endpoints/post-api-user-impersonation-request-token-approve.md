---
method: POST
path: /api/user/impersonation/request/:token/approve
auth: user
handler: controller.ApproveImpersonationRequest
source: router/api-router.go:88
request:
  path_params:
    - token
response:
  success_http_status: 200
  verification_required_http_status: 403
  envelope: custom_success
---

# POST `/api/user/impersonation/request/:token/approve`

目标用户批准管理员模拟访问请求。

## 路径参数字段

- `token`: 字符串，必填。模拟访问批准 token。

## 成功响应字段

- `success`: `true`。
- `message`: `访问请求已批准，客服需在 24 小时内开始使用`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无权批准该访问请求`，或服务层返回的状态错误。
- `code`: 当 HTTP 403 且需要安全验证时为 `VERIFICATION_REQUIRED`。

