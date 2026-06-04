---
method: GET
path: /api/token/
auth: user
handler: controller.GetAllTokens
source: router/api-router.go:222
request:
  query_params:
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/token/`

分页获取当前用户的 API Token。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，最大 `100`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页码。
- `data.page_size`: 实际每页条数。
- `data.total`: Token 总数。
- `data.items`: Token 数组。
- `data.items[].id`: Token ID。
- `data.items[].user_id`: 所属用户 ID。
- `data.items[].key`: Token 明文 key；列表接口可能返回数据库值。
- `data.items[].status`: 状态，`1` 启用、`2` 禁用、`3` 过期、`4` 耗尽。
- `data.items[].name`: 名称。
- `data.items[].created_time`: 创建时间 Unix 秒。
- `data.items[].accessed_time`: 最近访问时间 Unix 秒。
- `data.items[].expired_time`: 过期时间 Unix 秒，`-1` 表示永不过期。
- `data.items[].remain_quota`: 剩余额度。
- `data.items[].unlimited_quota`: 是否无限额度。
- `data.items[].model_limits_enabled`: 是否启用模型限制。
- `data.items[].model_limits`: 模型限制字符串。
- `data.items[].allow_ips`: 允许 IP 列表字符串或 `null`。
- `data.items[].used_quota`: 已用额度。
- `data.items[].group`: 主分组。
- `data.items[].groups`: 分组数组。

## 失败响应

- `success`: `false`。
- `message`: 查询错误。

