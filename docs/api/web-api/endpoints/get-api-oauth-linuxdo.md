---
method: GET
path: /api/oauth/linuxdo
auth: public_oauth_callback
handler: controller.LinuxdoOAuth
source: router/api-router.go:47
request:
  query_params:
    - state
    - code
    - error
    - error_description
response:
  success_http_status: 200
  forbidden_http_status: 403
  envelope: custom_or_login
---

# GET `/api/oauth/linuxdo`

LinuxDo OAuth 回调。未登录时用于登录/注册；已有登录 session 时转为绑定 LinuxDo 账号。

## 查询参数字段

- `state`: 字符串。正常回调必填，必须等于 session 中保存的 OAuth state。
- `code`: 字符串。正常回调必填，LinuxDo 授权码。
- `error`: 字符串，可选。LinuxDo 授权失败代码；存在时直接返回失败。
- `error_description`: 字符串，可选。授权失败说明。

## 成功响应字段

登录或注册成功后响应同 `POST /api/user/login` 的用户数据字段：`cah_id`、`username`、`display_name`、`role`、`status`、`group`、`email`、`force_password_reset`、`force_email_bind`、`required_actions`。

## 失败响应

- HTTP 403: `state is empty or not same`。
- HTTP 200 且 `success=false`: `message` 可能为 `error_description`、用户被禁用提示、注册关闭提示，或 LinuxDo API 错误。

