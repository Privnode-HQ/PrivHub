---
method: DELETE
path: /api/redemption/invalid
auth: admin
handler: controller.DeleteInvalidRedemption
source: router/api-router.go:255
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# DELETE `/api/redemption/invalid`

删除已使用、已禁用或已过期的兑换码。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 整数。删除行数。

## 失败响应

- `success`: `false`。
- `message`: 删除错误。

