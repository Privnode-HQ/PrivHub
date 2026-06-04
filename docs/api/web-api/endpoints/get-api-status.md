---
method: GET
path: /api/status
auth: public
handler: controller.GetStatus
source: router/api-router.go:18
request:
  path_params: []
  query_params: []
  body: none
response:
  success_http_status: 200
  envelope: common
  data: Status
---

# GET `/api/status`

获取前端启动和登录页所需的系统状态、登录方式、控制台开关和展示配置。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.version`: 后端版本号。
- `data.start_time`: 后端启动时间，Unix 秒。
- `data.email_verification`: 是否启用邮箱验证。
- `data.github_oauth`: 是否启用 GitHub 登录。
- `data.github_client_id`: GitHub OAuth Client ID，未配置时为空。
- `data.discord_oauth`: 是否启用 Discord 登录。
- `data.discord_client_id`: Discord OAuth Client ID，未配置时为空。
- `data.linuxdo_oauth`: 是否启用 LinuxDo 登录。
- `data.linuxdo_client_id`: LinuxDo OAuth Client ID，未配置时为空。
- `data.linuxdo_minimum_trust_level`: LinuxDo 最低信任等级。
- `data.telegram_oauth`: 是否启用 Telegram 登录。
- `data.telegram_bot_name`: Telegram 机器人名称。
- `data.system_name`: 系统展示名称。
- `data.logo`: Logo URL 或空字符串。
- `data.footer_html`: 页脚 HTML。
- `data.wechat_qrcode`: 微信二维码图片 URL。
- `data.wechat_login`: 是否启用微信登录。
- `data.server_address`: 对外服务地址。
- `data.turnstile_check`: 是否启用 Turnstile 校验。
- `data.turnstile_site_key`: Turnstile Site Key。
- `data.top_up_link`: 自定义充值链接。
- `data.docs_link`: 自定义文档链接。
- `data.quota_per_unit`: 金额到额度的换算单位。
- `data.display_in_currency`: 是否按货币展示额度。
- `data.quota_display_type`: 额度展示类型，例如全局 CNY/USD 配置。
- `data.enable_batch_update`: 是否启用批量更新。
- `data.enable_drawing`: 是否启用绘图模块。
- `data.enable_task`: 是否启用任务模块。
- `data.enable_data_export`: 是否启用数据导出。
- `data.data_export_default_time`: 数据导出默认时间粒度。
- `data.default_collapse_sidebar`: 是否默认折叠侧边栏。
- `data.mj_notify_enabled`: 是否启用 Midjourney 通知。
- `data.chats`: 聊天入口配置。
- `data.demo_site_enabled`: 是否启用演示站点模式。
- `data.self_use_mode_enabled`: 是否启用自用模式。
- `data.default_use_auto_group`: 是否默认使用 auto 分组。
- `data.price`: 默认价格配置。
- `data.stripe_unit_price`: Stripe 单价配置。
- `data.api_info_enabled`: 控制台 API 信息面板是否启用。
- `data.uptime_kuma_enabled`: Uptime Kuma 面板是否启用。
- `data.announcements_enabled`: 公告面板是否启用；当前固定为 `false`。
- `data.faq_enabled`: FAQ 面板是否启用。
- `data.HeaderNavModules`: 顶部导航模块配置字符串。
- `data.SidebarModulesAdmin`: 管理员侧边栏模块配置字符串。
- `data.oidc_enabled`: 是否启用 OIDC。
- `data.oidc_client_id`: OIDC Client ID。
- `data.oidc_authorization_endpoint`: OIDC 授权端点。
- `data.passkey_login`: 是否启用 Passkey 登录。
- `data.passkey_display_name`: Passkey RP 展示名称。
- `data.passkey_rp_id`: Passkey RP ID。
- `data.passkey_origins`: Passkey 允许来源列表。
- `data.passkey_allow_insecure`: 是否允许不安全来源。
- `data.passkey_user_verification`: Passkey 用户验证偏好。
- `data.passkey_attachment`: Passkey 设备附件偏好。
- `data.setup`: 系统是否完成初始化。
- `data.user_agreement_enabled`: 用户协议是否非空。
- `data.privacy_policy_enabled`: 隐私政策是否非空。
- `data.api_info`: 仅在 `api_info_enabled=true` 时返回，内容为 API 信息面板配置。
- `data.faq`: 仅在 `faq_enabled=true` 时返回，内容为 FAQ 配置。

## 错误响应

当前控制器没有显式错误分支。

