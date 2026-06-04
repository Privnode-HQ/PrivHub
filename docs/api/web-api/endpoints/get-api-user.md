---
method: GET
path: /api/user/
auth: support
handler: controller.GetAllUsers
source: router/api-router.go:123
request:
  query_params:
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/user/`

支持人员或管理员分页查看用户列表。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，最大 `100`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页码。
- `data.page_size`: 实际每页条数。
- `data.total`: 总用户数。
- `data.items`: 用户数组。
- `data.items[].id`: 用户 ID。
- `data.items[].cah_id`: CAH ID。
- `data.items[].username`: 用户名。
- `data.items[].display_name`: 显示名。
- `data.items[].role`: 角色，可能值 `0`、`1`、`5`、`10`、`100`。
- `data.items[].status`: 状态，`1` 启用，`2` 禁用。
- `data.items[].ban_reason`: 禁用原因。
- `data.items[].email`: 邮箱。
- `data.items[].force_password_reset`: 是否强制重置密码。
- `data.items[].force_email_bind`: 是否强制绑定邮箱。
- `data.items[].quota`: 剩余额度。
- `data.items[].used_quota`: 已用额度。
- `data.items[].request_count`: 请求次数。
- `data.items[].group`: 分组。
- `data.items[].aff_code`: 邀请码。
- `data.items[].aff_count`: 邀请人数。
- `data.items[].aff_quota`: 邀请奖励剩余额度。
- `data.items[].aff_history_quota`: 邀请奖励历史额度。
- `data.items[].inviter_id`: 邀请人 ID。
- `data.items[].remark`: 管理备注。
- `data.items[].stripe_customer`: Stripe Customer ID。

## 失败响应

- `success`: `false`。
- `message`: 用户查询错误。

