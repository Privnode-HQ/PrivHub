---
method: POST
path: /api/user/topup/complete
auth: admin
handler: controller.AdminCompleteTopUp
source: router/api-router.go:133
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/user/topup/complete`

管理员按充值单号手动补单，完成待支付订单并给用户充值。

## 请求体字段

- `trade_no`: 字符串，必填。充值单号。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: `null`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`，或订单不存在、订单状态错误、充值处理错误。

## 业务规则

- 该接口使用订单级锁，防止同一单号并发补单。
- 成功后会尝试清理 Stripe Checkout 临时优惠券。

