---
method: GET
path: /api/user/models
auth: user
handler: controller.GetUserModels
source: router/api-router.go:81
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
  data: string[]
---

# GET `/api/user/models`

获取当前用户可用分组启用的模型名称列表。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 字符串数组。去重后的模型名称列表。

## 失败响应

- `success`: `false`。
- `message`: 用户缓存查询错误。

