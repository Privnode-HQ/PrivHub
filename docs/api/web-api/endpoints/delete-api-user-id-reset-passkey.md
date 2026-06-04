---
method: DELETE
path: /api/user/:id/reset_passkey
auth: admin
handler: controller.AdminResetPasskey
source: router/api-router.go:141
request:
  path_params:
    - id
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# DELETE `/api/user/:id/reset_passkey`

管理员重置指定用户 Passkey，即删除其已绑定凭证。

## 路径参数字段

- `id`: 整数，必填。目标用户 ID。

## 成功响应字段

- `success`: `true`。
- `message`: `Passkey 已重置`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的用户 ID`、`该用户尚未绑定 Passkey`，或用户/凭证查询错误。

