---
method: GET
path: /api/ratio_sync/channels
auth: admin
handler: controller.GetSyncableChannels
source: router/api-router.go:187
request:
  query_params: []
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/ratio_sync/channels`

获取可用于倍率同步比对的上游渠道列表。

## 请求字段

无请求体，无查询参数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 可同步渠道数组。
- `data[].id`: 渠道 ID；特殊值 `-100` 表示官方倍率预设。
- `data[].name`: 渠道名称。
- `data[].base_url`: 渠道基础地址。
- `data[].status`: 渠道状态，通常 `1` 表示启用，非 `1` 表示不可用或禁用。

## 失败响应

- `success`: `false`。
- `message`: 查询渠道失败时的错误信息。

