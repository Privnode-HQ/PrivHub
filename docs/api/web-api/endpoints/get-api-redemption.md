---
method: GET
path: /api/redemption/
auth: support
handler: controller.GetAllRedemptions
source: router/api-router.go:246
request:
  query_params:
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/redemption/`

支持人员或管理员分页查看兑换码。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，最大 `100`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页码。
- `data.page_size`: 实际每页条数。
- `data.total`: 总条数。
- `data.items`: 兑换码数组。
- `data.items[].id`: 兑换码 ID。
- `data.items[].user_id`: 创建者用户 ID。
- `data.items[].key`: 兑换码 key。
- `data.items[].status`: 状态，`1` 启用、`2` 禁用、`3` 已使用。
- `data.items[].name`: 名称。
- `data.items[].quota`: 可兑换额度。
- `data.items[].created_time`: 创建时间 Unix 秒。
- `data.items[].redeemed_time`: 兑换时间 Unix 秒，未兑换时为 `0`。
- `data.items[].count`: 批量创建数量；列表中通常为 `0`。
- `data.items[].used_user_id`: 使用者用户 ID。
- `data.items[].expired_time`: 过期时间 Unix 秒，`0` 表示不过期。

## 失败响应

- `success`: `false`。
- `message`: 查询错误。

