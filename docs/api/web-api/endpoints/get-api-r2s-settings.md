---
method: GET
path: /api/r2s/settings
auth: admin
handler: controller.GetR2SSettings
source: router/api-router.go:290
request:
  query_params: []
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/r2s/settings`

管理员读取 R2S 系统级设置。该接口只读取本地配置，不会从上游供应商
自动获取数据。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.receipt_required`: 是否要求新增供应商付款记录时填写
  `receipt_url`，用于收据或支付截图地址。
- `data.default_currency_code`: R2S 默认货币。未显式设置时使用当前
  系统额度展示货币。
- `data.balance_reminder_days`: 默认供应商余额提醒间隔天数，`0` 表示
  不提醒。
- `data.system_currency_code`: 当前系统额度展示货币。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。
