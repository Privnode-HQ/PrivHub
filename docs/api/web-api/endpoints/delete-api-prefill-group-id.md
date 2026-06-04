---
method: DELETE
path: /api/prefill_group/:id
auth: admin
handler: controller.DeletePrefillGroup
source: router/api-router.go:320
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# DELETE `/api/prefill_group/:id`

删除预填组。

## 路径参数字段

- `id`: 整数，必填。预填组 ID。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: `null`。

## 失败响应

- `success`: `false`。
- `message`: ID 解析或删除错误。

