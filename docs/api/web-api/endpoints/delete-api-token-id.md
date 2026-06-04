---
method: DELETE
path: /api/token/:id
auth: user
handler: controller.DeleteToken
source: router/api-router.go:227
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: custom_success
---

# DELETE `/api/token/:id`

删除当前用户的一个 Token。

## 路径参数字段

- `id`: 整数，必填。Token ID。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 删除错误。

