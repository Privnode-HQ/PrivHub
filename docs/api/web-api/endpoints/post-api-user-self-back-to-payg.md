---
method: POST
path: /api/user/self/back_to_payg
auth: user
handler: controller.BackToPayAsYouGo
source: router/api-router.go:82
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/self/back_to_payg`

把当前用户从订阅分组切回默认 pay-as-you-go 分组。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `当前账户不是订阅分组，无需操作`，或用户更新错误。

