---
method: GET
path: /api/models/missing
auth: support
handler: controller.GetMissingModels
source: router/api-router.go:352
request:
  query_params: []
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/models/missing`

列出渠道引用但模型元数据表中不存在的模型名。

## 请求字段

无请求体，无查询参数。

## 成功响应字段

- `success`: `true`。
- `data`: 字符串数组。
- `data[]`: 缺失的模型名。

## 失败响应

- `success`: `false`。
- `message`: 查询缺失模型失败原因。

