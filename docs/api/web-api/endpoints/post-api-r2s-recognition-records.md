---
method: POST
path: /api/r2s/recognition-records
auth: admin
handler: controller.CreateR2SRecognitionRecord
source: router/api-router.go:312
request:
  body:
    content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/r2s/recognition-records`

管理员创建收入识别记录。创建时会从当前 R2S 供应商和渠道绑定读取名称、
上游分组倍率、货币和汇率并写入快照，保证之后本地或上游倍率变更时，
历史数据仍按当时快照计算。

## 请求体字段

- `source_type`: 可选。`manual`、`promotion` 或 `usage`；为空且传入
  `promotion_campaign_id` 时使用 `promotion`，否则使用 `manual`。管理员手动
  创建普通识别记录时通常不需要传 `usage`，历史消费日志同步会自动使用
  `usage` 来源。
- `source_reference`: 可选。外部或内部来源引用。
- `supplier_id`: R2S 供应商 ID，必填。
- `channel_id`: 可选。PrivHub 渠道 ID。未传 `channel_binding_id` 时，
  可用该字段查找启用中的 R2S 渠道绑定。
- `channel_binding_id`: 可选。R2S 渠道绑定 ID，优先用于快照渠道和倍率。
- `promotion_campaign_id`: 可选。关联促销活动 ID。
- `currency_code`: 可选。识别记录货币；为空时使用供应商默认货币。
- `exchange_rate`: 可选。识别记录级别汇率；为空时使用供应商默认汇率。
- `revenue_amount`: 原币收入金额，不能小于 `0`。
- `cost_amount`: 原币成本金额，不能小于 `0`。
- `group_multiplier_snapshot`: 可选。没有绑定但传入 `channel_id` 时可作为
  手动倍率快照；为空时为 `1`。
- `period_start`: 可选。识别周期开始 Unix 秒。
- `period_end`: 可选。识别周期结束 Unix 秒，不能早于 `period_start`。
- `note`: 可选。备注，最多 500 个字符。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 新建收入识别记录，字段同 GET
  `/api/r2s/recognition-records` 的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: 参数错误、供应商不存在、渠道或绑定不存在、促销活动不存在、
  收入或成本小于 `0`、货币或汇率不正确、结束时间早于开始时间，或数据库
  错误。
