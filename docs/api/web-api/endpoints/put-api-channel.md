---
method: PUT
path: /api/channel/
auth: admin
handler: controller.UpdateChannel
source: router/api-router.go:204
request:
  body: PatchChannel
response:
  success_http_status: 200
  envelope: raw-json
---

# PUT `/api/channel/`

管理员更新渠道。多密钥渠道会保留原 `channel_info`，并可追加或覆盖密钥。

## 请求体字段

- `id`: 渠道 ID，必填。
- `multi_key_mode`: 可选新多密钥模式；非空时覆盖原模式。
- `key_mode`: 多密钥渠道密钥处理方式。`append` 追加新 key；`replace` 或为空表示覆盖/默认行为。
- `type`: 渠道类型。
- `key`: 新密钥；多密钥追加时可为换行分隔 key 或 Vertex AI JSON。
- `openai_organization`: OpenAI organization。
- `test_model`: 测试模型。
- `status`: 状态，`1` 启用，`2` 手动禁用，`3` 自动禁用。
- `name`: 渠道名称。
- `weight`: 权重。
- `base_url`: 基础地址。
- `other`: 渠道特定配置。
- `balance`: 余额。
- `models`: 逗号分隔模型列表。
- `group`: 逗号分隔分组。
- `used_quota`: 已用额度。
- `model_mapping`: 模型映射。
- `status_code_mapping`: 状态码映射。
- `priority`: 优先级。
- `auto_ban`: 自动禁用开关。
- `other_info`: 扩展信息。
- `tag`: 标签。
- `setting`: 渠道额外设置。
- `param_override`: 参数覆盖 JSON。
- `header_override`: 请求头覆盖 JSON。
- `remark`: 备注。
- `settings`: 其他设置。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 更新后的渠道对象。
- `data.key`: 被置为空字符串，不返回真实密钥。
- `data.channel_info`: 更新后的多密钥信息；禁用原因和禁用时间会在响应中清理。
- `data.id`: 渠道 ID。
- `data.status`: 当前状态。
- `data.models`: 当前模型列表。
- `data.group`: 当前分组。
- `data.updated_time`: 渠道模型未单独暴露该字段；实际更新时间体现在数据库保存。

## 失败响应

- `success`: `false`。
- `message`: JSON 绑定错误、渠道不存在、渠道设置格式错误、Vertex AI 配置错误、多密钥追加解析失败或数据库错误。

