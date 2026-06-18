---
method: PUT
path: /api/r2s/suppliers/:id
auth: admin
handler: controller.UpdateR2SSupplier
source: router/api-router.go:297
request:
  path_params:
    - id
  body:
    content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# PUT `/api/r2s/suppliers/:id`

管理员更新 R2S 供应商基础信息、默认汇率和余额提醒配置。供应商余额本身
应通过 POST `/api/r2s/balance-updates` 或 POST `/api/r2s/payments`
写入历史记录后更新。

## 路径参数字段

- `id`: R2S 供应商 ID，必须是整数。

## 请求体字段

- `name`: 供应商名称，必填，最多 100 个字符。
- `description`: 供应商说明，最多 255 个字符。
- `status`: `active` 或 `disabled`。
- `default_currency_code`: 默认货币。
- `default_exchange_rate`: 默认汇率，必须大于 `0`。
- `balance_reminder_days`: 余额提醒间隔天数，不能小于 `0`；`0` 表示停用
  该供应商的余额提醒。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 更新后的 R2S 供应商对象，字段同 GET `/api/r2s/suppliers`
  的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: ID 解析错误、参数错误、供应商不存在、字段校验失败或数据库
  错误。
