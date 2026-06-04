---
method: GET
path: /api/group/
auth: support
handler: controller.GetGroups
source: router/api-router.go:307
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
  data: string[]
---

# GET `/api/group/`

获取全部分组名称。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 字符串数组。来自分组倍率配置的分组名。

## 失败响应

当前控制器没有显式错误分支。

