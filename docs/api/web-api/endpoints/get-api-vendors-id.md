---
method: GET
path: /api/vendors/:id
auth: support
handler: controller.GetVendorMeta
source: router/api-router.go:338
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/vendors/:id`

按 ID 查看供应商详情。

## 路径参数字段

- `id`: 供应商 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 供应商 ID。
- `data.name`: 名称。
- `data.description`: 说明。
- `data.icon`: 图标名。
- `data.status`: 状态，通常 `1` 启用。
- `data.created_time`: 创建时间 Unix 秒。
- `data.updated_time`: 更新时间 Unix 秒。

## 失败响应

- `success`: `false`。
- `message`: `id` 非整数、供应商不存在或数据库错误。

