---
method: GET
path: /api/message/
auth: support
handler: controller.GetAdminMessages
source: router/api-router.go:168
request:
  query_params:
    - p
    - page_size
response:
  success_http_status: 200
  envelope: message_admin_list
---

# GET `/api/message/`

支持人员或管理员查看站内消息管理列表。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，最大 `100`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.items`: 消息数组。
- `data.items[].id`: 消息 ID。
- `data.items[].title`: 标题。
- `data.items[].content`: 内容。
- `data.items[].status`: 状态。
- `data.items[].source`: 来源。
- `data.items[].target_type`: 投递目标类型。
- `data.items[].target_groups`: 目标分组数组。
- `data.items[].target_user_ids`: 目标用户 ID 数组。
- `data.items[].target_user_options`: 目标用户下拉选项数组。
- `data.items[].published_at`: 发布时间。
- `data.items[].created_at`: 创建时间。
- `data.items[].updated_at`: 更新时间。
- `data.items[].html_content`: 用当前模板渲染后的 HTML。
- `data.items[].delivery_total`: 投递总数。
- `data.items[].read_total`: 已读总数。
- `data.items[].email_sent`: 邮件发送成功数。
- `data.items[].email_failed`: 邮件发送失败数。
- `data.total`: 总消息数。

## 失败响应

- `success`: `false`。
- `message`: 消息或投递统计查询错误。

