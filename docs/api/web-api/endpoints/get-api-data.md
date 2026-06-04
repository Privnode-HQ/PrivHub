---
method: GET
path: /api/data/
auth: support
handler: controller.GetAllQuotaDates
source: router/api-router.go:297
request:
  query_params:
    - start_timestamp
    - end_timestamp
    - username
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/data/`

支持人员或管理员查询全站用量数据看板。

## 查询参数字段

- `start_timestamp`: 整数，可选。起始时间 Unix 秒。
- `end_timestamp`: 整数，可选。结束时间 Unix 秒。
- `username`: 字符串，可选。按用户名过滤。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 用量数据数组。
- `data[].id`: 数据记录 ID。
- `data[].user_id`: 用户 ID。
- `data[].username`: 用户名。
- `data[].model_name`: 模型名称。
- `data[].created_at`: 聚合时间 Unix 秒，按小时聚合。
- `data[].token_used`: token 用量。
- `data[].count`: 请求次数。
- `data[].quota`: 额度消耗。

## 失败响应

- `success`: `false`。
- `message`: 查询错误。

