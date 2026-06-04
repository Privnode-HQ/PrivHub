---
method: GET
path: /api/admin/topups/:trade_no
auth: admin
handler: controller.GetAdminTopUpByTradeNo
source: router/api-router.go:24
request:
  path_params:
    - trade_no
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/admin/topups/:trade_no`

管理员按充值单号查询充值订单及关联用户安全摘要。

## 路径参数字段

- `trade_no`: 字符串，必填。充值订单号，前后空白会被去除。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.topup`: 充值订单对象。
- `data.topup.id`: 订单 ID。
- `data.topup.user_id`: 付款用户 ID。
- `data.topup.amount`: 购买额度单位数量。
- `data.topup.money`: 订单实际计费金额。
- `data.topup.trade_no`: 充值单号。
- `data.topup.payment_method`: 支付方式，例如 Epay 支付类型、`stripe` 或 `creem`。
- `data.topup.create_time`: 创建时间 Unix 秒。
- `data.topup.complete_time`: 完成时间 Unix 秒；未完成时为 `0`。
- `data.topup.status`: 订单状态，可能值 `pending`、`success`、`expired`。
- `data.topup.coupon_id`: 使用的充值优惠券 ID，未使用时为 `0`。
- `data.topup.coupon_name`: 使用的优惠券名称。
- `data.topup.original_money`: 折扣前金额。
- `data.topup.platform_discount`: 平台阶梯折扣金额。
- `data.topup.coupon_discount`: 优惠券抵扣金额。
- `data.topup.promotion_campaign_id`: 使用的促销活动 ID。
- `data.topup.promotion_code_id`: 使用的促销码 ID。
- `data.topup.promotion_code`: 使用的促销码文本。
- `data.topup.promotion_discount`: 促销码抵扣金额。
- `data.topup.promotion_redemption_id`: 促销码核销记录 ID。
- `data.topup.processing_fee`: 手续费字段；当前 Epay 费用已移除时通常为 `0`。
- `data.topup.pay_money`: 最终应付金额。
- `data.user`: 关联用户安全摘要，不包含 `password` 或 `access_token`。
- `data.user.id`: 用户 ID。
- `data.user.cah_id`: 用户 CAH ID。
- `data.user.username`: 用户名。
- `data.user.display_name`: 显示名。
- `data.user.role`: 用户角色，可能值 `0` 访客、`1` 普通用户、`5` 支持人员、`10` 管理员、`100` Root。
- `data.user.status`: 用户状态，可能值 `1` 启用、`2` 禁用。
- `data.user.email`: 绑定邮箱。
- `data.user.force_password_reset`: 是否被要求重置密码。
- `data.user.force_email_bind`: 是否被要求绑定邮箱。
- `data.user.github_id`: GitHub 绑定 ID。
- `data.user.discord_id`: Discord 绑定 ID。
- `data.user.oidc_id`: OIDC 绑定 ID。
- `data.user.wechat_id`: 微信绑定 ID。
- `data.user.telegram_id`: Telegram 绑定 ID。
- `data.user.linux_do_id`: LinuxDo 绑定 ID。
- `data.user.quota`: 当前剩余额度。
- `data.user.used_quota`: 已用额度。
- `data.user.request_count`: 请求次数。
- `data.user.group`: 用户分组。
- `data.user.aff_code`: 用户邀请码。
- `data.user.aff_count`: 邀请人数。
- `data.user.aff_quota`: 邀请奖励剩余额度。
- `data.user.aff_history_quota`: 邀请奖励历史额度。
- `data.user.inviter_id`: 邀请人用户 ID。
- `data.user.remark`: 管理员备注。
- `data.user.stripe_customer`: Stripe Customer ID。
- `data.user.deleted`: 布尔值，用户是否已软删除。
- `data.user.deleted_at`: 仅 `deleted=true` 时返回，软删除时间 Unix 秒。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `充值单号不能为空`、`充值订单不存在`，或用户查询错误。

