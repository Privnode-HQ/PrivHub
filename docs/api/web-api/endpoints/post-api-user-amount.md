---
method: POST
path: /api/user/amount
auth: user
handler: controller.RequestAmount
source: router/api-router.go:105
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: amount_legacy
---

# POST `/api/user/amount`

计算 Epay 默认充值路径下指定额度数量的应付金额，不创建订单。

## 请求体字段

- `amount`: 整数，必填。充值额度数量，不能小于系统最小充值。
- `top_up_code`: 字符串。历史字段，当前控制器不直接使用。
- `coupon_id`: 整数。历史字段，本接口当前不参与优惠券计算。

## 成功响应字段

- `message`: `success`。
- `data`: 字符串。保留两位小数的应付金额。

## 失败响应

- `message`: `error`。
- `data`: 错误说明，例如 `参数错误`、`充值数量不能小于 ...`、`获取用户分组失败`、`充值金额过低`。

