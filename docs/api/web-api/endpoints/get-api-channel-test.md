---
method: GET
path: /api/channel/test
auth: admin
handler: controller.TestAllChannels
source: router/api-router.go:199
request:
  query_params: []
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/channel/test`

异步测试所有渠道，并根据测试结果更新响应时间和可能的自动启停状态。

## 请求字段

无请求体，无查询参数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `测试已在运行中`、查询渠道失败或测试任务启动失败。

