---
method: PUT
path: /api/message/template
auth: admin
handler: controller.UpdateMessageTemplate
source: router/api-router.go:175
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# PUT `/api/message/template`

管理员更新站内消息邮件模板。

## 请求体字段

- `template`: 字符串，必填。HTML 模板，必须通过模板校验。

## 成功响应字段

- `success`: `true`。
- `message`: `模板已更新`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`，模板校验错误，或选项保存错误。

