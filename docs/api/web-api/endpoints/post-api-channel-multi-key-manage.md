---
method: POST
path: /api/channel/multi_key/manage
auth: admin
handler: controller.ManageMultiKeys
source: router/api-router.go:217
request:
  body: MultiKeyManageRequest
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/channel/multi_key/manage`

管理多密钥渠道中的单个 key 或批量 key 状态。

## 请求体字段

- `channel_id`: 渠道 ID，必填。
- `action`: 操作类型。可用值：`get_key_status`、`disable_key`、`enable_key`、`enable_all_keys`、`disable_all_keys`、`delete_key`、`delete_disabled_keys`。
- `key_index`: key 索引，`disable_key`、`enable_key`、`delete_key` 必填。索引从 `0` 开始。
- `page`: 状态查询页码，仅 `get_key_status` 使用；默认 `1`。
- `page_size`: 状态查询每页条数，仅 `get_key_status` 使用；默认 `50`。
- `status`: 状态过滤，仅 `get_key_status` 使用。`1` 启用，`2` 手动禁用，`3` 自动禁用，`null` 表示全部。

## `get_key_status` 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.keys`: 当前页 key 状态数组。
- `data.keys[].index`: key 索引。
- `data.keys[].status`: key 状态，`1` 启用，`2` 手动禁用，`3` 自动禁用。
- `data.keys[].disabled_time`: 禁用时间 Unix 秒，仅非启用状态可能返回。
- `data.keys[].reason`: 禁用原因，仅非启用状态可能返回。
- `data.keys[].key_preview`: key 前 10 位预览，超过长度追加 `...`。
- `data.total`: 过滤后的 key 总数。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total_pages`: 总页数，至少为 `1`。
- `data.enabled_count`: 全部 key 中启用数量。
- `data.manual_disabled_count`: 全部 key 中手动禁用数量。
- `data.auto_disabled_count`: 全部 key 中自动禁用数量。

## 其他操作成功响应字段

- `success`: `true`。
- `message`: 操作结果文本，例如 `密钥已禁用`、`密钥已启用`、`已启用 N 个密钥`、`已禁用 N 个密钥`、`密钥已删除`、`已删除 N 个自动禁用的密钥`。
- `data`: 仅 `delete_disabled_keys` 返回删除数量。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `渠道不存在`、`该渠道不是多密钥模式`、`未指定要禁用的密钥索引`、`未指定要启用的密钥索引`、`未指定要删除的密钥索引`、`密钥索引超出范围`、`不能删除最后一个密钥`、`没有可禁用的密钥`、`没有需要删除的自动禁用密钥`、`不支持的操作` 或数据库错误。

