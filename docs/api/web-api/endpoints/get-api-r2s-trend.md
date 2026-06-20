---
method: GET
path: /api/r2s/trend
auth: admin
handler: controller.GetR2STrend
source: router/api-router.go:293
request:
  query_params:
    - start_time
    - end_time
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/r2s/trend`

管理员查看 R2S 已识别收入、成本、利润和利润率的按天趋势。该接口读取
本地收入识别记录，不会从上游供应商拉取数据。

## 查询参数字段

- `start_time`: 可选。Unix 秒，趋势开始时间。为空时默认最近 30 天。
- `end_time`: 可选。Unix 秒，趋势结束时间。为空时使用当前时间。

趋势按自然日聚合，查询范围最多 366 天。识别记录有业务周期时按
`period_end` 归档；没有业务周期的旧记录按 `created_time` 归档。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data[]`: 按时间升序返回的趋势点。
- `data[].time`: 当日开始时间 Unix 秒。
- `data[].label`: 面向图表展示的日期标签。
- `data[].recognized_revenue_amount`: 当日已识别收入的系统货币金额。
- `data[].recognized_cost_amount`: 当日已识别供应商成本的系统货币金额。
- `data[].recognized_profit_amount`: 当日已识别利润。
- `data[].profit_margin`: 当日利润率百分比。

## 失败响应

- `success`: `false`。
- `message`: 结束时间早于开始时间、查询范围超过 366 天，或数据库错误。
