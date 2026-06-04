---
method: GET
path: /api/oauth/telegram/bind
auth: session_user
handler: controller.TelegramBind
source: router/api-router.go:53
request:
  query_params: telegram_login_widget
response:
  success_http_status: 302
  failure_http_status:
    - 200
    - 401
  envelope: redirect_or_custom
---

# GET `/api/oauth/telegram/bind`

为当前登录用户绑定 Telegram 账号。

## 查询参数字段

字段同 Telegram Login Widget：

- `id`: 字符串，必填。Telegram 用户 ID。
- `auth_date`: 字符串，必填。授权时间。
- `hash`: 字符串，必填。Telegram 签名。
- `first_name`、`last_name`、`username`、`photo_url`: 可选展示字段。

## 成功响应字段

成功后 HTTP 302 重定向到 `/console/personal`。

## 失败响应

- HTTP 401 且 `success=false`: 未登录或 session 无效。
- HTTP 200 且 `success=false`: 可能为 `管理员未开启通过 Telegram 登录以及注册`、`无效的请求`、`该 Telegram 账户已被绑定`、`用户已注销`，或用户更新错误。

