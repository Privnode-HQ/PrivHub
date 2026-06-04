---
method: GET
path: /api/user/token
auth: user
handler: controller.GenerateAccessToken
source: router/api-router.go:85
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
  data: string
---

# GET `/api/user/token`

为当前用户重新生成用户级 Access Token。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 字符串。新的 Access Token；会覆盖用户原有 `access_token`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `生成失败`、`请重试，系统生成的 UUID 竟然重复了！`，或用户更新错误。

