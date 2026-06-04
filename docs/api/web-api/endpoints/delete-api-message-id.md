---
method: DELETE
path: /api/message/:id
auth: admin
handler: controller.DeleteMessage
source: router/api-router.go:181
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: custom_success
---

# DELETE `/api/message/:id`

删除消息草稿。

## 路径参数字段

- `id`: 整数，必填。消息 ID。

## 成功响应字段

- `success`: `true`。
- `message`: `草稿已删除`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的消息ID`，或删除错误。

