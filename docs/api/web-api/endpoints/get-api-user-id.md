---
method: GET
path: /api/user/:id
auth: support
handler: controller.GetUser
source: router/api-router.go:127
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/user/:id`

获取单个用户详情。

## 路径参数字段

- `id`: 整数，必填。目标用户 ID。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 用户对象，字段同 `GET /api/user/` 的 `data.items[]`；控制器直接返回用户模型，可能包含更多绑定 ID 和设置字段，但不应依赖密码字段。

## 失败响应

- `success`: `false`。
- `message`: 可能为路径 ID 解析错误、用户不存在，或 `无权获取同级或更高等级用户的信息`。

