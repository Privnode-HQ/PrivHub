---
method: DELETE
path: /api/redemption/:id
auth: admin
handler: controller.DeleteRedemption
source: router/api-router.go:256
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: custom_success
---

# DELETE `/api/redemption/:id`

删除一个兑换码。

## 路径参数字段

- `id`: 整数，必填。兑换码 ID。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 删除错误。

