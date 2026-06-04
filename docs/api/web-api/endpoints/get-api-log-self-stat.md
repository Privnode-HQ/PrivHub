---
method: GET
path: /api/log/self/stat
auth: user
handler: controller.GetLogsSelfStat
source: router/api-router.go:291
request:
  query_params:
    - type
    - start_timestamp
    - end_timestamp
    - token_name
    - model_name
    - channel
    - group
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/log/self/stat`

统计当前用户的日志消耗额度、最近 60 秒 RPM 和 TPM。

## 查询参数字段

- `type`: 日志类型。
- `start_timestamp`: 起始 Unix 秒。
- `end_timestamp`: 结束 Unix 秒。
- `token_name`: Token 名称过滤。
- `model_name`: 模型名过滤，使用 SQL `LIKE`。
- `channel`: 渠道 ID。
- `group`: 分组过滤。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.quota`: 当前用户在查询范围内的消耗额度总和。
- `data.rpm`: 当前用户最近 60 秒请求数。
- `data.tpm`: 当前用户最近 60 秒输入和输出 token 总和。

## 失败响应

该控制器没有显式业务失败分支；认证失败由中间件处理。

