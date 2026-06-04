---
method: GET
path: /api/channel/models
auth: admin
handler: controller.ChannelListModels
source: router/api-router.go:195
request:
  query_params: []
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/channel/models`

获取系统内置的渠道模型元数据列表。

## 请求字段

无请求体，无查询参数。

## 成功响应字段

- `success`: `true`。
- `data`: 模型数组。
- `data[].id`: 模型名称。
- `data[].object`: 对象类型，通常为 `model`。
- `data[].created`: 创建时间戳，内置模型使用固定值。
- `data[].owned_by`: 所属渠道或供应商名称。
- `data[].supported_endpoint_types`: 支持的端点类型数组；无数据时可为空或省略。

## 失败响应

该控制器不包含业务失败分支；认证失败由中间件返回。

