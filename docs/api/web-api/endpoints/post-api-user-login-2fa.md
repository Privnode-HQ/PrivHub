---
method: POST
path: /api/user/login/2fa
auth: pending_login_session
handler: controller.Verify2FALogin
source: router/api-router.go:67
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/login/2fa`

完成密码登录后的两步验证。调用前必须先通过 `POST /api/user/login` 建立 pending login session。

## 请求体字段

- `code`: 字符串，必填。TOTP 数字验证码或备用码。

## 成功响应字段

验证成功后响应与普通登录相同：

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

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、`会话已过期，请重新登录`、`会话数据无效，请重新登录`、`用户不存在`、`用户未启用2FA`、`验证码或备用码错误，请重试`，或备用码校验错误。

