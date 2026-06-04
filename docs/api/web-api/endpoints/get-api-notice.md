---
method: GET
path: /api/notice
auth: public
handler: controller.GetNotice
source: router/api-router.go:34
request:
  body: none
response:
  success_http_status: 200
  envelope: common
  data: string
---

# GET `/api/notice`

获取公告内容。当前实现固定返回空字符串。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 字符串。当前固定为 `""`。

## 错误响应

当前控制器没有显式错误分支。

