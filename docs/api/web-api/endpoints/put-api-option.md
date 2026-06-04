---
method: PUT
path: /api/option/
auth: root
handler: controller.UpdateOption
source: router/api-router.go:150
request:
  content_type: application/json
response:
  success_http_status: 200
  bad_request_http_status: 400
  envelope: custom_success
---

# PUT `/api/option/`

Root 更新一个全局选项。

## 请求体字段

- `key`: 字符串，必填。选项键名。
- `value`: 任意 JSON 值，必填。布尔、数字和字符串都会转为字符串保存。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- HTTP 400: `success=false`，`message=无效的参数`。
- HTTP 200: `success=false`，可能为配置校验错误，例如额度展示货币仅支持 `USD` 或 `CNY`、启用 OAuth/邮箱/Turnstile 前缺少必要配置、Postmark 大批量模式无效、分组倍率/采集率 JSON 无效、音频/图片倍率配置无效、用量限制规则无效、控制台配置无效。

