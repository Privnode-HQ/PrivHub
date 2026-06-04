---
method: PUT
path: /api/user/
auth: admin
handler: controller.UpdateUser
source: router/api-router.go:139
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# PUT `/api/user/`

管理员更新用户资料、角色、状态、额度或密码。

## 请求体字段

- `id`: 整数，必填。目标用户 ID。
- `username`: 字符串。用户名。
- `password`: 字符串，可选。新密码；为空时不更新密码。
- `display_name`: 字符串。显示名。
- `role`: 整数。目标角色，不能大于等于当前管理员角色，Root 除外。
- `status`: 整数。用户状态，`1` 启用、`2` 禁用。
- `ban_reason`: 字符串。禁用原因。
- `email`: 字符串。邮箱。
- `quota`: 整数。剩余额度。
- `group`: 字符串。用户分组。
- `remark`: 字符串。管理备注。
- `force_password_reset`: 布尔值。是否要求重置密码。
- `force_email_bind`: 布尔值。是否要求绑定邮箱。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`输入不合法 ...`、`无权更新同权限等级或更高权限等级的用户信息`、`无权将其他用户权限等级提升到大于等于自己的权限等级`，或数据库错误。

