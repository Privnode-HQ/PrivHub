---
method: GET
path: /api/data/self
auth: user
handler: controller.GetUserQuotaDates
source: router/api-router.go:298
request:
  query_params:
    - start_timestamp
    - end_timestamp
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/data/self`

查询当前用户用量数据看板。

## 查询参数字段

- `start_timestamp`: 整数，可选。起始时间 Unix 秒。
- `end_timestamp`: 整数，可选。结束时间 Unix 秒；与起始时间跨度不能超过 1 个月。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 用量数据数组，字段同 `GET /api/data/` 的 `data[]`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `时间跨度不能超过 1 个月`，或查询错误。

