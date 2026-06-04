---
method: PUT
path: /api/topup-promotion/
auth: admin
handler: controller.UpdateTopUpPromotion
source: router/api-router.go:283
request:
  body: TopUpPromotionRequest
response:
  success_http_status: 200
  envelope: common
---

# PUT `/api/topup-promotion/`

管理员更新或撤销充值促销活动。

## 请求体字段

- `id`: 活动 ID，必填。
- `action`: 操作类型。传 `revoke` 时撤销活动；其他值执行普通更新。
- `revoke_reason`: 撤销原因，仅撤销时使用。
- `name`: 新名称，非空时覆盖。
- `description`: 新说明，会直接覆盖，可为空。
- `currency_code`: 新币种，非空时覆盖。
- `rules`: 新规则数组；非空时覆盖原规则。
- `rules[].min_amount`: 规则门槛。
- `rules[].discount_type`: `fixed` 或 `percent`。
- `rules[].discount_value`: 优惠金额或百分比。
- `allowed_groups`: 允许分组数组，会覆盖原值；空数组表示不限制。
- `valid_from`: 新生效时间；大于 `0` 时覆盖。
- `expires_at`: 新过期时间；直接覆盖，`0` 表示不过期。
- `max_redemptions`: 活动总兑换限制。
- `max_redemptions_per_user`: 单用户兑换限制。
- `codes`: 更新活动时忽略。
- `auto_code_count`: 更新活动时忽略。
- `code_prefix`: 更新活动时忽略。
- `code_valid_from`: 更新活动时忽略。
- `code_expires_at`: 更新活动时忽略。
- `code_max_redemptions`: 更新活动时忽略。
- `code_max_redemptions_per_user`: 更新活动时忽略。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 活动 ID。
- `data.name`: 活动名称。
- `data.description`: 活动说明。
- `data.currency_code`: 币种。
- `data.rules`: 当前规则。
- `data.allowed_groups`: 当前允许分组。
- `data.status`: 撤销后为 `revoked`；普通更新保留或修正原状态。
- `data.effective_status`: 实际状态。
- `data.valid_from`: 生效时间。
- `data.expires_at`: 过期时间。
- `data.max_redemptions`: 活动总兑换限制。
- `data.max_redemptions_per_user`: 单用户限制。
- `data.revoked_at`: 撤销时间。
- `data.revoked_by_admin_id`: 撤销管理员 ID。
- `data.revoke_reason`: 撤销原因。
- `data.updated_time`: 更新时间。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、活动不存在、`已撤销的促销活动不能编辑`、规则校验错误或撤销失败。

