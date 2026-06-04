---
method: GET
path: /api/oauth/email/bind
auth: session_user
handler: controller.EmailBind
source: router/api-router.go:51
request:
  query_params:
    - email
    - code
response:
  success_http_status: 200
  unauthorized_http_status: 401
  envelope: custom_success
---

# GET `/api/oauth/email/bind`

为当前登录用户绑定邮箱。

## 查询参数字段

- `email`: 字符串，必填。要绑定的邮箱。
- `code`: 字符串，必填。邮箱验证码，必须匹配邮箱验证用途。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- HTTP 401 且 `success=false`: 未登录或 session 无效。
- HTTP 200 且 `success=false`: `验证码错误或已过期`，或用户查询/更新错误。

