---
method: POST
path: /api/token/
auth: user
handler: controller.AddToken
source: router/api-router.go:225
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/token/`

创建当前用户 API Token。

## 请求体字段

- `name`: 字符串，必填。Token 名称，最长 30 个字符。
- `expired_time`: 整数。过期时间 Unix 秒，`-1` 表示永不过期。
- `remain_quota`: 整数。初始剩余额度。
- `unlimited_quota`: 布尔值。是否无限额度。
- `model_limits_enabled`: 布尔值。是否启用模型限制。
- `model_limits`: 字符串。允许模型列表或模型限制配置。
- `allow_ips`: 字符串或 `null`。允许 IP 列表。
- `group`: 字符串。主分组；会根据 `groups` 规范化。
- `groups`: 字符串数组。可用分组列表。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `令牌名称过长`、`生成令牌失败`、训练数据分组同意错误，或创建错误。

