---
method: GET
path: /api/user/search/api-key
auth: support
handler: controller.SearchUserByAPIKey
source: router/api-router.go:126
request:
  query_params:
    - key
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/user/search/api-key`

按 API Key 查找所属用户和 Token 摘要。

## 查询参数字段

- `key`: 字符串，必填。API Key，可带或不带 `sk-` 前缀。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.user`: 用户对象，字段同用户模型安全查询结果。
- `data.token.id`: Token ID。
- `data.token.user_id`: Token 所属用户 ID。
- `data.token.name`: Token 名称。
- `data.token.group`: Token 分组。
- `data.token.status`: Token 状态，`1` 启用、`2` 禁用、`3` 过期、`4` 耗尽。
- `data.token.created_time`: 创建时间 Unix 秒。
- `data.token.accessed_time`: 最近访问时间 Unix 秒。
- `data.token.expired_time`: 过期时间 Unix 秒，`-1` 表示永不过期。
- `data.token.deleted`: 布尔值，Token 是否已软删除。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `请输入 API Key`，或查找错误。

