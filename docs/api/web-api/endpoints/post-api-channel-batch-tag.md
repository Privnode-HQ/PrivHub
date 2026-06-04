---
method: POST
path: /api/channel/batch/tag
auth: admin
handler: controller.BatchSetChannelTag
source: router/api-router.go:214
request:
  body: ChannelBatch
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/channel/batch/tag`

批量设置渠道标签。

## 请求体字段

- `ids`: 渠道 ID 数组，必填且不能为空。
- `tag`: 新标签。可为字符串或 `null`；传 `null` 表示清空标签。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 已处理的渠道数量。

## 失败响应

- `success`: `false`。
- `message`: `参数错误` 或批量设置失败原因。

