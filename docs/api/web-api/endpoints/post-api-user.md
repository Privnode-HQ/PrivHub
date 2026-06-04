---
method: POST
path: /api/user/
auth: admin
handler: controller.CreateUser
source: router/api-router.go:134
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/`

管理员创建用户。

## 请求体字段

- `username`: 字符串，必填。用户名，前后空白会被去除，最长 20。
- `password`: 字符串，必填。密码，长度 8 到 20。
- `display_name`: 字符串，可选。显示名；为空时使用用户名。
- `role`: 整数，可选。目标角色，不能为访客，不能大于等于当前管理员角色；`0` 会被转为普通用户。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`输入不合法 ...`、`无效的用户角色`、`无法创建权限大于等于自己的用户`，或创建错误。

