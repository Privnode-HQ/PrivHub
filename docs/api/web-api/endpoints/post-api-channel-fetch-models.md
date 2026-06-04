---
method: POST
path: /api/channel/fetch_models
auth: admin
handler: controller.FetchModels
source: router/api-router.go:213
request:
  body:
    - base_url
    - type
    - key
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/channel/fetch_models`

使用请求体中提供的上游地址和密钥拉取模型名称，不要求渠道已保存。

## 请求体字段

- `base_url`: 上游基础地址；为空时使用 `type` 对应的默认渠道地址。
- `type`: 渠道类型整数，用于查找默认基础地址。
- `key`: 上游密钥；会去除首尾空白，包含换行时只使用第一行。

## 成功响应字段

- `success`: `true`。
- `data`: 模型名称数组。
- `data[]`: 上游 `/v1/models` 返回的 `data[].id`。

## 失败响应

- HTTP `400`: `success=false`，`message=Invalid request`。
- HTTP `500`: `success=false`，`message` 为构造请求、上游请求、非 200 响应或 JSON 解码错误。

