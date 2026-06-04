---
method: POST
path: /api/creem/webhook
auth: creem_signature
handler: controller.CreemWebhook
source: router/api-router.go:57
request:
  headers:
    - creem-signature
  content_type: application/json
response:
  success_http_status: 200
  failure_http_status:
    - 400
    - 401
    - 500
  envelope: empty_status
---

# POST `/api/creem/webhook`

Creem checkout webhook。由 Creem 调用，用于处理产品充值支付完成事件。

## 请求头字段

- `creem-signature`: 字符串，生产环境必填。使用系统配置的 `CreemWebhookSecret` 做 HMAC-SHA256 校验；测试模式可能允许跳过。

## 请求体字段

- `id`: 字符串。Creem 事件 ID。
- `eventType`: 字符串。事件类型；当前处理 `checkout.completed`。
- `created_at`: 整数。事件创建时间。
- `object.request_id`: 字符串。本地充值单号。
- `object.order.id`: 字符串。Creem 订单 ID。
- `object.order.status`: 字符串。订单状态；完成处理要求为 `paid`。
- `object.order.type`: 字符串。订单类型；当前只处理 `onetime`。
- `object.order.amount_paid`: 整数。实付金额。
- `object.order.currency`: 字符串。币种。
- `object.product.name`: 字符串。产品名称。
- `object.customer.email`: 字符串。客户邮箱。
- `object.customer.name`: 字符串。客户名称。
- `object.metadata`: 对象。创建 checkout 时写入的元数据。

## 成功响应字段

返回 HTTP 200，响应体为空。未处理的事件类型也返回 HTTP 200。

## 失败响应

- HTTP 400: 读取请求体失败、JSON 解析失败、缺少本地订单号、本地订单不存在。
- HTTP 401: 缺少签名或签名校验失败。
- HTTP 500: Creem 充值处理失败。

