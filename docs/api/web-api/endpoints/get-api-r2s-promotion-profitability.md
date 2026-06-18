---
method: GET
path: /api/r2s/promotion-profitability
auth: admin
handler: controller.GetR2SPromotionProfitability
source: router/api-router.go:293
request:
  query_params:
    - campaign_id
    - start_time
    - end_time
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/r2s/promotion-profitability`

管理员查看促销活动盈利性。收入来自本地促销核销记录，成本来自已经写入
且关联 `promotion_campaign_id` 的 R2S 收入识别记录。促销实收金额使用促销
核销记录货币；已识别成本使用系统货币。只有两者货币一致时才计算利润和
利润率。

## 查询参数字段

- `campaign_id`: 可选。促销活动 ID；为空时返回所有有核销记录的活动。
- `start_time`: 可选。Unix 秒，按促销核销时间和识别周期过滤。
- `end_time`: 可选。Unix 秒，按促销核销时间和识别周期过滤。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data[]`: 促销活动盈利性数组。
- `data[].campaign_id`: 促销活动 ID。
- `data[].campaign_name`: 促销活动名称快照。
- `data[].top_up_count`: 促销核销次数。
- `data[].gross_revenue_amount`: 优惠前充值收入。
- `data[].discount_amount`: 促销优惠金额。
- `data[].net_revenue_amount`: 优惠后实收金额。
- `data[].recognized_cost_amount`: 已识别供应商成本。
- `data[].profit_amount`: 促销利润，等于实收金额减已识别成本。
- `data[].profit_margin`: 促销利润率百分比。
- `data[].currency_code`: 促销活动货币。
- `data[].system_currency_code`: 当前系统货币。
- `data[].profit_calculated`: 是否已计算利润；当促销货币与系统货币不一致
  时为 `false`，此时 `profit_amount` 与 `profit_margin` 不应作为盈利性
  结论使用。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。
