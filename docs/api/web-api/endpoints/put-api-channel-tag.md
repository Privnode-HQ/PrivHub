---
method: PUT
path: /api/channel/tag
auth: admin
handler: controller.EditTagChannels
source: router/api-router.go:208
request:
  body: ChannelTag
response:
  success_http_status: 200
  envelope: raw-json
---

# PUT `/api/channel/tag`

按标签批量编辑渠道字段。

## 请求体字段

- `tag`: 目标标签，必填。
- `new_tag`: 新标签；为 `null` 时不修改标签。
- `priority`: 新优先级；为 `null` 时不修改。
- `weight`: 新权重；为 `null` 时不修改。
- `model_mapping`: 新模型映射字符串；为 `null` 时不修改。
- `models`: 新模型列表字符串；为 `null` 时不修改。
- `groups`: 新分组字符串；为 `null` 时不修改。
- `param_override`: 新参数覆盖 JSON 字符串；非空时必须是合法 JSON。
- `header_override`: 新请求头覆盖 JSON 字符串；非空时必须是合法 JSON。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、`tag不能为空`、`参数覆盖必须是合法的 JSON 格式`、`请求头覆盖必须是合法的 JSON 格式` 或批量编辑错误。

