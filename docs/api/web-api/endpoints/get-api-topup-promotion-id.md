---
method: GET
path: /api/topup-promotion/:id
auth: support
handler: controller.GetTopUpPromotion
source: router/api-router.go:277
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/topup-promotion/:id`

按 ID 查看单个充值促销活动。

## 路径参数字段

- `id`: 促销活动 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 活动 ID。
- `data.name`: 活动名称。
- `data.description`: 活动说明。
- `data.currency_code`: 促销币种。
- `data.rules`: 优惠规则数组，按 `min_amount` 升序规范化。
- `data.rules[].min_amount`: 规则门槛金额。
- `data.rules[].discount_type`: `fixed` 或 `percent`。
- `data.rules[].discount_value`: 优惠金额或百分比。
- `data.allowed_groups`: 允许使用的用户分组。
- `data.status`: 存储状态。
- `data.effective_status`: 实际状态，可能为 `active`、`paused`、`revoked`、`scheduled`、`expired`。
- `data.valid_from`: 生效时间 Unix 秒。
- `data.expires_at`: 过期时间 Unix 秒。
- `data.max_redemptions`: 活动总次数限制。
- `data.max_redemptions_per_user`: 单用户次数限制。
- `data.created_by_admin_id`: 创建管理员 ID。
- `data.created_time`: 创建时间。
- `data.updated_time`: 更新时间。
- `data.revoked_at`: 撤销时间。
- `data.revoked_by_admin_id`: 撤销管理员 ID。
- `data.revoke_reason`: 撤销原因。
- `data.code_count`: 促销码数量。
- `data.reserved_count`: 预约中次数。
- `data.used_count`: 已使用次数。

## 失败响应

- `success`: `false`。
- `message`: `id` 非整数、活动不存在或数据库错误。

