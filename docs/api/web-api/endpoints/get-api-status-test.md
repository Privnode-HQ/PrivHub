---
method: GET
path: /api/status/test
auth: admin
handler: controller.TestStatus
source: router/api-router.go:21
request:
  body: none
response:
  success_http_status: 200
  failure_http_status: 503
  envelope: custom
---

# GET `/api/status/test`

管理员测试后端和数据库连通性，同时返回 HTTP 统计信息。

## 请求字段

无路径参数、查询参数或请求体。需要管理员鉴权，并会记录管理员审计。

## 成功响应字段

- `success`: `true`。
- `message`: 固定为 `Server is running`。
- `http_stats`: HTTP 请求统计对象，结构来自 `middleware.GetStats()`；用于观察当前进程的请求统计。

## 失败响应

数据库连接失败时返回 HTTP 503：

- `success`: `false`。
- `message`: 固定为 `数据库连接失败`。

