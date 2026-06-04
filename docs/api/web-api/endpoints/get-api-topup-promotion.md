---
method: GET
path: /api/topup-promotion/
auth: support
handler: controller.GetAllTopUpPromotions
source: router/api-router.go:274
request:
  query_params:
    - p
    - page_size
    - keyword
    - status
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/topup-promotion/`

支持人员或管理员分页查看充值促销活动。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。
- `keyword`: 可选关键字；可按活动 ID 或名称过滤。
- `status`: 可选存储状态。常见值包括 `active`、`paused`、`revoked`，实际状态还可能展示 `scheduled`、`expired`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 总数。
- `data.items`: 促销活动数组。
- `data.items[].id`: 活动 ID。
- `data.items[].name`: 活动名称。
- `data.items[].description`: 活动说明。
- `data.items[].currency_code`: 促销币种。
- `data.items[].rules`: 规则数组。
- `data.items[].rules[].min_amount`: 命中该档规则的最低支付金额。
- `data.items[].rules[].discount_type`: 优惠类型，`fixed` 固定金额，`percent` 百分比。
- `data.items[].rules[].discount_value`: 固定优惠金额或百分比数值。
- `data.items[].allowed_groups`: 允许使用的用户分组；空数组表示不限制。
- `data.items[].status`: 存储状态。
- `data.items[].effective_status`: 按时间窗口计算后的实际状态。
- `data.items[].valid_from`: 生效时间 Unix 秒。
- `data.items[].expires_at`: 过期时间 Unix 秒，`0` 表示不过期。
- `data.items[].max_redemptions`: 活动总兑换次数限制，`0` 表示不限制。
- `data.items[].max_redemptions_per_user`: 单用户兑换次数限制，`0` 表示不限制。
- `data.items[].created_by_admin_id`: 创建管理员 ID。
- `data.items[].created_time`: 创建时间。
- `data.items[].updated_time`: 更新时间。
- `data.items[].revoked_at`: 撤销时间。
- `data.items[].revoked_by_admin_id`: 撤销管理员 ID。
- `data.items[].revoke_reason`: 撤销原因。
- `data.items[].code_count`: 活动下促销码数量。
- `data.items[].reserved_count`: 预约中的兑换数。
- `data.items[].used_count`: 已使用兑换数。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。

