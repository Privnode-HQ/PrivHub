---
method: POST
path: /api/token/batch
auth: user
handler: controller.DeleteTokenBatch
source: router/api-router.go:228
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/token/batch`

批量删除当前用户 Token。

## 请求体字段

- `ids`: 整数数组，必填。Token ID 列表，不能为空。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 整数。实际删除数量。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`，或批量删除错误。

