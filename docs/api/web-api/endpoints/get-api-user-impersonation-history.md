---
method: GET
path: /api/user/impersonation/history
auth: user
handler: controller.GetImpersonationHistory
source: router/api-router.go:90
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/user/impersonation/history`

获取当前用户最近的 break glass 访问记录和当前开放的客服访问授权。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.break_glass_incidents`: break glass 事件数组，最多 20 条。
- `data.break_glass_incidents[].id`: 授权记录 ID。
- `data.break_glass_incidents[].operator_id`: 操作管理员 ID。
- `data.break_glass_incidents[].operator_username`: 操作管理员用户名。
- `data.break_glass_incidents[].started_at`: 开始时间。
- `data.break_glass_incidents[].ended_at`: 结束时间，仍在进行时为空。
- `data.break_glass_incidents[].active`: 是否仍活跃。
- `data.break_glass_incidents[].action_count`: 记录的请求动作数量。
- `data.break_glass_incidents[].actions`: 请求动作数组。
- `data.break_glass_incidents[].actions[].id`: 动作 ID。
- `data.break_glass_incidents[].actions[].created_at`: 动作发生时间。
- `data.break_glass_incidents[].actions[].method`: HTTP 方法。
- `data.break_glass_incidents[].actions[].path`: 请求路径。
- `data.break_glass_incidents[].actions[].route`: 匹配路由。
- `data.break_glass_incidents[].actions[].status_code`: HTTP 状态码。
- `data.break_glass_incidents[].actions[].success`: 请求是否成功。
- `data.open_support_access`: 可选。当前开放但未使用的客服访问授权。
- `data.open_support_access.id`: 授权 ID。
- `data.open_support_access.granted_expires_at`: 授权过期时间。
- `data.open_support_access.state`: 授权状态。

## 失败响应

- `success`: `false`。
- `message`: 历史记录查询错误。

