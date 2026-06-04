---
method: POST
path: /api/user/access_link/:token/consume
auth: public_rate_limited
handler: controller.ConsumeUserAccessLink
source: router/api-router.go:70
request:
  path_params:
    - token
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/access_link/:token/consume`

消费管理员生成的一次性账户访问链接，并把当前 session 切换到目标用户访问状态。

## 路径参数字段

- `token`: 字符串，必填。一次性访问链接 token，前后空白会被去除。

## 成功响应字段

- `success`: `true`。
- `message`: `访问链接已生效`。
- `data.grant_id`: 访问授权记录 ID。
- `data.user`: 当前用户响应对象，字段同 `GET /api/user/self`，会包含 access link session 状态。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `访问链接不能为空`、`访问链接无效`，或服务层返回的过期、已使用、数据库错误。

