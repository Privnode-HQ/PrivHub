---
method: PUT
path: /api/vendors/
auth: admin
handler: controller.UpdateVendorMeta
source: router/api-router.go:344
request:
  body: model.Vendor
response:
  success_http_status: 200
  envelope: common
---

# PUT `/api/vendors/`

管理员更新模型供应商。

## 请求体字段

- `id`: 供应商 ID，必填。
- `name`: 新名称；后端会检查与其他供应商是否重复。
- `description`: 新说明。
- `icon`: 新图标名。
- `status`: 新状态，通常 `1` 启用。
- `created_time`: 请求体可传但模型保存会以数据库行为为准。
- `updated_time`: 后端更新为当前时间。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 供应商 ID。
- `data.name`: 当前名称。
- `data.description`: 当前说明。
- `data.icon`: 当前图标。
- `data.status`: 当前状态。
- `data.created_time`: 创建时间。
- `data.updated_time`: 更新时间。

## 失败响应

- `success`: `false`。
- `message`: JSON 绑定错误、`缺少供应商 ID`、`供应商名称已存在` 或数据库错误。

