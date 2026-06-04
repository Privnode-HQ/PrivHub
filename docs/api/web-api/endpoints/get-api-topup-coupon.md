---
method: GET
path: /api/topup-coupon/
auth: support
handler: controller.GetAllTopUpCoupons
source: router/api-router.go:261
request:
  query_params:
    - p
    - page_size
    - keyword
    - status
    - user_id
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/topup-coupon/`

支持人员或管理员分页查看充值优惠券。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，最大值由分页工具限制。
- `keyword`: 可选关键字，用于按优惠券 ID、名称等信息筛选。
- `status`: 可选优惠券状态。常见值包括 `available`、`reserved`、`used`、`expired`、`revoked`。
- `user_id`: 可选绑定用户 ID，只返回绑定给该用户的优惠券。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页码。
- `data.page_size`: 每页条数。
- `data.total`: 命中的优惠券总数。
- `data.items`: 优惠券数组。
- `data.items[].id`: 优惠券 ID。
- `data.items[].name`: 优惠券名称。
- `data.items[].bound_user_id`: 绑定用户 ID。
- `data.items[].bound_username`: 绑定用户名，视图补充字段。
- `data.items[].deduction_amount`: 可抵扣金额。
- `data.items[].currency_code`: 抵扣币种，规范化为大写。
- `data.items[].status`: 数据库存储状态。
- `data.items[].effective_status`: 按有效期和预约状态计算后的实际状态。
- `data.items[].valid_from`: 生效时间 Unix 秒。
- `data.items[].expires_at`: 过期时间 Unix 秒，`0` 表示不过期。
- `data.items[].issued_by_admin_id`: 发放管理员 ID。
- `data.items[].issued_at`: 发放时间 Unix 秒。
- `data.items[].reserved_top_up_id`: 当前占用该券的充值订单 ID。
- `data.items[].reserved_at`: 占用时间 Unix 秒。
- `data.items[].used_top_up_id`: 实际使用该券的充值订单 ID。
- `data.items[].used_at`: 使用时间 Unix 秒。
- `data.items[].revoked_at`: 撤销时间 Unix 秒。
- `data.items[].revoked_by_admin_id`: 撤销管理员 ID。
- `data.items[].revoke_reason`: 撤销原因。
- `data.items[].created_time`: 创建时间 Unix 秒。
- `data.items[].updated_time`: 更新时间 Unix 秒。

## 失败响应

- `success`: `false`。
- `message`: 查询错误信息。

