---
method: GET
path: /api/topup-promotion/codes
auth: support
handler: controller.GetTopUpPromotionCodes
source: router/api-router.go:275
request:
  query_params:
    - p
    - page_size
    - campaign_id
    - keyword
    - status
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/topup-promotion/codes`

分页查看充值促销码。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。
- `campaign_id`: 可选活动 ID，只查看该活动下的促销码。
- `keyword`: 可选关键字，可匹配促销码、促销码 ID 或活动 ID。
- `status`: 可选存储状态。常见值为 `active`、`paused`、`revoked`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 总数。
- `data.items`: 促销码数组。
- `data.items[].id`: 促销码 ID。
- `data.items[].campaign_id`: 所属活动 ID。
- `data.items[].campaign_name`: 所属活动名称。
- `data.items[].code`: 促销码文本。
- `data.items[].status`: 存储状态。
- `data.items[].effective_status`: 按时间窗口计算后的实际状态。
- `data.items[].valid_from`: 生效时间 Unix 秒。
- `data.items[].expires_at`: 过期时间 Unix 秒。
- `data.items[].max_redemptions`: 总兑换限制。
- `data.items[].max_redemptions_per_user`: 每用户兑换限制。
- `data.items[].created_by_admin_id`: 创建管理员 ID。
- `data.items[].created_time`: 创建时间。
- `data.items[].updated_time`: 更新时间。
- `data.items[].revoked_at`: 撤销时间。
- `data.items[].revoked_by_admin_id`: 撤销管理员 ID。
- `data.items[].revoke_reason`: 撤销原因。
- `data.items[].reserved_count`: 预约中次数。
- `data.items[].used_count`: 已使用次数。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。

