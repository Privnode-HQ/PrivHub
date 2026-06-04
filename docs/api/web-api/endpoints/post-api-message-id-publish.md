---
method: POST
path: /api/message/:id/publish
auth: admin
handler: controller.PublishMessage
source: router/api-router.go:178
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/message/:id/publish`

管理员发布消息草稿并触发投递。

## 路径参数字段

- `id`: 整数，必填。消息 ID。

## 成功响应字段

- `success`: `true`。
- `message`: `消息已上线`。
- `data`: 发布后的消息对象。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的消息ID`，或发布/投递错误。

