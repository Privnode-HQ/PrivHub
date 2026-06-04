---
method: GET
path: /api/models
auth: user
handler: controller.DashboardListModels
source: router/api-router.go:20
request:
  body: none
response:
  success_http_status: 200
  envelope: dashboard_models_legacy
---

# GET `/api/models`

获取仪表盘可展示的渠道 ID 到模型列表映射。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `data`: 对象。键为渠道 ID 或渠道映射键，值为该渠道模型名称数组。

## 失败响应

该控制器没有显式错误分支；鉴权失败由中间件处理。

