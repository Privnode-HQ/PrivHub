---
method: POST
path: /api/r2s/suppliers
auth: admin
handler: controller.CreateR2SSupplier
source: router/api-router.go:295
request:
  body:
    content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/r2s/suppliers`

管理员创建 R2S 上游供应商。余额和汇率均为本地手动维护，不会自动同步
上游。

## 请求体字段

- `name`: 供应商名称，必填，最多 100 个字符。
- `description`: 供应商说明，最多 255 个字符。
- `status`: 可选。`active` 或 `disabled`，为空时为 `active`。
- `default_currency_code`: 可选。默认货币；为空时使用 R2S 默认货币。
- `default_exchange_rate`: 可选。默认汇率，必须大于 `0`，为空时为 `1`。
- `balance_amount`: 可选。初始供应商余额。
- `balance_currency_code`: 可选。余额货币；为空时使用默认货币。
- `balance_reminder_days`: 可选。余额提醒间隔天数，`0` 使用系统默认。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 新建的 R2S 供应商对象，字段同 GET `/api/r2s/suppliers`
  的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: 参数错误、供应商名称不能为空、货币格式不正确、汇率必须大于
  `0`、余额提醒天数不能小于 `0`，或数据库错误。
