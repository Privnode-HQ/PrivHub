---
method: GET
path: /api/r2s/channel-bindings
auth: admin
handler: controller.GetR2SChannelBindings
source: router/api-router.go:299
request:
  query_params:
    - p
    - page_size
    - supplier_id
    - channel_id
    - keyword
    - status
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/r2s/channel-bindings`

管理员分页查看 R2S 供应商与渠道的本地绑定关系。一个供应商可关联多个
渠道；每个启用中的渠道只能绑定一个 R2S 供应商。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。
- `supplier_id`: 可选。按供应商 ID 过滤。
- `channel_id`: 可选。按渠道 ID 过滤。
- `keyword`: 可选。按渠道名称快照或上游分组名称搜索。
- `status`: 可选。`active` 或 `disabled`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 总数。
- `data.items[]`: 渠道绑定数组。
- `data.items[].id`: 绑定 ID。
- `data.items[].supplier_id`: R2S 供应商 ID。
- `data.items[].supplier_name`: 当前供应商名称。
- `data.items[].channel_id`: PrivHub 渠道 ID。
- `data.items[].channel_name_snapshot`: 创建或更新绑定时的渠道名称快照。
- `data.items[].upstream_group_name`: 供应商上游分组名称快照。
- `data.items[].group_multiplier`: 该渠道上游分组倍率。
- `data.items[].status`: 状态，枚举为 `active`、`disabled`。
- `data.items[].created_time`: 创建时间 Unix 秒。
- `data.items[].updated_time`: 更新时间 Unix 秒。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。
