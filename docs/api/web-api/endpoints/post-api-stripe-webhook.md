---
method: POST
path: /api/stripe/webhook
auth: stripe_signature
handler: controller.StripeWebhook
source: router/api-router.go:56
request:
  headers:
    - Stripe-Signature
  content_type: application/json
response:
  success_http_status: 200
  failure_http_status:
    - 400
    - 503
  envelope: empty_status
---

# POST `/api/stripe/webhook`

Stripe Checkout webhook。由 Stripe 调用，用于完成、失败或过期充值订单。

## 请求头字段

- `Stripe-Signature`: 字符串，必填。Stripe webhook 签名，使用系统配置的 `StripeWebhookSecret` 校验。

## 请求体字段

请求体为 Stripe Event JSON：

- `type`: 字符串。事件类型；处理 `checkout.session.completed`、`checkout.session.async_payment_failed`、`checkout.session.expired`。
- `data.object.client_reference_id`: 字符串。本地充值单号。
- `data.object.customer`: 字符串。Stripe Customer ID。
- `data.object.status`: 字符串。Checkout session 状态，完成事件要求为 `complete`。
- `data.object.amount_total`: 整数。支付金额最小货币单位。
- `data.object.currency`: 字符串。支付币种。

## 成功响应字段

返回 HTTP 200，响应体为空。

## 失败响应

- HTTP 503: 读取请求体失败。
- HTTP 400: Stripe 签名校验失败。

