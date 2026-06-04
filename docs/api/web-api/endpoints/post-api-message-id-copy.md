---
method: POST
path: /api/message/:id/copy
auth: admin
handler: controller.CopyMessage
source: router/api-router.go:179
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/message/:id/copy`

复制一条消息为新草稿。

## 路径参数字段

- `id`: 整数，必填。源消息 ID。

## 成功响应字段

- `success`: `true`。
- `message`: `草稿已复制`。
- `data`: 新草稿消息对象。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的消息ID`，或服务层错误。

