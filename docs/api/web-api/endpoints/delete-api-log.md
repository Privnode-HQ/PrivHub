---
method: DELETE
path: /api/log/
auth: admin
handler: controller.DeleteHistoryLogs
source: router/api-router.go:289
request:
  query_params:
    - target_timestamp
response:
  success_http_status: 200
  envelope: raw-json
---

# DELETE `/api/log/`

管理员删除早于指定时间的历史日志和管理员审计日志。

## 查询参数字段

- `target_timestamp`: 目标 Unix 秒，必填。后端删除 `created_at < target_timestamp` 的记录，每批最多 100 条直到删完。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 删除总数，包含普通日志和审计日志。

## 失败响应

- `success`: `false`。
- `message`: `target timestamp is required`、普通日志删除错误或审计日志删除错误。

