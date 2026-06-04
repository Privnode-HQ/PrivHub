---
method: POST
path: /api/vendors/
auth: admin
handler: controller.CreateVendorMeta
source: router/api-router.go:343
request:
  body: model.Vendor
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/vendors/`

管理员创建模型供应商。

## 请求体字段

- `name`: 供应商名称，必填，软删除范围内唯一。
- `description`: 供应商说明，可为空。
- `icon`: 图标名，可为空。
- `status`: 状态，通常 `1` 表示启用；未传时模型默认值为 `1`。
- `id`: 创建时忽略。
- `created_time`: 创建时由后端设置。
- `updated_time`: 创建时由后端设置。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 新供应商 ID。
- `data.name`: 名称。
- `data.description`: 说明。
- `data.icon`: 图标名。
- `data.status`: 状态。
- `data.created_time`: 创建时间。
- `data.updated_time`: 更新时间。

## 失败响应

- `success`: `false`。
- `message`: JSON 绑定错误、`供应商名称不能为空`、`供应商名称已存在` 或数据库错误。

