---
method: GET
path: /api/admin/users/:cah_id/remain_actual_paid_amount
auth: admin
handler: controller.GetAdminUserRemainActualPaidAmountByCAHID
source: router/api-router.go:23
request:
  path_params:
    - cah_id
  query_params:
    - rate
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/admin/users/:cah_id/remain_actual_paid_amount`

管理员按 CAH ID 查询用户剩余付费额度估算，并可按比例计算调整值。

## 路径参数字段

- `cah_id`: 字符串，必填。目标用户 CAH ID。

## 查询参数字段

- `rate`: 字符串，可选。`0` 到 `1` 之间的小数；传入时额外返回按比例调整后的剩余付费额度。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.cah_id`: 用户 CAH ID。
- `data.user_id`: 用户数字 ID。
- `data.username`: 用户名。
- `data.quota_per_unit`: 当前全局金额到额度换算单位。
- `data.user_total_quota`: 用户总额度，等于剩余额度加已用额度。
- `data.user_remain_quota`: 用户当前剩余额度。
- `data.user_remain_paid_quota`: 从成功充值记录推导的剩余付费额度估算。
- `data.user_remain_non_paid_quota`: 剩余额度中非付费部分估算。
- `data.user_total_actual_paid_quota`: 用户历史成功充值推导的总付费额度。
- `data.rate`: 仅传入合法 `rate` 时返回，字符串形式的小数。
- `data.user_remain_paid_quota_adjusted`: 仅传入合法 `rate` 时返回，`user_remain_paid_quota * rate` 的整数部分。

## 失败响应

- `success`: `false`。
- `message`: 用户不存在、查询失败，或 `rate 必须是 0 到 1 之间的小数`。

