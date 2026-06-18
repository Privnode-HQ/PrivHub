---
method: POST
path: /api/r2s/channel-bindings
auth: admin
handler: controller.CreateR2SChannelBinding
source: router/api-router.go:300
request:
  body:
    content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/r2s/channel-bindings`

管理员创建供应商与渠道的 R2S 绑定，并设置该渠道的供应商上游分组倍率。
该倍率会在收入识别记录中快照，后续修改不会影响已识别历史数据。

## 请求体字段

- `supplier_id`: R2S 供应商 ID，必填。
- `channel_id`: PrivHub 渠道 ID，必填。
- `channel_name_snapshot`: 可选。为空时使用当前渠道名称。
- `upstream_group_name`: 可选。为空时使用渠道分组中的第一个分组，缺省为
  `default`。
- `group_multiplier`: 上游分组倍率，必须大于 `0`，为空时为 `1`。
- `status`: 可选。`active` 或 `disabled`，为空时为 `active`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 新建的渠道绑定对象，字段同 GET
  `/api/r2s/channel-bindings` 的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: 参数错误、供应商不存在、渠道不存在、倍率不正确、渠道已绑定
  到启用中的 R2S 供应商，或数据库错误。
