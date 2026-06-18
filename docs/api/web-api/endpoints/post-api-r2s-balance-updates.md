---
method: POST
path: /api/r2s/balance-updates
auth: admin
handler: controller.CreateR2SBalanceUpdate
source: router/api-router.go:306
request:
  body:
    content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/r2s/balance-updates`

管理员手动更新供应商余额。该接口只写入本地余额和历史记录，不会自动
请求上游供应商。

## 请求体字段

- `supplier_id`: R2S 供应商 ID，必填。
- `balance_after`: 更新后的供应商余额。
- `currency_code`: 可选。余额货币；为空时使用供应商当前余额货币。
- `exchange_rate`: 可选。本次余额更新级别汇率；为空时使用供应商默认汇率。
- `balance_reminder_days`: 可选。新的余额提醒间隔天数；未传时保留供应商
  当前设置，传 `0` 表示关闭提醒。
- `note`: 可选。备注，最多 500 个字符。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 新建余额更新记录，字段同 GET `/api/r2s/balance-updates`
  的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: 参数错误、供应商不存在、货币格式不正确、汇率必须大于 `0`、
  余额提醒天数不能小于 `0`，或数据库错误。
