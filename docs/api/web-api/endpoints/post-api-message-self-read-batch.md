---
method: POST
path: /api/message/self/read/batch
auth: user
handler: controller.BatchReadMyMessages
source: router/api-router.go:162
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/message/self/read/batch`

批量把当前用户的消息标记为已读。

## 请求体字段

- `ids`: 整数数组，必填。用户消息 ID 列表，不能为空。

## 成功响应字段

- `success`: `true`。
- `message`: `批量已读成功`。
- `data`: 整数。实际标记已读的记录数。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`请选择至少一条消息`，或批量更新错误。

