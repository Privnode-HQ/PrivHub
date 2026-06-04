---
method: PUT
path: /api/beta/self_quota_take_away
auth: user
handler: controller.SelfQuotaTakeAway
source: router/api-router.go:378
request:
  body:
    - money
response:
  success_http_status: 200
  envelope: common
---

# PUT `/api/beta/self_quota_take_away`

当前用户按金额折算额度并从自己的剩余额度中扣减。该接口不会增加 `used_quota`。

## 请求体字段

- `money`: 要扣减的金额，必须大于 `0`。后端按 `money * QuotaPerUnit` 取整数额度。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.quota_taken`: 实际扣减额度。
- `data.quota_after`: 扣减后的剩余额度。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、`money 必须大于 0`、`扣减额度过小`、`扣减额度过大`、`额度不足`、查询用户额度失败或扣减失败。

