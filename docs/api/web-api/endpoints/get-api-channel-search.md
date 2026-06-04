---
method: GET
path: /api/channel/search
auth: admin
handler: controller.SearchChannels
source: router/api-router.go:194
request:
  query_params:
    - keyword
    - group
    - model
    - status
    - id_sort
    - tag_mode
    - type
    - p
    - page_size
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/channel/search`

管理员搜索渠道。响应不返回渠道密钥 `key`。

## 查询参数字段

- `keyword`: 搜索关键字，可匹配渠道 ID、名称、密钥精确值或 `base_url`。
- `group`: 分组过滤；为空或 `null` 表示不按分组过滤。
- `model`: 模型名关键字，匹配 `models` 字段。
- `status`: 状态过滤。`enabled` 或 `1` 表示启用；`disabled` 或 `0` 表示非启用。
- `id_sort`: 布尔值。`true` 时按 ID 倒序。
- `tag_mode`: 布尔值。`true` 时先搜索 tag，再展开渠道。
- `type`: 渠道类型整数过滤。
- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数；小于等于 `0` 时使用 `20`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.items`: 当前页渠道数组。
- `data.items[].id`: 渠道 ID。
- `data.items[].type`: 渠道类型。
- `data.items[].status`: `0` 未知，`1` 启用，`2` 手动禁用，`3` 自动禁用。
- `data.items[].name`: 渠道名称。
- `data.items[].weight`: 权重。
- `data.items[].created_time`: 创建时间。
- `data.items[].test_time`: 最近测试时间。
- `data.items[].response_time`: 响应耗时毫秒。
- `data.items[].base_url`: 基础地址。
- `data.items[].balance`: 余额。
- `data.items[].models`: 模型列表字符串。
- `data.items[].group`: 分组字符串。
- `data.items[].used_quota`: 已用额度。
- `data.items[].model_mapping`: 模型映射。
- `data.items[].priority`: 优先级。
- `data.items[].auto_ban`: 自动禁用开关。
- `data.items[].tag`: 标签。
- `data.items[].setting`: 渠道额外设置。
- `data.items[].param_override`: 参数覆盖。
- `data.items[].header_override`: 请求头覆盖。
- `data.items[].remark`: 备注。
- `data.items[].channel_info`: 多密钥信息；搜索响应会清理多密钥禁用原因和时间。
- `data.total`: 搜索命中总数。
- `data.type_counts`: 搜索结果按渠道类型统计的数量。

## 失败响应

- `success`: `false`。
- `message`: 搜索失败原因。

