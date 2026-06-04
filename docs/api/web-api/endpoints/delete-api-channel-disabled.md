---
method: DELETE
path: /api/channel/disabled
auth: admin
handler: controller.DeleteDisabledChannel
source: router/api-router.go:205
request:
  query_params: []
response:
  success_http_status: 200
  envelope: raw-json
---

# DELETE `/api/channel/disabled`

删除所有禁用渠道，包括手动禁用和自动禁用渠道。

## 请求字段

无请求体，无查询参数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 删除行数。

## 失败响应

- `success`: `false`。
- `message`: 删除失败原因。

