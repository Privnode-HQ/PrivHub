---
method: POST
path: /api/r2s/payments
auth: admin
handler: controller.CreateR2SPayment
source: router/api-router.go:304
request:
  body:
    content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/r2s/payments`

管理员新增向供应商付款历史。该接口是手动录入，不会自动向上游付款或
查询上游余额。

## 请求体字段

- `supplier_id`: R2S 供应商 ID，必填。
- `payment_type`: 付款类型，枚举为 `prepaid`、`postpaid`、`grant`、
  `refund`、`adjustment`。
- `amount`: 付款货币金额，必须大于 `0`。
- `currency_code`: 可选。付款货币，支持外币和 `USDT` 等；为空时使用
  供应商默认货币。
- `exchange_rate`: 可选。本次付款级别汇率；为空时使用供应商默认汇率。
- `balance_after`: 可选。显式指定付款后供应商余额。未提供时，`prepaid`、
  `grant`、`adjustment` 增加余额，`refund` 减少余额，`postpaid`
  不改变余额。
- `receipt_url`: 可选。收据或支付截图地址；系统设置要求时必填。
- `note`: 可选。备注，最多 500 个字符。
- `paid_at`: 可选。付款时间 Unix 秒；为空时使用当前时间。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.payment`: 新建付款记录，字段同 GET `/api/r2s/payments`
  的 `data.items[]`。
- `data.balance_update`: 自动生成的余额更新记录；如果付款未改变余额则为
  `null`。

## 失败响应

- `success`: `false`。
- `message`: 参数错误、供应商不存在、付款类型不正确、金额或汇率不正确、
  当前设置要求上传收据或支付截图，或数据库错误。
