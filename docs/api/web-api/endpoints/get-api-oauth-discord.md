---
method: GET
path: /api/oauth/discord
auth: public_oauth_callback
handler: controller.DiscordOAuth
source: router/api-router.go:45
request:
  query_params:
    - state
    - code
response:
  success_http_status: 200
  forbidden_http_status: 403
  envelope: custom_or_login
---

# GET `/api/oauth/discord`

Discord OAuth 回调。未登录时用于登录/注册；已有登录 session 时转为绑定 Discord 账号。

## 查询参数字段

- `state`: 字符串，必填。必须等于 session 中保存的 OAuth state。
- `code`: 字符串，必填。Discord 授权码。

## 成功响应字段

登录或注册成功后响应同 `POST /api/user/login` 的用户数据字段：`cah_id`、`username`、`display_name`、`role`、`status`、`group`、`email`、`force_password_reset`、`force_email_bind`、`required_actions`。

## 失败响应

- HTTP 403: `state is empty or not same`。
- HTTP 200 且 `success=false`: 可能为 `管理员未开启通过 Discord 登录以及注册`、`用户已注销`、`管理员关闭了新用户注册`、用户被禁用提示，或 Discord API 错误。

