---
method: PUT
path: /api/prefill_group/
auth: admin
handler: controller.UpdatePrefillGroup
source: router/api-router.go:319
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# PUT `/api/prefill_group/`

更新预填组。

## 请求体字段

- `id`: 整数，必填。预填组 ID。
- `name`: 字符串。组名称，不能与其他组重复。
- `type`: 字符串。组类型。
- `items`: JSON 值。组内条目。
- `description`: 字符串。描述。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 更新后的预填组对象。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `缺少组 ID`、`组名称已存在`，或更新错误。

