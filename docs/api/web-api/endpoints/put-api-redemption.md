---
method: PUT
path: /api/redemption/
auth: admin
handler: controller.UpdateRedemption
source: router/api-router.go:254
request:
  query_params:
    - status_only
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# PUT `/api/redemption/`

更新兑换码信息或仅更新状态。

## 查询参数字段

- `status_only`: 字符串，可选。非空时只更新 `status`。

## 请求体字段

- `id`: 整数，必填。兑换码 ID。
- `name`: 字符串。名称。
- `quota`: 整数。可兑换额度。
- `expired_time`: 整数。过期时间 Unix 秒，`0` 表示不过期。
- `status`: 整数。状态，`1` 启用、`2` 禁用、`3` 已使用；`status_only` 非空时使用。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 更新后的兑换码对象。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `过期时间不能早于当前时间`，或兑换码查询/更新错误。

