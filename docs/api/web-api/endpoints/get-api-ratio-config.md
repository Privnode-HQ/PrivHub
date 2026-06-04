---
method: GET
path: /api/ratio_config
auth: public_rate_limited
handler: controller.GetRatioConfig
source: router/api-router.go:54
request:
  body: none
response:
  success_http_status: 200
  failure_http_status: 403
  envelope: common
---

# GET `/api/ratio_config`

获取公开的倍率配置。仅当管理员启用倍率配置暴露时可用。

## 请求字段

无路径参数、查询参数或请求体。接口带 Critical 限流。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 对象。公开倍率配置数据，结构由 `ratio_setting.GetExposedData()` 决定。

## 失败响应

当倍率配置暴露未启用时返回 HTTP 403：

- `success`: `false`。
- `message`: `倍率配置接口未启用`。

