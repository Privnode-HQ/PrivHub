---
method: GET
path: /api/oauth/state
auth: public_rate_limited
handler: controller.GenerateOAuthCode
source: router/api-router.go:48
request:
  query_params:
    - aff
response:
  success_http_status: 200
  envelope: custom_success
  data: string
---

# GET `/api/oauth/state`

生成 OAuth state 并写入当前 session，用于第三方登录/绑定防 CSRF。

## 查询参数字段

- `aff`: 字符串，可选。邀请码；传入时保存到 session，供后续 OAuth 注册使用。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 字符串。12 位随机 state，后续 OAuth 回调必须带回。

## 失败响应

- `success`: `false`。
- `message`: session 保存错误。

