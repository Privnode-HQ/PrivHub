---
method: GET
path: /api/log/
auth: support
handler: controller.GetAllLogs
source: router/api-router.go:288
request:
  query_params:
    - p
    - page_size
    - type
    - start_timestamp
    - end_timestamp
    - username
    - token_name
    - model_name
    - channel
    - group
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/log/`

支持人员或管理员分页查看系统日志。`type=3` 时读取管理员审计日志并转换为普通日志结构。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。
- `type`: 日志类型。`0` 全部，`1` 充值，`2` 消耗，`3` 管理，`4` 系统，`5` 错误，`6` 退款。
- `start_timestamp`: 起始 Unix 秒，`0` 表示不限。
- `end_timestamp`: 结束 Unix 秒，`0` 表示不限。
- `username`: 用户名精确过滤。
- `token_name`: Token 名称精确过滤。
- `model_name`: 模型名过滤，后端使用 SQL `LIKE`。
- `channel`: 渠道 ID，`0` 表示不限。
- `group`: 分组精确过滤。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 总数。
- `data.items`: 日志数组。
- `data.items[].id`: 日志 ID。
- `data.items[].user_id`: 用户 ID。
- `data.items[].created_at`: 创建时间 Unix 秒。
- `data.items[].type`: 日志类型。
- `data.items[].content`: 日志内容。
- `data.items[].username`: 用户名。
- `data.items[].token_name`: Token 名称。
- `data.items[].model_name`: 模型名称。
- `data.items[].quota`: 本次额度变化或消耗。
- `data.items[].prompt_tokens`: 输入 token 数。
- `data.items[].completion_tokens`: 输出 token 数。
- `data.items[].use_time`: 耗时秒。
- `data.items[].is_stream`: 是否流式请求。
- `data.items[].channel`: 渠道 ID。
- `data.items[].channel_name`: 渠道名称，列表查询会按渠道 ID 补充。
- `data.items[].token_id`: Token ID。
- `data.items[].group`: 分组。
- `data.items[].ip`: 记录的 IP，取决于用户设置。
- `data.items[].other`: 其他 JSON 字符串。

## 失败响应

- `success`: `false`。
- `message`: 查询日志或审计日志失败原因。

