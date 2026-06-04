---
method: DELETE
path: /api/models/:id
auth: admin
handler: controller.DeleteModelMeta
source: router/api-router.go:363
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# DELETE `/api/models/:id`

管理员删除模型元数据。模型使用 Gorm 软删除字段。

## 路径参数字段

- `id`: 模型元数据 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: `null`。

## 失败响应

- `success`: `false`。
- `message`: `id` 非整数或删除失败原因。

