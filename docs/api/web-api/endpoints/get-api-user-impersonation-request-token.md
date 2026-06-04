---
method: GET
path: /api/user/impersonation/request/:token
auth: user
handler: controller.GetImpersonationRequest
source: router/api-router.go:87
request:
  path_params:
    - token
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/user/impersonation/request/:token`

目标用户查看管理员发起的模拟访问批准请求。

## 路径参数字段

- `token`: 字符串，必填。模拟访问批准 token。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 授权记录 ID。
- `data.state`: 状态，可能值 `pending`、`approved`、`rejected`、`cancelled`、`expired`、`active`、`completed`。
- `data.mode`: 模式，可能值 `standard`、`break_glass`。
- `data.requested_read_only`: 管理员请求是否只读。
- `data.operator_id`: 请求管理员 ID。
- `data.operator_username`: 请求管理员用户名。
- `data.operator_cah_id`: 请求管理员 CAH ID。
- `data.target_user_id`: 目标用户 ID。
- `data.target_username`: 目标用户名。
- `data.requested_at`: 请求时间。
- `data.granted_expires_at`: 批准窗口过期时间。
- `data.approved_at`: 批准时间，未批准时为空。
- `data.requires_verification`: 当前用户批准时是否需要先完成安全验证。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `访问请求不存在`、`无权查看该访问请求`，或数据库错误。

