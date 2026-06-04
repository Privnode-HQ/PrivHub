---
method: GET
path: /api/about
auth: public
handler: controller.GetAbout
source: router/api-router.go:37
request:
  body: none
response:
  success_http_status: 200
  envelope: common
  data: string
---

# GET `/api/about`

获取关于页面内容。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 字符串。来自全局选项 `About`，未配置时为空字符串。

## 错误响应

当前控制器没有显式错误分支。

