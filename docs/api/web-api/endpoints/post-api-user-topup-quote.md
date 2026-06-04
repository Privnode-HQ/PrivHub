---
method: POST
path: /api/user/topup/quote
auth: user
handler: controller.RequestTopUpQuote
source: router/api-router.go:101
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/user/topup/quote`

预览充值应付金额、平台折扣、优惠券/促销码抵扣和可用优惠券列表。创建支付订单前应以该接口结果为准。

## 请求体字段

- `payment_method`: 字符串，必填。支付方式；可能为 Epay 支付类型、`stripe`、`creem`。
- `amount`: 整数。充值额度数量；Epay/Stripe 使用。
- `product_id`: 字符串。Creem 产品 ID；`payment_method=creem` 时使用。
- `currency_code`: 字符串。请求币种；Stripe 可用，后端会规范化。
- `coupon_id`: 整数。要尝试使用的优惠券 ID；与 `promotion_code` 不能同时使用。
- `promotion_code`: 字符串。要尝试使用的促销码；后端会转大写并去除空白。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.payment_method`: 请求支付方式。
- `data.amount`: 充值额度数量。
- `data.product_id`: Creem 产品 ID。
- `data.product_name`: Creem 产品名。
- `data.product_quota`: Creem 产品包含额度。
- `data.currency_code`: 本次报价币种。
- `data.supported_currency_codes`: 支持币种数组，主要用于 Stripe。
- `data.original_amount`: 折扣前金额。
- `data.base_payable_amount`: 优惠券/促销码抵扣前的基础应付金额。
- `data.platform_discount_amount`: 平台折扣金额。
- `data.coupon_discount_amount`: 优惠券抵扣金额。
- `data.promotion_discount_amount`: 促销码抵扣金额。
- `data.final_payable_amount`: 最终应付金额。
- `data.min_payable_threshold`: 最低应付阈值。
- `data.selected_coupon_id`: 实际选中的优惠券 ID。
- `data.promotion_campaign_id`: 命中的促销活动 ID。
- `data.promotion_code_id`: 命中的促销码 ID。
- `data.promotion_code`: 命中的促销码文本。
- `data.promotion_rule`: 命中的促销规则展示文本。
- `data.ineligible_reason`: 请求的优惠券/促销码不可用时的原因。
- `data.available_coupons`: 当前可用优惠券数组。
- `data.available_coupons[].id`: 优惠券 ID。
- `data.available_coupons[].name`: 优惠券名称。
- `data.available_coupons[].deduction_amount`: 抵扣金额。
- `data.available_coupons[].currency_code`: 优惠券币种。
- `data.available_coupons[].status`: 优惠券状态，可能值 `available`、`reserved`、`used`、`expired`、`revoked`。
- `data.available_coupons[].expires_at`: 过期时间 Unix 秒，`0` 表示不按时间过期。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、`用户不存在`、`请选择支付方式`、`优惠券和促销码不能同时使用`、`支付方式不存在`、`充值金额过低`、`当前支付方式暂不支持促销码`，或产品/优惠券/促销码错误。

