---
method: DELETE
path: /api/user/:id
auth: admin
handler: controller.DeleteUser
source: router/api-router.go:140
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: custom_success
---

# DELETE `/api/user/:id`

管理员硬删除指定用户。

## 路径参数字段

- `id`: 整数，必填。目标用户 ID。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为路径 ID 解析错误、用户不存在、`无权删除同权限等级或更高权限等级的用户`，或删除错误。

