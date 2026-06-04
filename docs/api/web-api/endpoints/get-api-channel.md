---
method: GET
path: /api/channel/
auth: admin
handler: controller.GetAllChannels
source: router/api-router.go:193
request:
  query_params:
    - p
    - page_size
    - id_sort
    - tag_mode
    - status
    - type
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/channel/`

管理员分页查看渠道列表。响应不返回渠道密钥 `key`。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。
- `id_sort`: 布尔值。`true` 时按 ID 倒序，否则按 `priority desc`。
- `tag_mode`: 布尔值。`true` 时先按 tag 分页，再展开 tag 下渠道。
- `status`: 状态过滤。`enabled` 或 `1` 表示仅启用；`disabled` 或 `0` 表示非启用；其他值或空表示全部。
- `type`: 渠道类型整数；为空表示全部类型。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.items`: 渠道数组。
- `data.items[].id`: 渠道 ID。
- `data.items[].type`: 渠道类型整数。
- `data.items[].openai_organization`: OpenAI organization，可为 `null`。
- `data.items[].test_model`: 测试模型，可为 `null`。
- `data.items[].status`: 渠道状态，`0` 未知，`1` 启用，`2` 手动禁用，`3` 自动禁用。
- `data.items[].name`: 渠道名称。
- `data.items[].weight`: 负载权重。
- `data.items[].created_time`: 创建时间 Unix 秒。
- `data.items[].test_time`: 最近测试时间 Unix 秒。
- `data.items[].response_time`: 最近测试响应耗时，毫秒。
- `data.items[].base_url`: 自定义基础地址；为空时使用类型默认地址。
- `data.items[].other`: 渠道特定扩展配置，例如区域 JSON。
- `data.items[].balance`: 渠道余额，单位 USD。
- `data.items[].balance_updated_time`: 余额更新时间 Unix 秒。
- `data.items[].models`: 逗号分隔模型列表。
- `data.items[].group`: 逗号分隔用户分组。
- `data.items[].used_quota`: 已用额度。
- `data.items[].model_mapping`: 模型映射 JSON 字符串，可为 `null`。
- `data.items[].status_code_mapping`: 状态码映射 JSON 字符串。
- `data.items[].priority`: 调度优先级。
- `data.items[].auto_ban`: 是否自动禁用，`1` 是，`0` 否。
- `data.items[].other_info`: 运行时扩展信息 JSON 字符串。
- `data.items[].tag`: 渠道标签，可为 `null`。
- `data.items[].setting`: 渠道额外设置 JSON 字符串。
- `data.items[].param_override`: 请求参数覆盖 JSON 字符串。
- `data.items[].header_override`: 请求头覆盖 JSON 字符串。
- `data.items[].remark`: 备注，最长 255 字符。
- `data.items[].channel_info.is_multi_key`: 是否多密钥渠道。
- `data.items[].channel_info.multi_key_size`: 多密钥数量。
- `data.items[].channel_info.multi_key_status_list`: key 索引到状态的映射，`1` 启用，`2` 手动禁用，`3` 自动禁用。
- `data.items[].channel_info.multi_key_polling_index`: 轮询模式下的下一个索引。
- `data.items[].channel_info.multi_key_mode`: 多密钥模式，例如随机或轮询模式。
- `data.items[].settings`: 其他渠道设置 JSON 字符串。
- `data.total`: 总数。`tag_mode=true` 时是 tag 总数。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.type_counts`: 按渠道类型统计的数量映射。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。

