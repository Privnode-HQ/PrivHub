---
method: POST
path: /api/topup-promotion/
auth: admin
handler: controller.AddTopUpPromotion
source: router/api-router.go:282
request:
  body: TopUpPromotionRequest
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/topup-promotion/`

管理员创建充值促销活动，并可同时创建促销码。

## 请求体字段

- `name`: 活动名称，必填，最多 `80` 个字符。
- `description`: 活动说明，最多 `255` 个字符。
- `currency_code`: 促销币种，必填，规范化为大写。
- `rules`: 优惠规则数组，至少一条。
- `rules[].min_amount`: 规则门槛金额，不能小于 `0`。
- `rules[].discount_type`: 优惠类型，必须为 `fixed` 或 `percent`。
- `rules[].discount_value`: 优惠值。`fixed` 时必须大于 `0`；`percent` 时必须大于 `0` 且不超过 `100`。
- `allowed_groups`: 允许使用的用户分组数组；空数组表示所有分组。
- `valid_from`: 活动生效时间 Unix 秒；传 `0` 时使用当前时间。
- `expires_at`: 活动过期时间 Unix 秒；`0` 表示不过期，非 `0` 必须晚于 `valid_from`。
- `max_redemptions`: 活动总兑换次数限制，`0` 表示不限制，不能小于 `0`。
- `max_redemptions_per_user`: 单用户兑换次数限制，`0` 表示不限制，不能小于 `0`。
- `codes`: 可选手工促销码数组；会去重并规范化为大写。
- `auto_code_count`: 自动生成促销码数量；为 `0` 且未提供 `codes` 时默认生成 `1` 个，最大 `500`。
- `code_prefix`: 自动生成促销码前缀。
- `code_valid_from`: 新建促销码生效时间；为 `0` 时继承活动生效时间。
- `code_expires_at`: 新建促销码过期时间；为 `0` 时继承活动过期时间。
- `code_max_redemptions`: 单个促销码总兑换限制。
- `code_max_redemptions_per_user`: 单个促销码每用户兑换限制。
- `id`: 创建时忽略。
- `action`: 创建时忽略。
- `revoke_reason`: 创建时忽略。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 创建后的活动详情。
- `data.id`: 活动 ID。
- `data.status`: 初始为 `active`。
- `data.effective_status`: 实际状态。
- `data.rules`: 规范化后的规则。
- `data.allowed_groups`: 规范化后的分组。
- `data.code_count`: 已创建促销码数量。
- `data.created_by_admin_id`: 当前管理员 ID。
- `data.created_time`: 创建时间。
- `data.updated_time`: 更新时间。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、名称/币种/规则/有效期/次数限制校验错误、促销码重复或数据库错误。

