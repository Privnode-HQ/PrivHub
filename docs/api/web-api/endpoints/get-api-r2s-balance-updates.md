---
method: GET
path: /api/r2s/balance-updates
auth: admin
handler: controller.GetR2SBalanceUpdates
source: router/api-router.go:307
request:
  query_params:
    - p
    - page_size
    - supplier_id
    - start_time
    - end_time
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/r2s/balance-updates`

管理员分页查看 R2S 供应商余额更新历史。余额更新可以来自手动更新，也
可以来自付款记录自动生成的余额变更。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。
- `supplier_id`: 可选。按 R2S 供应商过滤。
- `start_time`: 可选。Unix 秒，按创建时间起始过滤。
- `end_time`: 可选。Unix 秒，按创建时间结束过滤。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 总数。
- `data.items[]`: 余额更新记录数组。
- `data.items[].id`: 余额更新记录 ID。
- `data.items[].supplier_id`: R2S 供应商 ID。
- `data.items[].supplier_name_snapshot`: 更新时的供应商名称快照。
- `data.items[].update_type`: 更新类型，例如 `manual`、`prepaid`、
  `grant`、`refund`。
- `data.items[].balance_before`: 更新前余额。
- `data.items[].balance_after`: 更新后余额。
- `data.items[].delta_amount`: 余额变化量。
- `data.items[].currency_code`: 余额货币。
- `data.items[].exchange_rate`: 本次余额更新级别汇率。
- `data.items[].system_delta_amount`: 余额变化折算到系统货币后的金额。
- `data.items[].reminder_days_snapshot`: 更新时的提醒间隔快照。
- `data.items[].next_reminder_at`: 下一次提醒时间 Unix 秒，`0` 表示不提醒。
- `data.items[].note`: 备注。
- `data.items[].created_by_admin_id`: 创建管理员用户 ID。
- `data.items[].created_time`: 创建时间 Unix 秒。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。
