---
method: GET
path: /api/mj/
auth: support
handler: controller.GetAllMidjourney
source: router/api-router.go:325
request:
  query_params:
    - channel_id
    - mj_id
    - start_timestamp
    - end_timestamp
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/mj/`

支持人员或管理员获取全站 Midjourney 任务列表。

## 查询参数字段

- `channel_id`: 字符串，可选。渠道 ID。
- `mj_id`: 字符串，可选。Midjourney 任务 ID。
- `start_timestamp`: 字符串，可选。提交时间下限。
- `end_timestamp`: 字符串，可选。提交时间上限。
- `p`: 页码。
- `page_size`: 每页条数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 分页对象，`items[]` 字段同 `GET /api/mj/self`。

## 失败响应

- `success`: `false`。
- `message`: 查询错误。

