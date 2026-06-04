---
method: POST
path: /api/message/
auth: admin
handler: controller.CreateMessage
source: router/api-router.go:176
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/message/`

管理员创建消息草稿。

## 请求体字段

- `title`: 字符串，必填。标题，去除空白后不能为空。
- `content`: 字符串，必填。消息内容，去除空白后不能为空。
- `target_type`: 字符串。目标类型，由服务层解析，例如全体、分组或指定用户。
- `target_groups`: 字符串数组。目标分组。
- `target_user_ids`: 整数数组。目标用户 ID。

## 成功响应字段

- `success`: `true`。
- `message`: `草稿已创建`。
- `data`: 创建后的消息对象。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`标题和内容不能为空`，或服务层校验错误。

