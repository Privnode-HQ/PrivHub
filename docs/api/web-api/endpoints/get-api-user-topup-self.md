---
method: GET
path: /api/user/topup/self
auth: user
handler: controller.GetUserTopUps
source: router/api-router.go:102
request:
  query_params:
    - keyword
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/user/topup/self`

分页获取当前用户充值订单记录。

## 查询参数字段

- `keyword`: 字符串，可选。存在时按本用户订单搜索。
- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，最大 `100`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页码。
- `data.page_size`: 实际每页条数。
- `data.total`: 总条数。
- `data.items`: 充值订单数组。
- `data.items[].id`: 订单 ID。
- `data.items[].user_id`: 用户 ID。
- `data.items[].amount`: 充值额度数量。
- `data.items[].money`: 订单计费金额。
- `data.items[].trade_no`: 充值单号。
- `data.items[].payment_method`: 支付方式。
- `data.items[].create_time`: 创建时间 Unix 秒。
- `data.items[].complete_time`: 完成时间 Unix 秒。
- `data.items[].status`: 订单状态，可能值 `pending`、`success`、`expired`。
- `data.items[].coupon_id`: 优惠券 ID。
- `data.items[].coupon_name`: 优惠券名称。
- `data.items[].original_money`: 折扣前金额。
- `data.items[].platform_discount`: 平台折扣金额。
- `data.items[].coupon_discount`: 优惠券抵扣金额。
- `data.items[].promotion_campaign_id`: 促销活动 ID。
- `data.items[].promotion_code_id`: 促销码 ID。
- `data.items[].promotion_code`: 促销码文本。
- `data.items[].promotion_discount`: 促销码抵扣金额。
- `data.items[].promotion_redemption_id`: 促销码核销记录 ID。
- `data.items[].processing_fee`: 手续费。
- `data.items[].pay_money`: 最终应付金额。

## 失败响应

- `success`: `false`。
- `message`: 查询错误。

