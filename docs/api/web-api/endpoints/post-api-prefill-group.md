---
method: POST
path: /api/prefill_group/
auth: admin
handler: controller.CreatePrefillGroup
source: router/api-router.go:318
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/prefill_group/`

创建预填组。

## 请求体字段

- `name`: 字符串，必填。组名称，不能为空且不能重复。
- `type`: 字符串，必填。组类型。
- `items`: JSON 值。组内条目，通常为字符串数组。
- `description`: 字符串，可选。描述。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 创建后的预填组对象，字段同 `GET /api/prefill_group/` 的 `data[]`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `组名称和类型不能为空`、`组名称已存在`，或创建错误。

