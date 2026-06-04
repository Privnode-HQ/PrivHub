---
method: GET
path: /api/task/self
auth: user
handler: controller.GetUserTask
source: router/api-router.go:329
request:
  query_params:
    - platform
    - task_id
    - status
    - action
    - start_timestamp
    - end_timestamp
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/task/self`

获取当前用户的异步任务列表。

## 查询参数字段

- `platform`: 字符串，可选。任务平台。
- `task_id`: 字符串，可选。上游任务 ID。
- `status`: 字符串，可选。任务状态，可能值 `NOT_START`、`SUBMITTED`、`QUEUED`、`IN_PROGRESS`、`FAILURE`、`SUCCESS`、`UNKNOWN`。
- `action`: 字符串，可选。任务动作。
- `start_timestamp`: 整数，可选。提交时间下限。
- `end_timestamp`: 整数，可选。提交时间上限。
- `p`: 页码。
- `page_size`: 每页条数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 分页对象。
- `data.items[].id`: 本地任务 ID。
- `data.items[].created_at`: 创建时间 Unix 秒。
- `data.items[].updated_at`: 更新时间 Unix 秒。
- `data.items[].task_id`: 上游任务 ID。
- `data.items[].platform`: 平台。
- `data.items[].user_id`: 用户 ID。
- `data.items[].group`: 计费分组。
- `data.items[].channel_id`: 渠道 ID。
- `data.items[].quota`: 消耗额度。
- `data.items[].action`: 任务动作。
- `data.items[].status`: 任务状态。
- `data.items[].fail_reason`: 失败原因。
- `data.items[].submit_time`: 提交时间。
- `data.items[].start_time`: 开始时间。
- `data.items[].finish_time`: 完成时间。
- `data.items[].progress`: 进度。
- `data.items[].properties`: 属性对象。
- `data.items[].properties.input`: 输入文本。
- `data.items[].properties.upstream_model_name`: 上游模型名。
- `data.items[].properties.origin_model_name`: 原始模型名。
- `data.items[].data`: 任务结果 JSON。

## 失败响应

- `success`: `false`。
- `message`: 查询错误。

