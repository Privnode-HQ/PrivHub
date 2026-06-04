---
method: DELETE
path: /api/channel/:id
auth: admin
handler: controller.DeleteChannel
source: router/api-router.go:209
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: raw-json
---

# DELETE `/api/channel/:id`

删除单个渠道。

## 路径参数字段

- `id`: 渠道 ID，必须是整数；控制器将非整数视为 `0`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 删除失败原因。

