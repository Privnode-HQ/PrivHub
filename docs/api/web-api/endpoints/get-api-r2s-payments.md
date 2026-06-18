---
method: GET
path: /api/r2s/payments
auth: admin
handler: controller.GetR2SPayments
source: router/api-router.go:303
request:
  query_params:
    - p
    - page_size
    - supplier_id
    - keyword
    - start_time
    - end_time
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/r2s/payments`

管理员分页查看向供应商付款历史。付款记录包含付款类型、货币、汇率、
收据要求和付款前后余额快照。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。
- `supplier_id`: 可选。按 R2S 供应商过滤。
- `keyword`: 可选。按供应商名称快照、付款类型或备注搜索。
- `start_time`: 可选。Unix 秒，按 `paid_at` 起始时间过滤。
- `end_time`: 可选。Unix 秒，按 `paid_at` 结束时间过滤。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 总数。
- `data.items[]`: 付款记录数组。
- `data.items[].id`: 付款记录 ID。
- `data.items[].supplier_id`: R2S 供应商 ID。
- `data.items[].supplier_name_snapshot`: 付款时的供应商名称快照。
- `data.items[].payment_type`: 付款类型，枚举为 `prepaid`、`postpaid`、
  `grant`、`refund`、`adjustment`。
- `data.items[].amount`: 付款货币金额。
- `data.items[].currency_code`: 付款货币，支持外币和 `USDT` 等。
- `data.items[].exchange_rate`: 本次付款级别汇率。
- `data.items[].system_amount`: 折算到系统货币后的金额。
- `data.items[].balance_before`: 付款前供应商余额。
- `data.items[].balance_after`: 付款后供应商余额。
- `data.items[].receipt_url`: 收据或支付截图地址。
- `data.items[].receipt_required`: 创建时系统是否要求收据。
- `data.items[].note`: 备注。
- `data.items[].paid_at`: 付款时间 Unix 秒。
- `data.items[].created_by_admin_id`: 创建管理员用户 ID。
- `data.items[].created_time`: 创建时间 Unix 秒。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。
