---
method: GET
path: /api/r2s/recognition-records
auth: admin
handler: controller.GetR2SRecognitionRecords
source: router/api-router.go:307
request:
  query_params:
    - p
    - page_size
    - supplier_id
    - channel_id
    - keyword
    - start_time
    - end_time
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/r2s/recognition-records`

管理员分页查看 R2S 收入识别记录。每条记录保存供应商、渠道、上游分组
倍率、货币和汇率快照，因此之后修改供应商或渠道倍率不会改变历史利润。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。
- `supplier_id`: 可选。按 R2S 供应商过滤。
- `channel_id`: 可选。按 PrivHub 渠道过滤。
- `keyword`: 可选。按供应商名称快照、渠道名称快照或来源引用搜索。
- `start_time`: 可选。Unix 秒，按识别周期结束时间过滤。
- `end_time`: 可选。Unix 秒，按识别周期开始时间过滤。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 总数。
- `data.items[]`: 收入识别记录数组。
- `data.items[].id`: 收入识别记录 ID。
- `data.items[].source_type`: 来源类型，枚举为 `manual`、`promotion`。
- `data.items[].source_reference`: 外部或内部来源引用。
- `data.items[].supplier_id`: R2S 供应商 ID。
- `data.items[].supplier_name_snapshot`: 识别时的供应商名称快照。
- `data.items[].channel_id`: PrivHub 渠道 ID。
- `data.items[].channel_binding_id`: R2S 渠道绑定 ID。
- `data.items[].channel_name_snapshot`: 识别时的渠道名称快照。
- `data.items[].upstream_group_name_snapshot`: 识别时的上游分组快照。
- `data.items[].group_multiplier_snapshot`: 识别时的上游分组倍率快照。
- `data.items[].promotion_campaign_id`: 关联促销活动 ID。
- `data.items[].promotion_campaign_name`: 关联促销活动名称快照。
- `data.items[].currency_code`: 识别记录货币。
- `data.items[].exchange_rate`: 识别记录级别汇率。
- `data.items[].revenue_amount`: 原币收入金额。
- `data.items[].cost_amount`: 原币成本金额。
- `data.items[].system_revenue_amount`: 系统货币收入金额。
- `data.items[].system_cost_amount`: 系统货币成本金额。
- `data.items[].system_profit_amount`: 系统货币利润金额。
- `data.items[].profit_margin`: 利润率百分比。
- `data.items[].period_start`: 识别周期开始 Unix 秒。
- `data.items[].period_end`: 识别周期结束 Unix 秒。
- `data.items[].note`: 备注。
- `data.items[].created_by_admin_id`: 创建管理员用户 ID。
- `data.items[].created_time`: 创建时间 Unix 秒。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。
