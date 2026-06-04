---
method: GET
path: /api/beta/remain_actual_paid_amount
auth: user
handler: controller.GetRemainActualPaidAmount
source: router/api-router.go:377
request:
  query_params: []
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/beta/remain_actual_paid_amount`

获取当前用户剩余余额中按比例归属于实际付费额度的部分。

## 请求字段

无请求体，无查询参数。

## 计算规则

- `R`: 当前剩余额度。
- `T`: 当前剩余额度加已用额度。
- `P`: 用户历史充值得到的总付费额度。
- 返回值为 `floor(P * R / T)`。
- 当 `R <= 0`、`T <= 0` 或 `P <= 0` 时返回 `0`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.remain_actual_paid_amount`: 当前剩余额度中估算的付费额度部分。

## 失败响应

- `success`: `false`。
- `message`: 查询用户额度、已用额度或付费额度失败原因。

