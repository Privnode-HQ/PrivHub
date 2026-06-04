---
method: GET
path: /api/usage/token/
auth: bearer_token
handler: controller.GetTokenUsage
source: router/api-router.go:239
request:
  headers:
    - Authorization
response:
  success_http_status: 200
  unauthorized_http_status: 401
  envelope: token_usage_legacy
---

# GET `/api/usage/token/`

用 API Key 查询该 Token 的额度使用情况。

## 请求头字段

- `Authorization`: 字符串，必填。格式必须为 `Bearer <token>`；`<token>` 可带或不带 `sk-` 前缀。

## 成功响应字段

- `code`: `true`。
- `message`: `ok`。
- `data.object`: 固定为 `token_usage`。
- `data.name`: Token 名称。
- `data.total_granted`: 总授予额度，等于剩余额度加已用额度。
- `data.total_used`: 已用额度。
- `data.total_available`: 剩余额度。
- `data.unlimited_quota`: 是否无限额度。
- `data.model_limits`: 模型限制映射。
- `data.model_limits_enabled`: 是否启用模型限制。
- `data.expires_at`: 过期时间 Unix 秒；永不过期时为 `0`。

## 失败响应

- HTTP 401: `success=false`，`message` 为 `No Authorization header` 或 `Invalid Bearer token`。
- HTTP 200: `success=false`，`message` 为 Token 查找错误。

