---
method: POST
path: /api/option/rest_model_ratio
auth: root
handler: controller.ResetModelRatio
source: router/api-router.go:151
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/option/rest_model_ratio`

重置模型倍率配置为默认值。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: `重置模型倍率成功`。

## 失败响应

- `success`: `false`。
- `message`: 选项保存或倍率解析错误。

