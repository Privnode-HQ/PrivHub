---
method: DELETE
path: /api/admin/service-accounts/:id
auth: admin
handler: controller.DeleteAdminServiceAccount
source: router/api-router.go:32
request:
  path_params:
    - id
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# DELETE `/api/admin/service-accounts/:id`

删除 Admin Service Account 并撤销其凭据。

## 路径参数字段

- `id`: 整数，必填。Service Account 数据库 ID。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `Service Account ID 无效`、目标管理员不可管理、记录不存在或数据库错误。

