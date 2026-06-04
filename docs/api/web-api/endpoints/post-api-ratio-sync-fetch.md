---
method: POST
path: /api/ratio_sync/fetch
auth: admin
handler: controller.FetchUpstreamRatios
source: router/api-router.go:188
request:
  body: dto.UpstreamRequest
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/ratio_sync/fetch`

从指定上游或渠道抓取倍率配置，并与本地暴露的倍率数据计算差异。

## 请求体字段

- `channel_ids`: 渠道 ID 数组。未提供 `upstreams` 时，后端会按这些 ID 查询渠道并使用其 `base_url`。
- `upstreams`: 直接指定的上游列表；非空时优先于 `channel_ids`。
- `upstreams[].id`: 上游 ID，可选，用于展示名称后缀。
- `upstreams[].name`: 上游名称，必填。
- `upstreams[].base_url`: 上游基础地址，必须以 `http` 开头。
- `upstreams[].endpoint`: 抓取端点；为空时默认 `/api/ratio_config`，也可传完整 URL。
- `timeout`: 单个上游请求超时时间秒；小于等于 `0` 时默认 `10`。

## 成功响应字段

- `success`: `true`。
- `data.differences`: 差异对象，按模型名聚合。
- `data.differences.<model>.<ratio_type>.current`: 本地当前值，可能为数字或 `null`。
- `data.differences.<model>.<ratio_type>.upstreams`: 各上游返回值，值可能是数字、`same` 或 `null`。
- `data.differences.<model>.<ratio_type>.confidence`: 各上游值可信度布尔表。
- `data.test_results`: 上游连通性测试结果数组。
- `data.test_results[].name`: 上游显示名称。
- `data.test_results[].status`: `success` 或 `error`。
- `data.test_results[].error`: 错误详情，仅失败时返回。

## 特殊成功但业务失败响应

- `success`: `false`。
- `message`: `无有效上游渠道`，表示请求体和本地渠道均没有可抓取地址。

## 失败响应

- HTTP `400` 且 `success=false`: JSON 绑定失败。
- HTTP `500` 且 `success=false`: 通过 `channel_ids` 查询渠道失败。
- HTTP `200` 且 `success=false`: 上游不可用、无法解析或返回非成功结构时会进入 `test_results[].error`，接口整体仍可能成功返回差异数据。

