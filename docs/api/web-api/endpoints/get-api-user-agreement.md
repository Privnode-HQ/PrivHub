---
method: GET
path: /api/user-agreement
auth: public
handler: controller.GetUserAgreement
source: router/api-router.go:35
request:
  body: none
response:
  success_http_status: 200
  envelope: common
  data: string
---

# GET `/api/user-agreement`

获取用户协议正文。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 字符串。系统设置中的用户协议内容，未配置时为空字符串。

## 错误响应

当前控制器没有显式错误分支。

