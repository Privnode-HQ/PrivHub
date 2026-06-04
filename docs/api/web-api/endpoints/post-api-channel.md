---
method: POST
path: /api/channel/
auth: admin
handler: controller.AddChannel
source: router/api-router.go:203
request:
  body: AddChannelRequest
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/channel/`

管理员新增渠道，支持单密钥、批量密钥和多密钥合并为单渠道。

## 请求体字段

- `mode`: 添加模式。`single` 单渠道；`batch` 按多行 key 批量创建多个渠道；`multi_to_single` 将多个 key 保存为一个多密钥渠道。
- `multi_key_mode`: 多密钥调度模式，仅 `multi_to_single` 使用。
- `batch_add_set_key_prefix_2_name`: 批量添加时是否把 key 前缀追加到渠道名称。
- `channel`: 渠道对象，必填。
- `channel.type`: 渠道类型整数。
- `channel.key`: 渠道密钥。批量模式下通常为换行分隔；Vertex AI JSON key 模式下可为 JSON 数组。
- `channel.openai_organization`: OpenAI organization，可选。
- `channel.test_model`: 测试模型，可选。
- `channel.status`: 初始状态，通常 `1` 启用。
- `channel.name`: 渠道名称。
- `channel.weight`: 权重。
- `channel.base_url`: 自定义基础地址。
- `channel.other`: 渠道特定配置；Vertex AI 必须是包含 `default` 的区域 JSON。
- `channel.models`: 逗号分隔模型列表；新增时每个模型名长度不能超过 `255`。
- `channel.group`: 逗号分隔可用用户分组。
- `channel.model_mapping`: 模型映射 JSON 字符串。
- `channel.status_code_mapping`: 状态码映射 JSON 字符串。
- `channel.priority`: 优先级。
- `channel.auto_ban`: 自动禁用开关，`1` 是，`0` 否。
- `channel.tag`: 渠道标签。
- `channel.setting`: 渠道额外设置 JSON 字符串。
- `channel.param_override`: 请求参数覆盖 JSON 字符串。
- `channel.header_override`: 请求头覆盖 JSON 字符串。
- `channel.remark`: 备注。
- `channel.settings`: 其他渠道设置 JSON 字符串。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 JSON 绑定错误、`channel cannot be empty`、`不支持的添加模式`、模型名过长、Vertex AI 配置错误、渠道设置格式错误或数据库错误。

