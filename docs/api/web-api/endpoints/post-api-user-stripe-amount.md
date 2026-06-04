---
method: POST
path: /api/user/stripe/amount
auth: user
handler: controller.RequestStripeAmount
source: router/api-router.go:107
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: stripe_amount_legacy
---

# POST `/api/user/stripe/amount`

计算 Stripe 充值最终应付金额，不创建订单。

## 请求体字段

- `amount`: 整数，必填。充值额度数量，必须满足 Stripe 最小充值。
- `payment_method`: 字符串，可选。控制器计算时固定按 `stripe` 处理。
- `currency_code`: 字符串，可选。请求币种。
- `coupon_id`: 整数。要尝试使用的优惠券 ID。

## 成功响应字段

- `message`: `success`。
- `data`: 字符串。保留两位小数的最终应付金额。

## 失败响应

- `message`: `error`。
- `data`: 错误说明，例如 `参数错误`、`充值数量不能小于 ...`、`用户不存在`、`充值金额过低`，或币种/优惠券错误。

