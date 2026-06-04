---
method: GET
path: /api/user/topup/info
auth: user
handler: controller.GetTopUpInfo
source: router/api-router.go:100
request:
  body: none
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/user/topup/info`

获取当前用户充值页所需的支付方式、最小充值、金额选项、平台折扣和优惠券摘要。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.enable_online_topup`: 布尔值，Epay 在线充值是否可用。
- `data.enable_stripe_topup`: 布尔值，Stripe 充值是否可用。
- `data.enable_creem_topup`: 布尔值，Creem 充值是否可用。
- `data.creem_products`: 字符串，Creem 产品 JSON 配置。
- `data.pay_methods`: 支付方式数组。
- `data.pay_methods[].name`: 支付方式展示名。
- `data.pay_methods[].type`: 支付方式类型，例如 Epay 类型或 `stripe`。
- `data.pay_methods[].color`: 前端展示颜色。
- `data.pay_methods[].min_topup`: 该方式最小充值，字符串或数字文本。
- `data.min_topup`: Epay/余额计算最小充值数量。
- `data.stripe_min_topup`: Stripe 最小充值数量。
- `data.amount_options`: 预设充值数量选项。
- `data.discount`: 平台金额折扣规则数值。
- `data.coupon_summary`: 可选。当前用户优惠券摘要。
- `data.coupon_summary.has_available_coupon`: 是否有可用优惠券。
- `data.coupon_summary.available_count`: 可用优惠券数量。
- `data.coupon_summary.strongest_discount_amount`: 最大抵扣金额。
- `data.coupon_summary.strongest_currency_code`: 最大抵扣金额币种。
- `data.coupon_summary.has_mixed_currency`: 是否存在多币种优惠券。
- `data.coupon_summary.banner_message`: 前端提示文案。

## 失败响应

当前控制器没有显式错误分支；优惠券摘要查询失败时会省略 `coupon_summary`。

