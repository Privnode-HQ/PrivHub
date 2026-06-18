---
method: GET
path: /api/r2s/suppliers
auth: admin
handler: controller.GetR2SSuppliers
source: router/api-router.go:294
request:
  query_params:
    - p
    - page_size
    - keyword
    - status
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/r2s/suppliers`

管理员分页查看 R2S 上游供应商。R2S 供应商是收入识别用的本地财务对象，
不同于模型元数据供应商。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，受通用分页上限限制。
- `keyword`: 可选。按供应商 ID、名称或说明搜索。
- `status`: 可选。`active` 或 `disabled`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 总数。
- `data.items[]`: R2S 供应商数组。
- `data.items[].id`: 供应商 ID。
- `data.items[].name`: 供应商名称。
- `data.items[].description`: 供应商说明。
- `data.items[].status`: 状态，枚举为 `active`、`disabled`。
- `data.items[].default_currency_code`: 供应商默认货币。
- `data.items[].default_exchange_rate`: 默认汇率，表示供应商货币到系统
  货币的换算比例。
- `data.items[].balance_amount`: 当前供应商余额。
- `data.items[].balance_currency_code`: 当前余额货币。
- `data.items[].system_balance_amount`: 当前余额折算到系统货币后的金额。
- `data.items[].balance_updated_time`: 最近余额更新时间 Unix 秒。
- `data.items[].balance_reminder_days`: 余额提醒间隔天数。
- `data.items[].next_balance_reminder_at`: 下一次余额提醒时间 Unix 秒。
- `data.items[].channel_count`: 启用中的关联渠道数量。
- `data.items[].last_payment_time`: 最近付款时间 Unix 秒，`0` 表示无记录。
- `data.items[].created_time`: 创建时间 Unix 秒。
- `data.items[].updated_time`: 更新时间 Unix 秒。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。
