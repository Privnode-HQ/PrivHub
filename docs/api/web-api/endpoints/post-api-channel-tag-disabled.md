---
method: POST
path: /api/channel/tag/disabled
auth: admin
handler: controller.DisableTagChannels
source: router/api-router.go:206
request:
  body: ChannelTag
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/channel/tag/disabled`

按标签批量禁用渠道。

## 请求体字段

- `tag`: 目标标签，必填。
- `new_tag`: 忽略。
- `priority`: 忽略。
- `weight`: 忽略。
- `model_mapping`: 忽略。
- `models`: 忽略。
- `groups`: 忽略。
- `param_override`: 忽略。
- `header_override`: 忽略。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: `参数错误` 或批量禁用失败原因。

