---
method: PUT
path: /api/topup-coupon/
auth: admin
handler: controller.UpdateTopUpCoupon
source: router/api-router.go:269
request:
  body: TopUpCouponRequest
response:
  success_http_status: 200
  envelope: common
---

# PUT `/api/topup-coupon/`

管理员更新或撤销充值优惠券。

## 请求体字段

- `id`: 优惠券 ID，必填。
- `action`: 操作类型。传 `revoke` 时执行撤销；为空或其他值时执行普通编辑。
- `revoke_reason`: 撤销原因，仅 `action=revoke` 时使用。
- `name`: 新名称；普通编辑时非空才覆盖。
- `bound_user_id`: 新绑定用户 ID；大于 `0` 且不同于原值时会校验用户存在后覆盖。
- `deduction_amount`: 新抵扣金额；大于 `0` 时覆盖。
- `currency_code`: 新币种；可规范化为有效币种时覆盖。
- `valid_from`: 新生效时间；大于 `0` 时覆盖。
- `expires_at`: 新过期时间；会直接覆盖，`0` 表示不过期。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 更新或撤销后的优惠券详情。
- `data.id`: 优惠券 ID。
- `data.status`: 撤销后为 `revoked`；普通编辑可能保持原状态，过期券恢复有效期后可回到 `available`。
- `data.effective_status`: 实际状态。
- `data.revoked_at`: 撤销时间，仅撤销后有值。
- `data.revoked_by_admin_id`: 撤销管理员 ID。
- `data.revoke_reason`: 撤销原因。
- `data.updated_time`: 更新时间。
- `data.name`: 名称。
- `data.bound_user_id`: 绑定用户 ID。
- `data.deduction_amount`: 抵扣金额。
- `data.currency_code`: 币种。
- `data.valid_from`: 生效时间。
- `data.expires_at`: 过期时间。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、优惠券不存在、`已使用的优惠券不能编辑`、`支付中的优惠券不能编辑`、`目标用户不存在`、撤销失败或模型校验错误。

