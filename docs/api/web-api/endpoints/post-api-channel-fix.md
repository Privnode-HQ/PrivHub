---
method: POST
path: /api/channel/fix
auth: admin
handler: controller.FixChannelsAbilities
source: router/api-router.go:211
request:
  body: none
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/channel/fix`

修复渠道能力表，使渠道模型能力与当前渠道配置重新对齐。

## 请求字段

无请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.success`: 修复成功数量。
- `data.fails`: 修复失败数量。

## 失败响应

- `success`: `false`。
- `message`: 能力修复失败原因。

