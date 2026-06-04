---
method: POST
path: /api/message/:id/retry
auth: admin
handler: controller.RetryMessageDelivery
source: router/api-router.go:180
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/message/:id/retry`

重试指定消息的失败邮件投递。

## 路径参数字段

- `id`: 整数，必填。消息 ID。

## 成功响应字段

- `success`: `true`。
- `message`: `重试任务已提交`。
- `data`: 整数。提交重试的数量。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的消息ID`，或服务层错误。

