---
method: GET
path: /api/usage/self/limits
auth: user
handler: controller.GetSelfUsageLimits
source: router/api-router.go:234
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/usage/self/limits`

获取当前用户用量限制快照。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 用量限制快照对象，结构由 `service.GetUserUsageLimitSnapshot` 返回。
- `data.group`: 若服务返回该字段，表示当前用户分组。
- `data.limits`: 若服务返回该字段，表示各维度限制数组或对象。
- `data.usage`: 若服务返回该字段，表示当前周期已用量。
- `data.reset_at`: 若服务返回该字段，表示下次重置时间。

## 失败响应

- `success`: `false`。
- `message`: 用户缓存或快照计算错误。

