---
method: GET
path: /api/channel/test/:id
auth: admin
handler: controller.TestChannel
source: router/api-router.go:200
request:
  path_params:
    - id
  query_params:
    - model
    - endpoint_type
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/channel/test/:id`

测试单个渠道。

## 路径参数字段

- `id`: 渠道 ID，必须是整数。

## 查询参数字段

- `model`: 可选测试模型。为空时优先使用渠道 `test_model`，再使用渠道模型列表首项，最后使用默认测试模型。
- `endpoint_type`: 可选端点类型。用于指定测试请求格式，例如 OpenAI、Responses、Anthropic、Gemini、Embedding、Image、Rerank 等内部端点类型。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `time`: 本次测试耗时，单位秒。

## 失败响应

- `success`: `false`。
- `message`: 渠道不存在、测试请求转换失败、上游错误、不支持的渠道测试类型或其他错误文本。
- `time`: 失败耗时；本地前置错误时为 `0.0`。

