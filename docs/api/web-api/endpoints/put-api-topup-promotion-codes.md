---
method: PUT
path: /api/topup-promotion/codes
auth: admin
handler: controller.UpdateTopUpPromotionCode
source: router/api-router.go:285
request:
  body: TopUpPromotionCodeRequest
response:
  success_http_status: 200
  envelope: common
---

# PUT `/api/topup-promotion/codes`

管理员更新或撤销单个充值促销码。

## 请求体字段

- `id`: 促销码 ID，必填。
- `action`: 操作类型。传 `revoke` 时撤销促销码；其他值执行普通更新。
- `revoke_reason`: 撤销原因，仅撤销时使用。
- `code`: 新促销码文本；非空时覆盖，会规范化为大写。
- `valid_from`: 新生效时间；大于 `0` 时覆盖。
- `expires_at`: 新过期时间；直接覆盖，`0` 表示不过期。
- `max_redemptions`: 总兑换限制。
- `max_redemptions_per_user`: 每用户兑换限制。
- `campaign_id`: 更新时不用于迁移所属活动。
- `codes`: 更新时忽略。
- `auto_code_count`: 更新时忽略。
- `code_prefix`: 更新时忽略。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 促销码 ID。
- `data.campaign_id`: 活动 ID。
- `data.campaign_name`: 活动名称。
- `data.code`: 当前促销码。
- `data.status`: 撤销后为 `revoked`。
- `data.effective_status`: 实际状态。
- `data.valid_from`: 生效时间。
- `data.expires_at`: 过期时间。
- `data.max_redemptions`: 总兑换限制。
- `data.max_redemptions_per_user`: 每用户限制。
- `data.revoked_at`: 撤销时间。
- `data.revoked_by_admin_id`: 撤销管理员 ID。
- `data.revoke_reason`: 撤销原因。
- `data.reserved_count`: 预约中次数。
- `data.used_count`: 已使用次数。
- `data.updated_time`: 更新时间。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、促销码不存在、`已撤销的促销码不能编辑`、格式校验错误或撤销失败。

