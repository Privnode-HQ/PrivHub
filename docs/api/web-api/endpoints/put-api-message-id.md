---
method: PUT
path: /api/message/:id
auth: admin
handler: controller.UpdateMessage
source: router/api-router.go:177
request:
  path_params:
    - id
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# PUT `/api/message/:id`

管理员更新消息草稿。

## 路径参数字段

- `id`: 整数，必填。消息 ID。

## 请求体字段

- `title`: 字符串，必填。
- `content`: 字符串，必填。
- `target_type`: 字符串。
- `target_groups`: 字符串数组。
- `target_user_ids`: 整数数组。

## 成功响应字段

- `success`: `true`。
- `message`: `草稿已更新`。
- `data`: 更新后的消息对象。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的消息ID`、`无效的参数`、`标题和内容不能为空`，或服务层错误。

