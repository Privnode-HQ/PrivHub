---
method: GET
path: /api/channel/update_balance/:id
auth: admin
handler: controller.UpdateChannelBalance
source: router/api-router.go:202
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/channel/update_balance/:id`

更新单个渠道余额。

## 路径参数字段

- `id`: 渠道 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `balance`: 最新余额，单位 USD。不同渠道类型会按上游返回计算。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `多密钥渠道不支持余额查询`、`尚未实现`、上游请求错误、响应解析错误、渠道不存在或 `id` 非整数。

