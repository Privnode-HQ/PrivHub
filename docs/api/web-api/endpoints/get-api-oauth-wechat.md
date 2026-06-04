---
method: GET
path: /api/oauth/wechat
auth: public_oauth_callback
handler: controller.WeChatAuth
source: router/api-router.go:49
request:
  query_params:
    - code
response:
  success_http_status: 200
  envelope: custom_or_login
---

# GET `/api/oauth/wechat`

微信扫码登录/注册回调。

## 查询参数字段

- `code`: 字符串，必填。微信授权码，用于换取微信用户 ID。

## 成功响应字段

登录或注册成功后响应同 `POST /api/user/login` 的用户数据字段：`cah_id`、`username`、`display_name`、`role`、`status`、`group`、`email`、`force_password_reset`、`force_email_bind`、`required_actions`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `管理员未开启通过微信登录以及注册`、`用户已注销`、`管理员关闭了新用户注册`、用户被禁用提示，或微信服务错误。

