---
method: POST
path: /api/topup-promotion/codes
auth: admin
handler: controller.AddTopUpPromotionCodes
source: router/api-router.go:284
request:
  body: TopUpPromotionCodeRequest
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/topup-promotion/codes`

管理员为已有充值促销活动新增促销码。

## 请求体字段

- `campaign_id`: 活动 ID，必填。
- `code`: 单个促销码文本，可选；会与 `codes` 合并。
- `codes`: 多个促销码文本，可选；会去重并规范化。
- `auto_code_count`: 自动生成数量，范围 `0-500`。当未提供任何手工码且该值为 `0` 时，模型层会要求至少产生一个促销码。
- `code_prefix`: 自动生成促销码前缀。
- `valid_from`: 促销码生效时间 Unix 秒；为 `0` 时使用当前时间。
- `expires_at`: 促销码过期时间 Unix 秒；`0` 表示不过期。
- `max_redemptions`: 单个促销码总兑换限制，`0` 表示不限制。
- `max_redemptions_per_user`: 单个促销码每用户兑换限制，`0` 表示不限制。
- `id`: 创建时忽略。
- `action`: 创建时忽略。
- `revoke_reason`: 创建时忽略。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 新增的促销码数组。
- `data[].id`: 促销码 ID。
- `data[].campaign_id`: 活动 ID。
- `data[].code`: 促销码文本。
- `data[].status`: 初始为 `active`。
- `data[].valid_from`: 生效时间。
- `data[].expires_at`: 过期时间。
- `data[].max_redemptions`: 总兑换限制。
- `data[].max_redemptions_per_user`: 每用户兑换限制。
- `data[].created_by_admin_id`: 创建管理员 ID。
- `data[].created_time`: 创建时间。
- `data[].updated_time`: 更新时间。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、`缺少促销活动 ID`、`自动生成数量必须在 0-500 之间`、促销码格式错误、促销码重复或数据库错误。

