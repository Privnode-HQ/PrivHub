---
method: GET
path: /api/task/
auth: support
handler: controller.GetAllTask
source: router/api-router.go:330
request:
  query_params:
    - platform
    - task_id
    - status
    - action
    - start_timestamp
    - end_timestamp
    - channel_id
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/task/`

支持人员或管理员获取全站异步任务列表。

## 查询参数字段

- `platform`: 字符串，可选。任务平台。
- `task_id`: 字符串，可选。上游任务 ID。
- `status`: 字符串，可选。任务状态。
- `action`: 字符串，可选。任务动作。
- `start_timestamp`: 整数，可选。提交时间下限。
- `end_timestamp`: 整数，可选。提交时间上限。
- `channel_id`: 字符串，可选。渠道 ID。
- `p`: 页码。
- `page_size`: 每页条数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 分页对象，`items[]` 字段同 `GET /api/task/self`。

## 失败响应

- `success`: `false`。
- `message`: 查询错误。

