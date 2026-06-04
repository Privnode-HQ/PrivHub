---
method: POST
path: /api/user/impersonation/open_access
auth: user
handler: controller.OpenSelfServiceSupportAccess
source: router/api-router.go:91
request:
  body: none
response:
  success_http_status: 200
  verification_required_http_status: 403
  envelope: custom_success
---

# POST `/api/user/impersonation/open_access`

用户主动开放一次客服访问授权，24 小时内可由管理员激活一次。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: `已开放一次客服访问，24 小时内可由管理员激活一次`。
- `data.id`: 授权记录 ID。
- `data.state`: 授权状态。
- `data.granted_expires_at`: 授权过期时间。

## 失败响应

- HTTP 403 且 `code=VERIFICATION_REQUIRED`: 需要先调用安全验证。
- HTTP 200 且 `success=false`: 用户查询或服务层错误。

