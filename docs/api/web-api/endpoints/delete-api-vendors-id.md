---
method: DELETE
path: /api/vendors/:id
auth: admin
handler: controller.DeleteVendorMeta
source: router/api-router.go:345
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# DELETE `/api/vendors/:id`

管理员删除模型供应商。模型使用 Gorm 软删除字段。

## 路径参数字段

- `id`: 供应商 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: `null`。

## 失败响应

- `success`: `false`。
- `message`: `id` 非整数或删除失败原因。

