---
method: PUT
path: /api/token/
auth: user
handler: controller.UpdateToken
source: router/api-router.go:226
request:
  query_params:
    - status_only
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# PUT `/api/token/`

更新当前用户 Token。

## 查询参数字段

- `status_only`: 字符串，可选。非空时只更新 `status`。

## 请求体字段

- `id`: 整数，必填。Token ID。
- `status`: 整数。Token 状态，`1` 启用、`2` 禁用、`3` 过期、`4` 耗尽。
- `name`: 字符串。名称，最长 30 个字符。
- `expired_time`: 整数。过期时间 Unix 秒，`-1` 表示永不过期。
- `remain_quota`: 整数。剩余额度。
- `unlimited_quota`: 布尔值。是否无限额度。
- `model_limits_enabled`: 布尔值。是否启用模型限制。
- `model_limits`: 字符串。模型限制。
- `allow_ips`: 字符串或 `null`。允许 IP 列表。
- `group`: 字符串。主分组。
- `groups`: 字符串数组。分组列表。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 更新后的 Token 对象。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `令牌名称过长`、`令牌已过期，无法启用，请先修改令牌过期时间，或者设置为永不过期`、`令牌可用额度已用尽，无法启用，请先修改令牌剩余额度，或者设置为无限额度`、训练数据分组同意错误，或更新错误。

