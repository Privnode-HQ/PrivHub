---
method: GET
path: /api/uptime/status
auth: public
handler: controller.GetUptimeKumaStatus
source: router/api-router.go:19
request:
  body: none
response:
  success_http_status: 200
  envelope: common
  data: UptimeGroupResult[]
---

# GET `/api/uptime/status`

聚合配置中的 Uptime Kuma 状态页，供前端状态面板展示。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 数组。没有配置 Uptime Kuma 分组时为空数组。
- `data[].categoryName`: 分组显示名，来自控制台配置。
- `data[].monitors`: 该分组下的监控项数组。
- `data[].monitors[].name`: 监控项名称。
- `data[].monitors[].uptime`: 24 小时可用率数值，来自 Uptime Kuma `uptimeList`。
- `data[].monitors[].status`: 最新心跳状态码，取 Uptime Kuma 最新 heartbeat 的 `status`。
- `data[].monitors[].group`: Uptime Kuma 公共分组名称；为空时省略。

## 错误响应

单个 Uptime Kuma 分组抓取失败时不会让整个接口失败，该分组会返回空 `monitors`。

