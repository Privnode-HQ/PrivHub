---
method: POST
path: /api/user/topup
auth: user_turnstile
handler: controller.TopUp
source: router/api-router.go:103
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/topup`

使用兑换码给当前用户充值额度。

## 请求体字段

- `key`: 字符串，必填。兑换码。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 整数。兑换获得的额度数量。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `充值处理中，请稍后重试`，或兑换码无效、已用、过期、数据库错误。

