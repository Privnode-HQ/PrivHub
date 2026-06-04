---
method: PUT
path: /api/user/self
auth: user
handler: controller.UpdateSelf
source: router/api-router.go:83
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# PUT `/api/user/self`

更新当前用户资料或个人侧边栏模块配置。

## 请求体字段

侧边栏更新模式：

- `sidebar_modules`: 字符串。传入时只更新用户设置中的侧边栏模块配置。

资料更新模式：

- `username`: 字符串，可选。新用户名，必须满足用户模型校验。
- `display_name`: 字符串，可选。显示名。
- `password`: 字符串，可选。新密码；传入时必须同时提供正确 `original_password`。
- `original_password`: 字符串，可选。原密码，仅用于修改密码校验。

## 成功响应字段

- `success`: `true`。
- `message`: 侧边栏模式为 `设置更新成功`；资料模式通常为空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`输入不合法 ...`、`原密码错误`、`更新设置失败: ...`、session 保存失败或数据库错误。

