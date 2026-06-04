---
method: GET
path: /api/mj/self
auth: user
handler: controller.GetUserMidjourney
source: router/api-router.go:324
request:
  query_params:
    - mj_id
    - start_timestamp
    - end_timestamp
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/mj/self`

获取当前用户 Midjourney 任务列表。

## 查询参数字段

- `mj_id`: 字符串，可选。Midjourney 任务 ID。
- `start_timestamp`: 字符串，可选。提交时间下限。
- `end_timestamp`: 字符串，可选。提交时间上限。
- `p`: 页码。
- `page_size`: 每页条数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 分页对象。
- `data.items[].id`: 本地任务 ID。
- `data.items[].code`: 上游状态码。
- `data.items[].user_id`: 用户 ID。
- `data.items[].action`: 任务动作。
- `data.items[].mj_id`: Midjourney 任务 ID。
- `data.items[].prompt`: 原始 prompt。
- `data.items[].prompt_en`: 英文 prompt。
- `data.items[].description`: 描述。
- `data.items[].state`: 状态说明。
- `data.items[].submit_time`: 提交时间 Unix 秒。
- `data.items[].start_time`: 开始时间 Unix 秒。
- `data.items[].finish_time`: 完成时间 Unix 秒。
- `data.items[].image_url`: 图片 URL；启用转发时会被改写为本服务地址。
- `data.items[].video_url`: 视频 URL。
- `data.items[].video_urls`: 视频 URL JSON 字符串。
- `data.items[].status`: 状态，例如 `SUCCESS`、`FAILURE`、`IN_PROGRESS`。
- `data.items[].progress`: 进度文本。
- `data.items[].fail_reason`: 失败原因。
- `data.items[].channel_id`: 渠道 ID。
- `data.items[].quota`: 消耗额度。
- `data.items[].buttons`: 按钮 JSON 字符串。
- `data.items[].properties`: 属性 JSON 字符串。

## 失败响应

- `success`: `false`。
- `message`: 查询错误。

