---
method: GET
path: /api/topup-coupon/search
auth: support
handler: controller.SearchTopUpCoupons
source: router/api-router.go:262
request:
  query_params:
    - p
    - page_size
    - keyword
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/topup-coupon/search`

按关键字搜索充值优惠券，返回分页结果。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。
- `keyword`: 搜索关键字。可匹配优惠券 ID、名称或模型层支持的其他可检索字段。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页码。
- `data.page_size`: 每页条数。
- `data.total`: 搜索结果总数。
- `data.items`: 优惠券数组。
- `data.items[].id`: 优惠券 ID。
- `data.items[].name`: 名称。
- `data.items[].bound_user_id`: 绑定用户 ID。
- `data.items[].bound_username`: 绑定用户名。
- `data.items[].deduction_amount`: 抵扣金额。
- `data.items[].currency_code`: 抵扣币种。
- `data.items[].status`: 存储状态。
- `data.items[].effective_status`: 实际状态。
- `data.items[].valid_from`: 生效时间 Unix 秒。
- `data.items[].expires_at`: 过期时间 Unix 秒。
- `data.items[].issued_by_admin_id`: 发券管理员 ID。
- `data.items[].issued_at`: 发放时间。
- `data.items[].reserved_top_up_id`: 预约中的充值 ID。
- `data.items[].reserved_at`: 预约时间。
- `data.items[].used_top_up_id`: 使用该券的充值 ID。
- `data.items[].used_at`: 使用时间。
- `data.items[].revoked_at`: 撤销时间。
- `data.items[].revoked_by_admin_id`: 撤销管理员 ID。
- `data.items[].revoke_reason`: 撤销原因。
- `data.items[].created_time`: 创建时间。
- `data.items[].updated_time`: 更新时间。

## 失败响应

- `success`: `false`。
- `message`: 搜索失败原因。

