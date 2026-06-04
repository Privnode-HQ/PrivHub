---
method: GET
path: /api/log/stat
auth: support
handler: controller.GetLogsStat
source: router/api-router.go:290
request:
  query_params:
    - type
    - start_timestamp
    - end_timestamp
    - username
    - token_name
    - model_name
    - channel
    - group
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/log/stat`

统计日志消耗额度、最近 60 秒 RPM 和 TPM。`type=3` 管理日志固定返回 0。

## 查询参数字段

- `type`: 日志类型；统计内部只累加消耗日志，`3` 时直接返回零值。
- `start_timestamp`: 起始 Unix 秒。
- `end_timestamp`: 结束 Unix 秒。
- `username`: 用户名过滤。
- `token_name`: Token 名称过滤。
- `model_name`: 模型名过滤，使用 SQL `LIKE`。
- `channel`: 渠道 ID。
- `group`: 分组过滤。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.quota`: 查询范围内消耗额度总和。
- `data.rpm`: 最近 60 秒命中的请求数量。
- `data.tpm`: 最近 60 秒输入和输出 token 总和。

## 失败响应

该控制器没有显式业务失败分支；认证失败由中间件处理。

