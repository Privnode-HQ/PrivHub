---
method: POST
path: /api/r2s/recognition-records/sync
auth: admin
handler: controller.SyncR2SRecognitionRecords
source: router/api-router.go:311
request:
  body:
    content_type: application/json
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/r2s/recognition-records/sync`

管理员从历史消费日志同步 R2S 收入识别记录，用于补识别 R2S 功能启用前已
存在的消费利润。该接口只处理有启用中 R2S 渠道绑定的消费日志；没有绑定的
日志会跳过。

同步使用消费日志的 `quota` 作为系统货币收入，按当前启用中的 R2S 渠道绑定
倍率计算供应商成本，并写入 `usage` 来源识别记录。来源引用固定为
`usage_log:<log id>`，因此重复执行会更新原记录，不会重复创建。

## 请求体字段

- `start_time`: 可选。Unix 秒，只同步该时间及之后的消费日志。
- `end_time`: 可选。Unix 秒，只同步该时间及之前的消费日志，不能早于
  `start_time`。

请求体可为空。为空时同步所有可用历史消费日志。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.started_at`: 同步开始时间 Unix 秒。
- `data.finished_at`: 同步结束时间 Unix 秒。
- `data.created_count`: 本次新增的识别记录数。
- `data.updated_count`: 本次更新的识别记录数。
- `data.skipped_count`: 因没有启用中 R2S 渠道绑定而跳过的消费日志数。
- `data.synced_count`: 本次新增或更新的识别记录总数。

## 失败响应

- `success`: `false`。
- `message`: 参数错误、结束时间早于开始时间、QuotaPerUnit 配置不正确、
  供应商或渠道绑定快照失败，或数据库错误。
