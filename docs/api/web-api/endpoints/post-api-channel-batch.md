---
method: POST
path: /api/channel/batch
auth: admin
handler: controller.DeleteChannelBatch
source: router/api-router.go:210
request:
  body: ChannelBatch
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/channel/batch`

批量删除渠道。

## 请求体字段

- `ids`: 渠道 ID 数组，必填且不能为空。
- `tag`: 忽略；同一结构也被批量改标签接口复用。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 已请求删除的 ID 数量。

## 失败响应

- `success`: `false`。
- `message`: `参数错误` 或批量删除失败原因。

