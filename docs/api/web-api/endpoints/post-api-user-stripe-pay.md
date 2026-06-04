---
method: POST
path: /api/user/stripe/pay
auth: user_rate_limited
handler: controller.RequestStripePay
source: router/api-router.go:106
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: stripe_legacy
---

# POST `/api/user/stripe/pay`

创建 Stripe Checkout 充值订单并返回支付链接。

## 请求体字段

- `amount`: 整数，必填。充值额度数量，必须满足 Stripe 最小充值。
- `payment_method`: 字符串，必填。必须为 `stripe`。
- `currency_code`: 字符串，可选。请求币种；后端会按 Stripe price 支持币种解析。
- `coupon_id`: 整数。要使用的优惠券 ID。Stripe 当前不支持促销码。

## 成功响应字段

- `message`: `success`。
- `data.pay_link`: Stripe Checkout URL。

## 失败响应

- `message`: `error` 或错误文本。
- `data`: 错误说明，例如 `参数错误`、`不支持的支付渠道`、`充值数量不能小于 ...`、`用户不存在`、`充值金额过低`、`拉起支付失败`、`创建订单失败`。

## 业务规则

- 订单号保存为 `ref_` 前缀引用 ID。
- 创建订单时状态为 `pending`；Stripe webhook 完成后才充值。
- 使用优惠券或平台折扣时，可能创建 Stripe 临时 coupon；订单失败会尝试清理。

