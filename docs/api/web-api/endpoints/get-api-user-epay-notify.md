---
method: GET
path: /api/user/epay/notify
auth: public_gateway_callback
handler: controller.EpayNotify
source: router/api-router.go:73
request:
  query_params: gateway_defined
response:
  success_http_status: 200
  envelope: plain_text
---

# GET `/api/user/epay/notify`

Epay 支付网关异步回调。该接口由支付平台调用，不面向普通前端调用。

## 查询参数字段

查询参数由 Epay 网关定义，控制器会把所有 URL query 转成字符串 map 并交给 Epay SDK 验签。关键字段通常包括：

- `trade_no`: 网关交易号。
- `out_trade_no` 或服务端交易号字段：本地充值单号，控制器使用 SDK 解析出的 `ServiceTradeNo`。
- `trade_status`: 交易状态；成功时 SDK 解析为 `TRADE_SUCCESS`。
- `sign`: 网关签名。
- `sign_type`: 签名算法。

## 成功响应字段

该接口返回纯文本：

- `success`: 验签通过时立即返回该文本。

## 失败响应

该接口返回纯文本：

- `fail`: 未配置 Epay 客户端、验签失败或写入失败时返回。

## 业务规则

- 只有 Epay SDK 验签通过且交易状态成功时才会处理充值。
- 本地订单必须是 `pending` 才会增加用户额度；重复回调不会重复充值。
- 成功处理后会把订单状态改为 `success`，核销优惠券/促销码，并写入充值日志。

