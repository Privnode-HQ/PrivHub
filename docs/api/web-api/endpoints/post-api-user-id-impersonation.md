---
method: POST
path: /api/user/:id/impersonation
auth: admin
handler: controller.StartUserImpersonation
source: router/api-router.go:135
request:
  path_params:
    - id
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/:id/impersonation`

管理员请求或开始模拟访问指定用户。可能直接激活、进入 break glass，或生成等待用户批准的请求。

## 路径参数字段

- `id`: 整数，必填。目标用户 ID。

## 请求体字段

- `read_only`: 布尔值。是否请求只读访问。
- `break_glass`: 布尔值。是否请求 break glass 紧急访问。

## 成功响应字段

直接激活或 break glass：

- `success`: `true`。
- `message`: 空字符串。
- `data.start_state`: 启动结果状态，例如 `activated` 或 `break_glass`。
- `data.grant_id`: 授权记录 ID。
- `data.user`: 目标用户当前响应对象，字段同 `GET /api/user/self`。

等待批准：

- `success`: `true`。
- `message`: `访问请求已发送，等待用户批准` 或 `该用户已有待处理的访问请求，请等待用户批准`。
- `data.start_state`: 启动结果状态。
- `data.grant_id`: 授权记录 ID。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的用户 ID`、`无效的请求参数`、`不能仿冒当前登录用户`、`目标用户当前不可访问`、`无权仿冒同级或更高权限的用户`，或服务层错误。

