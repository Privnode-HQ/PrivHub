---
method: GET
path: /api/user/self
auth: user
handler: controller.GetSelf
source: router/api-router.go:80
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/user/self`

获取当前登录用户资料、权限、侧边栏配置和当前 session 状态。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.cah_id`: 用户 CAH ID。
- `data.username`: 用户名。
- `data.display_name`: 显示名。
- `data.role`: 用户角色，可能值 `0`、`1`、`5`、`10`、`100`。
- `data.status`: 用户状态，`1` 启用，`2` 禁用。
- `data.email`: 绑定邮箱。
- `data.github_id`: GitHub 绑定 ID。
- `data.discord_id`: Discord 绑定 ID。
- `data.oidc_id`: OIDC 绑定 ID。
- `data.wechat_id`: 微信绑定 ID。
- `data.telegram_id`: Telegram 绑定 ID。
- `data.group`: 用户分组。
- `data.quota`: 当前剩余额度。
- `data.used_quota`: 已用额度。
- `data.request_count`: 请求次数。
- `data.aff_code`: 邀请码。
- `data.aff_count`: 邀请人数。
- `data.aff_quota`: 邀请奖励剩余额度。
- `data.aff_history_quota`: 邀请奖励历史额度。
- `data.linux_do_id`: LinuxDo 绑定 ID。
- `data.setting`: 用户设置 JSON 字符串。
- `data.stripe_customer`: Stripe Customer ID。
- `data.sidebar_modules`: 从用户设置解析出的侧边栏模块配置字符串。
- `data.permissions`: 当前角色可配置模块权限对象。
- `data.force_password_reset`: 是否需要强制重置密码。
- `data.force_email_bind`: 是否需要强制绑定邮箱。
- `data.required_actions`: 当前用户必须完成的动作列表。
- `data.require_display_name_enabled`: 系统是否要求用户设置显示名。
- `data.require_email_binding_enabled`: 系统是否要求用户绑定邮箱。
- `data.inviter_cah_id`: 可选。存在邀请人时返回邀请人 CAH ID。
- `data.impersonation`: 可选。当前处于管理员模拟访问时返回。
- `data.impersonation.active`: 是否正在模拟访问。
- `data.impersonation.grant_id`: 模拟访问授权 ID。
- `data.impersonation.read_only`: 是否只读。
- `data.impersonation.break_glass`: 是否为 break glass 模式。
- `data.impersonation.started_at`: 开始时间。
- `data.impersonation.expires_at`: session 过期时间。
- `data.impersonation.operator_id`: 原管理员用户 ID。
- `data.impersonation.operator_username`: 原管理员用户名。
- `data.impersonation.operator_cah_id`: 原管理员 CAH ID。
- `data.access_link_session`: 可选。当前处于 Access Link 会话时返回。
- `data.access_link_session.active`: 是否激活。
- `data.access_link_session.grant_id`: 授权 ID。
- `data.access_link_session.expires_at`: 过期时间。
- `data.break_glass_alert`: 可选。最近 break glass 访问提醒。

## 失败响应

- `success`: `false`。
- `message`: 用户查询错误或鉴权错误。

