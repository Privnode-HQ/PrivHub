---
method: GET
path: /api/channel/tag/models
auth: admin
handler: controller.GetTagModels
source: router/api-router.go:215
request:
  query_params:
    - tag
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/channel/tag/models`

获取指定标签下模型数量最多的渠道模型字符串。

## 查询参数字段

- `tag`: 渠道标签，必填。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 逗号分隔模型字符串；若标签下渠道都没有模型，则为空字符串。

## 失败响应

- HTTP `400`: `success=false`，`message=tag不能为空`。
- HTTP `500`: `success=false`，`message` 为查询失败原因。

