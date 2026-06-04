---
method: GET
path: /api/token/:id
auth: user
handler: controller.GetToken
source: router/api-router.go:224
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/token/:id`

获取当前用户的单个 Token。

## 路径参数字段

- `id`: 整数，必填。Token ID。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: Token 对象，字段同 `GET /api/token/` 的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: ID 解析、Token 不存在或不属于当前用户。

