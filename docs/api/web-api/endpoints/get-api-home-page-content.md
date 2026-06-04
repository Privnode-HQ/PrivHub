---
method: GET
path: /api/home_page_content
auth: public
handler: controller.GetHomePageContent
source: router/api-router.go:39
request:
  body: none
response:
  success_http_status: 200
  envelope: common
  data: string
---

# GET `/api/home_page_content`

获取首页自定义内容。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 字符串。来自全局选项 `HomePageContent`，未配置时为空字符串。

## 错误响应

当前控制器没有显式错误分支。

