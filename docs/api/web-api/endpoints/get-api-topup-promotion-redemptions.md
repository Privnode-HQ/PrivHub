---
method: GET
path: /api/topup-promotion/redemptions
auth: support
handler: controller.GetTopUpPromotionRedemptions
source: router/api-router.go:276
request:
  query_params:
    - p
    - page_size
    - campaign_id
    - code_id
    - user_id
    - keyword
    - status
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/topup-promotion/redemptions`

分页查看充值促销兑换记录。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。
- `campaign_id`: 可选活动 ID 过滤。
- `code_id`: 可选促销码 ID 过滤。
- `user_id`: 可选用户 ID 过滤。
- `keyword`: 可选关键字，可匹配兑换 ID、用户 ID、充值 ID、促销码或交易号。
- `status`: 可选兑换状态。常见值为 `reserved`、`used`、`expired`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 总数。
- `data.items`: 兑换记录数组。
- `data.items[].id`: 兑换记录 ID。
- `data.items[].campaign_id`: 活动 ID。
- `data.items[].campaign_name`: 活动名称。
- `data.items[].code_id`: 促销码 ID。
- `data.items[].code`: 促销码文本。
- `data.items[].user_id`: 用户 ID。
- `data.items[].username`: 用户名。
- `data.items[].top_up_id`: 关联充值订单 ID。
- `data.items[].trade_no`: 充值交易号。
- `data.items[].payment_method`: 支付方式。
- `data.items[].amount`: 原始充值额度或金额输入值。
- `data.items[].original_amount`: 原始应付金额。
- `data.items[].discount_amount`: 促销抵扣金额。
- `data.items[].final_payable_amount`: 抵扣后的最终应付金额。
- `data.items[].currency_code`: 币种。
- `data.items[].status`: 兑换状态。
- `data.items[].reserved_at`: 预约时间 Unix 秒。
- `data.items[].used_at`: 使用时间 Unix 秒。
- `data.items[].expired_at`: 过期时间 Unix 秒。
- `data.items[].created_time`: 创建时间。
- `data.items[].updated_time`: 更新时间。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。

