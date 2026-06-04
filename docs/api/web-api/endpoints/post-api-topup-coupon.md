---
method: POST
path: /api/topup-coupon/
auth: admin
handler: controller.AddTopUpCoupon
source: router/api-router.go:268
request:
  body: TopUpCouponRequest
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/topup-coupon/`

管理员向指定用户发放充值优惠券。

## 请求体字段

- `name`: 优惠券名称，去空白后长度必须为 `1-50` 个字符。
- `bound_user_id`: 目标用户 ID，必须存在。
- `deduction_amount`: 抵扣金额，必须通过优惠券模型校验并大于 `0`。
- `currency_code`: 抵扣币种，会规范化为大写；不能为空，长度和字符需符合币种格式。
- `valid_from`: 生效时间 Unix 秒；传 `0` 时后端使用当前时间。
- `expires_at`: 过期时间 Unix 秒；传 `0` 表示不过期，非 `0` 时需晚于生效时间。
- `id`: 创建时忽略。
- `action`: 创建时忽略。
- `revoke_reason`: 创建时忽略。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 创建后的优惠券详情。
- `data.id`: 新优惠券 ID。
- `data.name`: 名称。
- `data.bound_user_id`: 绑定用户 ID。
- `data.bound_username`: 绑定用户名。
- `data.deduction_amount`: 抵扣金额。
- `data.currency_code`: 抵扣币种。
- `data.status`: 初始为 `available`。
- `data.effective_status`: 实际状态。
- `data.valid_from`: 生效时间。
- `data.expires_at`: 过期时间。
- `data.issued_by_admin_id`: 当前管理员 ID。
- `data.issued_at`: 发放时间。
- `data.created_time`: 创建时间。
- `data.updated_time`: 更新时间。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、`优惠券名称长度必须在 1-50 之间`、`目标用户不存在`、`请选择优惠货币` 或模型校验错误。

