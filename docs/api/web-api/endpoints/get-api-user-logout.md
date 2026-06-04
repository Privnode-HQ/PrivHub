---
method: GET
path: /api/user/logout
auth: session_optional
handler: controller.Logout
source: router/api-router.go:72
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/user/logout`

清除当前 Web session，并停止当前模拟登录或 Access Link 会话状态。

## 请求字段

无路径参数、查询参数或请求体。路由本身没有 `UserAuth` 中间件；未登录调用也会执行 session 清理。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: session 保存失败时的错误文本。

