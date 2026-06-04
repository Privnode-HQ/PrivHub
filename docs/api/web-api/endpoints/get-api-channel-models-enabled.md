---
method: GET
path: /api/channel/models_enabled
auth: admin
handler: controller.EnabledListModels
source: router/api-router.go:196
request:
  query_params: []
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/channel/models_enabled`

获取当前已启用渠道可用的模型名列表。

## 请求字段

无请求体，无查询参数。

## 成功响应字段

- `success`: `true`。
- `data`: 字符串数组。
- `data[]`: 已启用模型名称。

## 失败响应

该控制器不包含业务失败分支；认证失败由中间件返回。

