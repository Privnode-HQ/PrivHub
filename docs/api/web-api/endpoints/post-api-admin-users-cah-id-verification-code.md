---
method: POST
path: /api/admin/users/:cah_id/verification_code
auth: admin
handler: controller.SendAdminUserVerificationCode
source: router/api-router.go:22
request:
  path_params:
    - cah_id
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/admin/users/:cah_id/verification_code`

管理员按 CAH ID 向用户绑定邮箱发送人工核验验证码，并在响应中返回同一枚验证码。该验证码不是登录、注册、密码重置、邮箱绑定或两步验证用途。

## 路径参数字段

- `cah_id`: 字符串，必填。目标用户 CAH ID。后端会通过 `model.GetUserByCAHID` 查找用户。

## 请求体字段

- `purpose`: 字符串，必填。验证码用途，最多 120 个字符；不能包含换行、制表符或控制字符；不能包含登录、注册、密码、邮箱绑定、两步验证等身份认证用途关键词。

## 成功响应字段

- `success`: `true`。
- `message`: `验证码已发送`。
- `data.cah_id`: 目标用户标准 CAH ID。
- `data.email`: 实际发送邮箱。
- `data.purpose`: 规范化后的用途。
- `data.code`: 8 位十六进制验证码。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`验证码用途不能为空`、`验证码用途不能超过 120 个字符`、`验证码用途不能包含换行、制表符或控制字符`、`验证码用途不能是登录、注册、密码重置、邮箱绑定或两步验证等身份认证用途`、`无权向同级或更高权限用户发送验证码`、`目标用户未绑定邮箱`、`目标用户邮箱格式无效`，或邮件发送错误。

