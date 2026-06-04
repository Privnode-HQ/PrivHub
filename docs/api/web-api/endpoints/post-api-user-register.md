---
method: POST
path: /api/user/register
auth: public_turnstile_rate_limited
handler: controller.Register
source: router/api-router.go:65
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/register`

注册普通用户账号。是否允许注册、是否允许密码注册、是否需要邮箱验证码由系统配置控制。

## 请求体字段

- `username`: 字符串，必填。用户名，必须满足 `model.User` 校验，最长 20 个字符。
- `password`: 字符串，必填。密码，必须满足 `model.User` 校验，长度 8 到 20。
- `email`: 字符串。邮箱验证开启时必填。
- `verification_code`: 字符串。邮箱验证开启时必填，必须匹配 `GET /api/verification` 发送的验证码。
- `aff_code`: 字符串，可选。邀请人的邀请码；注册成功后用于建立邀请关系。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `管理员关闭了新用户注册`、`管理员关闭了通过密码进行注册，请使用第三方账户验证的形式进行注册`、`无效的参数`、`输入不合法 ...`、`管理员开启了邮箱验证，请输入邮箱地址和验证码`、`验证码错误或已过期`、`用户名已存在，或已注销`、`数据库错误，请稍后重试`、默认令牌创建错误。

## 业务规则

- 新用户角色固定为普通用户。
- `display_name` 不从请求体采纳，当前实现使用用户名作为显示名。
- 如果启用默认令牌，会创建一个初始令牌；启用默认 auto 分组时令牌分组为 `auto`。

