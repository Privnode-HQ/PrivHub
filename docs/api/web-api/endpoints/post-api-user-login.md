---
method: POST
path: /api/user/login
auth: public_turnstile_rate_limited
handler: controller.Login
source: router/api-router.go:66
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/login`

使用用户名和密码登录 Web 控制台。成功后写入 Web session。

## 请求体字段

- `username`: 字符串，必填。用户名。
- `password`: 字符串，必填。明文密码，后端会与密码哈希校验。

## 成功响应字段

普通登录成功：

- `success`: `true`。
- `message`: 空字符串。
- `data.cah_id`: 用户 CAH ID。
- `data.username`: 用户名。
- `data.display_name`: 显示名。
- `data.role`: 用户角色，可能值 `0`、`1`、`5`、`10`、`100`。
- `data.status`: 用户状态，`1` 启用，`2` 禁用。
- `data.group`: 用户分组。
- `data.email`: 绑定邮箱。
- `data.force_password_reset`: 是否需要强制重置密码。
- `data.force_email_bind`: 是否需要强制绑定邮箱。
- `data.required_actions`: 当前用户必须完成的动作列表。

需要 2FA 时：

- `success`: `true`。
- `message`: `请输入两步验证码`。
- `data.require_2fa`: `true`，表示需要继续调用 `/api/user/login/2fa`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `管理员关闭了密码登录`、`无效的参数`、`无法保存会话信息，请重试`，或用户校验错误。

