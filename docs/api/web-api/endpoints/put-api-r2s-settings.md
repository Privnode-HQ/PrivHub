---
method: PUT
path: /api/r2s/settings
auth: admin
handler: controller.UpdateR2SSettings
source: router/api-router.go:291
request:
  body:
    content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# PUT `/api/r2s/settings`

管理员更新 R2S 系统级设置。该接口只更新本地配置，不会从上游供应商
自动获取数据。

## 请求体字段

- `receipt_required`: 布尔值。为 `true` 时，新增付款记录必须填写
  `receipt_url`。
- `default_currency_code`: 默认货币代码，支持 `USD`、`CNY`、`USDT` 等
  2-16 位大写字母、数字、下划线或短横线。为空时使用当前系统货币。
- `balance_reminder_days`: 默认余额提醒间隔天数，不能小于 `0`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.receipt_required`: 更新后的收据必填开关。
- `data.default_currency_code`: 更新后的默认货币。
- `data.balance_reminder_days`: 更新后的默认提醒间隔。
- `data.system_currency_code`: 当前系统额度展示货币。

## 失败响应

- `success`: `false`。
- `message`: 参数错误、默认货币格式不正确、余额提醒天数不能小于 `0`，
  或数据库错误。
