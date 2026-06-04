---
method: POST
path: /api/message/self/:id/read
auth: user
handler: controller.ReadMyMessage
source: router/api-router.go:161
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/message/self/:id/read`

把当前用户的一条消息标记为已读。

## 路径参数字段

- `id`: 整数，必填。用户消息 ID。

## 成功响应字段

- `success`: `true`。
- `message`: `消息已标记为已读`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的消息ID`，或标记已读错误。

