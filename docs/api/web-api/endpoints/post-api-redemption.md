---
method: POST
path: /api/redemption/
auth: admin
handler: controller.AddRedemption
source: router/api-router.go:253
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/redemption/`

管理员批量创建兑换码。

## 请求体字段

- `name`: 字符串，必填。兑换码名称，长度 1 到 20。
- `count`: 整数，必填。生成数量，范围 1 到 2000。
- `quota`: 整数，必填。每个兑换码可兑换额度。
- `expired_time`: 整数。过期时间 Unix 秒，`0` 表示不过期；不能早于当前时间。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 字符串数组。生成的兑换码 key 列表。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `兑换码名称长度必须在1-20之间`、`兑换码个数必须大于0`、`一次兑换码批量生成的个数不能大于 2000`、`过期时间不能早于当前时间`，或创建错误。
- `data`: 批量创建中途失败时包含已成功创建的 key 列表。

