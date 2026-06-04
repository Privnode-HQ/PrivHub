---
method: GET
path: /api/redemption/:id
auth: support
handler: controller.GetRedemption
source: router/api-router.go:248
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/redemption/:id`

获取单个兑换码。

## 路径参数字段

- `id`: 整数，必填。兑换码 ID。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 兑换码对象，字段同 `GET /api/redemption/` 的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: ID 解析、兑换码不存在或查询错误。

