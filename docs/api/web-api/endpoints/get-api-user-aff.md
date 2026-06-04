---
method: GET
path: /api/user/aff
auth: user
handler: controller.GetAffCode
source: router/api-router.go:99
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
  data: string
---

# GET `/api/user/aff`

获取当前用户邀请码；如果还没有邀请码，后端会生成并保存。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 字符串。当前用户邀请码。

## 失败响应

- `success`: `false`。
- `message`: 用户查询或更新错误。

