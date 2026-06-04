---
method: POST
path: /api/user/:id/access_link
auth: admin
handler: controller.GenerateUserAccessLink
source: router/api-router.go:136
request:
  path_params:
    - id
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/:id/access_link`

管理员为指定用户生成一次性账户访问链接。

## 路径参数字段

- `id`: 整数，必填。目标用户 ID。

## 成功响应字段

- `success`: `true`。
- `message`: `访问链接已生成，24 小时内有效`。
- `data.grant_id`: 授权记录 ID。
- `data.access_link`: 可访问的完整链接或路径。
- `data.expires_at`: 链接过期时间。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的用户 ID`、`目标用户当前不可访问`，或服务层错误。

