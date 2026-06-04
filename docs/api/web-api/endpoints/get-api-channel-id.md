---
method: GET
path: /api/channel/:id
auth: admin
handler: controller.GetChannel
source: router/api-router.go:197
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/channel/:id`

按 ID 查看单个渠道。响应不返回渠道密钥 `key`。

## 路径参数字段

- `id`: 渠道 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 渠道 ID。
- `data.type`: 渠道类型。
- `data.openai_organization`: OpenAI organization，可为 `null`。
- `data.test_model`: 测试模型，可为 `null`。
- `data.status`: `0` 未知，`1` 启用，`2` 手动禁用，`3` 自动禁用。
- `data.name`: 渠道名称。
- `data.weight`: 权重。
- `data.created_time`: 创建时间。
- `data.test_time`: 最近测试时间。
- `data.response_time`: 最近响应耗时毫秒。
- `data.base_url`: 基础地址。
- `data.other`: 渠道特定扩展字段。
- `data.balance`: 余额。
- `data.balance_updated_time`: 余额更新时间。
- `data.models`: 模型列表字符串。
- `data.group`: 分组列表字符串。
- `data.used_quota`: 已用额度。
- `data.model_mapping`: 模型映射。
- `data.status_code_mapping`: 状态码映射。
- `data.priority`: 优先级。
- `data.auto_ban`: 自动禁用开关。
- `data.other_info`: 扩展信息。
- `data.tag`: 标签。
- `data.setting`: 渠道设置 JSON。
- `data.param_override`: 参数覆盖 JSON。
- `data.header_override`: 请求头覆盖 JSON。
- `data.remark`: 备注。
- `data.channel_info.is_multi_key`: 是否多密钥。
- `data.channel_info.multi_key_size`: 密钥数量。
- `data.channel_info.multi_key_status_list`: 多密钥状态映射。
- `data.channel_info.multi_key_polling_index`: 轮询索引。
- `data.channel_info.multi_key_mode`: 多密钥模式。
- `data.settings`: 其他设置 JSON。

## 失败响应

- `success`: `false`。
- `message`: `id` 非整数、渠道不存在或数据库错误。

