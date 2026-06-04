---
method: GET
path: /api/oauth/wechat/bind
auth: session_user
handler: controller.WeChatBind
source: router/api-router.go:50
request:
  query_params:
    - code
response:
  success_http_status: 302
  failure_http_status: 200
  envelope: redirect_or_custom
---

# GET `/api/oauth/wechat/bind`

为当前登录用户绑定微信账号。

## 查询参数字段

- `code`: 字符串，必填。微信授权码，用于换取微信用户 ID。

## 成功响应字段

成功后 HTTP 302 重定向到 `/console/personal`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `管理员未开启通过微信登录以及注册`、`该微信账户已被绑定`、`用户已注销`、未登录/会话错误，或微信服务错误。

