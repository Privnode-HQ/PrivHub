---
method: GET
path: /api/log/self/search
auth: user
handler: controller.SearchUserLogs
source: router/api-router.go:294
request:
  query_params:
    - keyword
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/log/self/search`

当前用户按关键字搜索自己的日志。

## 查询参数字段

- `keyword`: 搜索关键字。模型层按当前用户 ID 和日志类型字段匹配。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 当前用户日志数组。
- `data[].id`: 脱敏后的日志 ID。
- `data[].user_id`: 当前用户 ID。
- `data[].created_at`: 创建时间。
- `data[].type`: 日志类型。
- `data[].content`: 内容。
- `data[].username`: 用户名。
- `data[].token_name`: Token 名称。
- `data[].model_name`: 模型名。
- `data[].quota`: 额度。
- `data[].prompt_tokens`: 输入 token 数。
- `data[].completion_tokens`: 输出 token 数。
- `data[].use_time`: 耗时秒。
- `data[].is_stream`: 是否流式。
- `data[].channel`: 渠道 ID。
- `data[].channel_name`: 空字符串。
- `data[].token_id`: Token ID。
- `data[].group`: 分组。
- `data[].ip`: IP。
- `data[].other`: 已移除管理员信息的 JSON 字符串。

## 失败响应

- `success`: `false`。
- `message`: 搜索失败原因。

