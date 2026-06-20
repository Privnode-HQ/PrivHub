---
method: GET
path: /api/r2s/summary
auth: admin
handler: controller.GetR2SSummary
source: router/api-router.go:292
request:
  query_params:
    - start_time
    - end_time
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/r2s/summary`

管理员查看 R2S 收入、成本、利润、供应商余额和余额提醒概览。利润数据
来自已经写入的本地收入识别记录，不会从上游自动拉取。

## 查询参数字段

- `start_time`: 可选。Unix 秒，统计周期开始。用于筛选付款和识别记录。
- `end_time`: 可选。Unix 秒，统计周期结束。用于筛选付款和识别记录。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.system_currency_code`: 当前系统额度展示货币。
- `data.recognized_revenue_amount`: 已识别收入的系统货币金额。
- `data.recognized_cost_amount`: 已识别供应商成本的系统货币金额。
- `data.recognized_profit_amount`: 已识别利润，等于收入减成本。
- `data.profit_margin`: 利润率百分比。
- `data.payment_system_amount`: 周期内向供应商付款的系统货币金额。
- `data.supplier_balance_amount`: 当前启用中供应商余额的系统货币合计；
  停用供应商不会进入余额统计。
- `data.supplier_count`: R2S 供应商总数。
- `data.active_supplier_count`: 启用中的 R2S 供应商数。
- `data.channel_binding_count`: 启用中的供应商渠道绑定数。
- `data.reminder_due_count`: 已到达余额提醒时间的启用中供应商数。
- `data.last_recognition_time`: 最近一次识别记录创建或更新时间 Unix 秒。
- `data.last_payment_time`: 最近一次付款记录创建时间 Unix 秒。
- `data.last_balance_update_time`: 最近一次余额更新记录创建时间 Unix 秒。
- `data.updated_at`: 看板数据最近更新时间 Unix 秒，取识别、付款和余额更新
  时间中的最大值。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。
