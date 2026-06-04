---
method: GET
path: /api/topup-coupon/:id
auth: support
handler: controller.GetTopUpCoupon
source: router/api-router.go:263
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/topup-coupon/:id`

按 ID 查看单张充值优惠券详情。

## 路径参数字段

- `id`: 优惠券 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 优惠券 ID。
- `data.name`: 优惠券名称。
- `data.bound_user_id`: 绑定用户 ID。
- `data.bound_username`: 绑定用户名。
- `data.deduction_amount`: 可抵扣金额。
- `data.currency_code`: 抵扣币种。
- `data.status`: 存储状态，常见值 `available`、`reserved`、`used`、`expired`、`revoked`。
- `data.effective_status`: 实际状态；会结合时间窗口和预约状态计算。
- `data.valid_from`: 生效时间 Unix 秒。
- `data.expires_at`: 过期时间 Unix 秒，`0` 表示不过期。
- `data.issued_by_admin_id`: 发放管理员 ID。
- `data.issued_at`: 发放时间 Unix 秒。
- `data.reserved_top_up_id`: 当前占用订单 ID。
- `data.reserved_at`: 占用时间 Unix 秒。
- `data.used_top_up_id`: 已使用订单 ID。
- `data.used_at`: 使用时间 Unix 秒。
- `data.revoked_at`: 撤销时间 Unix 秒。
- `data.revoked_by_admin_id`: 撤销管理员 ID。
- `data.revoke_reason`: 撤销原因。
- `data.created_time`: 创建时间 Unix 秒。
- `data.updated_time`: 更新时间 Unix 秒。

## 失败响应

- `success`: `false`。
- `message`: `id` 非整数、优惠券不存在或数据库错误。

