---
method: GET
path: /api/message/self/
auth: user
handler: controller.GetMyMessages
source: router/api-router.go:159
request:
  query_params:
    - p
    - page_size
response:
  success_http_status: 200
  envelope: message_list
---

# GET `/api/message/self/`

获取当前用户的站内消息列表。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，最大 `100`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.items`: 消息数组。
- `data.items[].id`: 用户消息 ID。
- `data.items[].title`: 标题。
- `data.items[].content`: Markdown 或文本内容。
- `data.items[].status`: 消息状态。
- `data.items[].source`: 消息来源。
- `data.items[].published_at`: 发布时间。
- `data.items[].created_at`: 创建时间。
- `data.items[].read_at`: 已读时间，未读时为空。
- `data.items[].email_sent_at`: 邮件发送时间，未发送时为空。
- `data.items[].html_content`: 用当前模板渲染后的 HTML；渲染失败时为空字符串。
- `data.total`: 总消息数。

## 失败响应

- `success`: `false`。
- `message`: 查询错误。

