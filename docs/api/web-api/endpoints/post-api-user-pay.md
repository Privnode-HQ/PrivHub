---
method: POST
path: /api/user/pay
auth: user_rate_limited
handler: controller.RequestEpay
source: router/api-router.go:104
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: epay_legacy
---

# POST `/api/user/pay`

创建 Epay 在线充值订单并返回支付参数。

## 请求体字段

- `amount`: 整数，必填。充值额度数量，不能小于系统最小充值。
- `payment_method`: 字符串，必填。Epay 支付类型，必须存在于系统支付方式配置。
- `top_up_code`: 字符串。历史字段，当前控制器不直接使用。
- `coupon_id`: 整数。要使用的优惠券 ID。
- `promotion_code`: 字符串。要使用的促销码；与优惠券不能同时使用。

## 成功响应字段

该接口使用历史响应形状：

- `message`: `success`。
- `data`: Epay SDK 生成的支付参数对象。
- `url`: Epay 收银台 URL。

## 失败响应

- `message`: `error`，或部分分支直接使用错误文本。
- `data`: 错误说明，例如 `参数错误`、`充值数量不能小于 ...`、`获取用户信息失败`、`支付方式不存在`、`当前管理员未配置支付信息`、`拉起支付失败`、`创建订单失败`。

## 业务规则

- 订单创建时会保存 `pending` 状态，并保留折扣、优惠券、促销码快照。
- 支付回调成功后才会把订单改为 `success` 并增加用户额度。

