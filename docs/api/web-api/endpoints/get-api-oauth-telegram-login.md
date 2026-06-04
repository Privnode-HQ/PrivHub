---
method: GET
path: /api/oauth/telegram/login
auth: public_telegram_callback
handler: controller.TelegramLogin
source: router/api-router.go:52
request:
  query_params: telegram_login_widget
response:
  success_http_status: 200
  envelope: custom_or_login
---

# GET `/api/oauth/telegram/login`

Telegram Login Widget 登录回调。

## 查询参数字段

Telegram Login Widget 会传入以下字段：

- `id`: 字符串，必填。Telegram 用户 ID。
- `first_name`: 字符串，可选。名。
- `last_name`: 字符串，可选。姓。
- `username`: 字符串，可选。Telegram 用户名。
- `photo_url`: 字符串，可选。头像 URL。
- `auth_date`: 字符串，必填。授权时间。
- `hash`: 字符串，必填。Telegram 签名。

## 成功响应字段

登录成功后响应同 `POST /api/user/login` 的用户数据字段。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `管理员未开启通过 Telegram 登录以及注册`、`无效的请求`，或用户查询错误。

