---
method: DELETE
path: /api/user/:id/2fa
auth: admin
handler: controller.AdminDisable2FA
source: router/api-router.go:143
request:
  path_params:
    - id
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# DELETE `/api/user/:id/2fa`

管理员强制禁用指定用户的 2FA。

## 路径参数字段

- `id`: 整数，必填。目标用户 ID。

## 成功响应字段

- `success`: `true`。
- `message`: `用户2FA已被强制禁用`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `用户ID格式错误`、`无权操作同级或更高级用户的2FA设置`、`用户未启用2FA`，或用户/2FA 查询错误。

