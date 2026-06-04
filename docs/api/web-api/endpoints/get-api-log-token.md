---
method: GET
path: /api/log/token
auth: none + CORS
handler: controller.GetLogByKey
source: router/api-router.go:302
request:
  query_params:
    - key
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/log/token`

按 API key 查询该 key 对应 Token 的日志。该路由只挂载 CORS 中间件，没有用户或管理员认证中间件。

## 查询参数字段

- `key`: API key。后端会移除 `sk-` 前缀后查询 Token。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 日志数组。
- `data[].id`: 脱敏后的日志 ID。
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
- `data[].channel_name`: 空字符串。
- `data[].token_id`: Token ID。
- `data[].group`: 分组。
- `data[].ip`: IP。
- `data[].other`: 已移除管理员信息的 JSON 字符串。

## 失败响应

- `success`: `false`。
- `message`: key 未匹配 Token、日志查询失败或数据库错误。

