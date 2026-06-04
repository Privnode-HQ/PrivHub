---
method: GET
path: /api/message/self/unread
auth: user
handler: controller.GetMyUnreadMessageCount
source: router/api-router.go:160
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
  data: integer
---

# GET `/api/message/self/unread`

获取当前用户未读消息数量。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 整数。未读消息数量。

## 失败响应

- `success`: `false`。
- `message`: 统计错误。

