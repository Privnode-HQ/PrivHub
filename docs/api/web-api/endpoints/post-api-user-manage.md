---
method: POST
path: /api/user/manage
auth: admin
handler: controller.ManageUser
source: router/api-router.go:137
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/manage`

管理员对用户执行禁用、启用、删除、升降级、登出和强制动作。

## 请求体字段

- `id`: 整数，必填。目标用户 ID。
- `action`: 字符串，必填。可能值为 `disable`、`enable`、`delete`、`promote`、`demote`、`logout`、`require_password_reset`、`require_email_bind`。
- `ban_reason`: 字符串，可选。`action=disable` 时保存为禁用原因。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.role`: 操作后的角色。
- `data.status`: 操作后的状态。
- `data.ban_reason`: 操作后的禁用原因。
- `data.force_password_reset`: 是否被要求重置密码。
- `data.force_email_bind`: 是否被要求绑定邮箱。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`用户不存在`、`无权更新同权限等级或更高权限等级的用户信息`、`无法禁用超级管理员用户`、`无法删除超级管理员用户`、`普通管理员用户无法提升支持人员为管理员`、`该用户已经是管理员`、`无法降级超级管理员用户`、`该用户已经是普通用户`，或数据库错误。

