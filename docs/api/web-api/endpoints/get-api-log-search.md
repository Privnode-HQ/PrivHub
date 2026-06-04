---
method: GET
path: /api/log/search
auth: support
handler: controller.SearchAllLogs
source: router/api-router.go:292
request:
  query_params:
    - keyword
    - type
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/log/search`

支持人员或管理员按关键字搜索日志。`type=3` 时搜索管理员审计日志。

## 查询参数字段

- `keyword`: 搜索关键字。普通日志中用于匹配 `type` 或 `content` 前缀；审计日志中交给审计日志搜索函数。
- `type`: 日志类型。`3` 表示管理/审计日志，其他值搜索普通日志。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 日志数组。
- `data[].id`: 日志 ID。
- `data[].user_id`: 用户 ID。
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
- `data[].channel_name`: 渠道名称。
- `data[].token_id`: Token ID。
- `data[].group`: 分组。
- `data[].ip`: IP。
- `data[].other`: 其他 JSON 字符串。

## 失败响应

- `success`: `false`。
- `message`: 搜索失败原因。

