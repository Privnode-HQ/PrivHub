---
method: PUT
path: /api/r2s/channel-bindings/:id
auth: admin
handler: controller.UpdateR2SChannelBinding
source: router/api-router.go:303
request:
  path_params:
    - id
  body:
    content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# PUT `/api/r2s/channel-bindings/:id`

管理员更新 R2S 渠道绑定。更新后的倍率只影响之后创建的收入识别记录；
已有收入识别记录使用各自保存的历史快照。

## 路径参数字段

- `id`: 渠道绑定 ID，必须是整数。

## 请求体字段

- `supplier_id`: R2S 供应商 ID，必填。
- `channel_id`: PrivHub 渠道 ID，必填。
- `channel_name_snapshot`: 可选。为空时使用当前渠道名称。
- `upstream_group_name`: 可选。为空时使用渠道分组中的第一个分组。
- `group_multiplier`: 上游分组倍率，必须大于 `0`。
- `status`: `active` 或 `disabled`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 更新后的渠道绑定对象，字段同 GET
  `/api/r2s/channel-bindings` 的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: ID 解析错误、参数错误、绑定不存在、供应商不存在、渠道不
  存在、倍率不正确、渠道已绑定到其他启用中的 R2S 供应商，或数据库错误。
