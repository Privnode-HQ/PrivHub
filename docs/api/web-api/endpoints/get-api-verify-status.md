---
method: GET
path: /api/verify/status
auth: user
handler: controller.GetVerificationStatus
source: router/api-router.go:61
request:
  body: none
response:
  success_http_status: 200
  unauthorized_http_status: 401
  envelope: custom_success
---

# GET `/api/verify/status`

查询当前 Web session 是否仍处于通用安全验证有效期内。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.verified`: 布尔值。`true` 表示当前 session 已验证且未过期；`false` 表示未验证或已过期。
- `data.expires_at`: 整数。仅 `verified=true` 时返回，验证状态过期时间 Unix 秒。

## 失败响应

- HTTP 401 且 `success=false`。
- `message`: `未登录`。

