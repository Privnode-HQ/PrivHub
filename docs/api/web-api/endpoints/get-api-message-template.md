---
method: GET
path: /api/message/template
auth: support
handler: controller.GetMessageTemplate
source: router/api-router.go:169
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/message/template`

获取站内消息邮件模板和可用占位符。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.template`: 当前 HTML 模板。
- `data.placeholders`: 可用占位符数组。

## 失败响应

当前控制器没有显式错误分支。

