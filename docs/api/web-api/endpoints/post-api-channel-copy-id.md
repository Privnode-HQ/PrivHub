---
method: POST
path: /api/channel/copy/:id
auth: admin
handler: controller.CopyChannel
source: router/api-router.go:216
request:
  path_params:
    - id
  query_params:
    - suffix
    - reset_balance
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/channel/copy/:id`

复制一个已有渠道，包括密钥。

## 路径参数字段

- `id`: 原渠道 ID，必须是整数。

## 查询参数字段

- `suffix`: 新渠道名称后缀，默认 `_复制`。
- `reset_balance`: 布尔值，默认 `true`。为 `true` 时新渠道 `balance` 和 `used_quota` 置 `0`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 新渠道 ID。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `invalid id`、原渠道查询错误或插入失败原因。

