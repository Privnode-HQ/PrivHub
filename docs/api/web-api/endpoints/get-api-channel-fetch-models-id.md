---
method: GET
path: /api/channel/fetch_models/:id
auth: admin
handler: controller.FetchUpstreamModels
source: router/api-router.go:212
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/channel/fetch_models/:id`

使用指定渠道的密钥从上游模型列表接口拉取模型名称。

## 路径参数字段

- `id`: 渠道 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 模型名称数组。
- `data[]`: 上游模型 ID；Gemini 渠道会去掉 `models/` 前缀。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `获取渠道密钥失败`、`解析响应失败`、渠道不存在、上游请求错误或 JSON 解析错误。

