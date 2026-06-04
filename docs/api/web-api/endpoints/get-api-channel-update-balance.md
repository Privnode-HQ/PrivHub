---
method: GET
path: /api/channel/update_balance
auth: admin
handler: controller.UpdateAllChannelsBalance
source: router/api-router.go:201
request:
  query_params: []
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/channel/update_balance`

批量更新所有启用且非多密钥渠道的余额。余额不足时可能自动禁用渠道。

## 请求字段

无请求体，无查询参数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 查询渠道或余额更新流程中的错误信息。

