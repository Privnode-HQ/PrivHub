---
method: POST
path: /api/usage/reset
auth: root
handler: controller.ResetUsageLimits
source: router/api-router.go:235
request:
  content_type: application/json
response:
  success_http_status: 200
  bad_request_http_status: 400
  envelope: common
---

# POST `/api/usage/reset`

Root 重置用户用量限制记录，可按范围、分组或用户 ID 执行。

## 请求体字段

- `scope`: 字符串，必填。重置范围，后端会去空白并转小写；具体允许值由 `service.ResetUserUsageLimits` 校验。
- `group_names`: 字符串数组。按分组重置时使用。
- `user_ids`: 整数数组。按用户重置时使用。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 重置结果对象，结构由服务层返回。

## 失败响应

- HTTP 400: `success=false`，`message=无效的参数`。
- HTTP 200: `success=false`，`message` 为范围校验或重置错误。

