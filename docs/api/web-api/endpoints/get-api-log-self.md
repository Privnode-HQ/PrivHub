---
method: GET
path: /api/log/self
auth: user
handler: controller.GetUserLogs
source: router/api-router.go:293
request:
  query_params:
    - p
    - page_size
    - type
    - start_timestamp
    - end_timestamp
    - token_name
    - model_name
    - group
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/log/self`

当前用户分页查看自己的日志。响应会隐藏渠道名称、移除 `other.admin_info`，并将 `id` 缩短为 `id % 1024`。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。
- `type`: 日志类型。
- `start_timestamp`: 起始 Unix 秒。
- `end_timestamp`: 结束 Unix 秒。
- `token_name`: Token 名称过滤。
- `model_name`: 模型名过滤，使用 SQL `LIKE`。
- `group`: 分组过滤。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 总数。
- `data.items`: 当前用户日志数组。
- `data.items[].id`: 脱敏后的日志 ID。
- `data.items[].user_id`: 当前用户 ID。
- `data.items[].created_at`: 创建时间。
- `data.items[].type`: 日志类型。
- `data.items[].content`: 内容。
- `data.items[].username`: 用户名。
- `data.items[].token_name`: Token 名称。
- `data.items[].model_name`: 模型名。
- `data.items[].quota`: 额度。
- `data.items[].prompt_tokens`: 输入 token 数。
- `data.items[].completion_tokens`: 输出 token 数。
- `data.items[].use_time`: 耗时秒。
- `data.items[].is_stream`: 是否流式。
- `data.items[].channel`: 渠道 ID。
- `data.items[].channel_name`: 空字符串。
- `data.items[].token_id`: Token ID。
- `data.items[].group`: 分组。
- `data.items[].ip`: IP，取决于用户设置。
- `data.items[].other`: 已移除管理员信息的 JSON 字符串。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。

