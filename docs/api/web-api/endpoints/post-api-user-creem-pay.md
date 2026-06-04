---
method: POST
path: /api/user/creem/pay
auth: user_rate_limited
handler: controller.RequestCreemPay
source: router/api-router.go:108
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: creem_legacy
---

# POST `/api/user/creem/pay`

创建 Creem 产品充值订单并返回 checkout URL。

## 请求体字段

- `product_id`: 字符串，必填。Creem 产品 ID，必须存在于系统 `CreemProducts` 配置。
- `payment_method`: 字符串，必填。必须为 `creem`。
- `coupon_id`: 整数。当前 Creem 不支持优惠券，非 `0` 会失败。

## 成功响应字段

- `message`: `success`。
- `data.checkout_url`: Creem checkout URL。
- `data.order_id`: 本地充值单号，也作为 Creem `request_id`。

## 失败响应

- `message`: `error`。
- `data`: 错误说明，例如 `read query error`、`参数错误`、`不支持的支付渠道`、`请选择产品`、`当前支付方式暂不支持优惠券`、`产品配置错误`、`产品不存在`、`创建订单失败`、`拉起支付失败`。

## 业务规则

- 金额和额度来自后台产品配置，不从客户端传入。
- 当前只处理一次性付款产品。

