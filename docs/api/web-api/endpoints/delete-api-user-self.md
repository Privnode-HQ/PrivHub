---
method: DELETE
path: /api/user/self
auth: user
handler: controller.DeleteSelf
source: router/api-router.go:84
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# DELETE `/api/user/self`

注销当前用户账号。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `不能删除超级管理员账户`，或删除错误。

