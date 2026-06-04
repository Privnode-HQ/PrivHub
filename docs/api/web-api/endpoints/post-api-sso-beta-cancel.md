---
method: POST
path: /api/sso-beta/cancel
auth: none
handler: controller.SSOCancel
source: router/api-router.go:370
request:
  body: none
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/sso-beta/cancel`

取消 SSO 授权。

## 请求字段

无请求体，无查询参数。

## 成功响应字段

- `success`: `true`。
- `data.redirect_url`: 固定为 `/`。

## 失败响应

该控制器没有显式失败分支。

