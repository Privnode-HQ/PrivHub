---
method: PUT
path: /api/admin/service-accounts/:id
auth: admin
handler: controller.UpdateAdminServiceAccount
source: router/api-router.go:30
request:
  path_params:
    - id
  content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# PUT `/api/admin/service-accounts/:id`

更新 Admin Service Account 的名称、描述、状态或 IP 白名单。

## 路径参数字段

- `id`: 整数，必填。Service Account 数据库 ID。

## 请求体字段

- `name`: 字符串，可选。非空时更新名称，最多 80 个字符。
- `description`: 字符串或 `null`，可选。传入时更新描述。
- `status`: 整数，可选。`1` 启用，`2` 禁用；`0` 表示不变。
- `allow_ips`: 字符串或 `null`，可选。传入时替换 IP 白名单。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 更新后的 Service Account 视图，字段同 `GET /api/admin/service-accounts/` 的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `Service Account ID 无效`、`无效的参数`、`Service Account 状态无效`、目标管理员不可管理、IP 白名单错误或数据库错误。

