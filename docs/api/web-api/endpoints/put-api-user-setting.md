---
method: PUT
path: /api/user/setting
auth: user
handler: controller.UpdateUserSetting
source: router/api-router.go:110
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# PUT `/api/user/setting`

更新当前用户通知、模型倍率接受策略、IP 日志和训练数据采集同意设置。

## 请求体字段

- `notify_type`: 字符串。当前仅支持邮件通知；非空且不是邮件类型会失败。成功保存时后端固定写入邮件类型。
- `quota_warning_threshold`: 数字，必填。额度预警阈值，必须大于 `0`。
- `webhook_url`: 字符串，可选。历史字段，当前控制器不保存。
- `webhook_secret`: 字符串，可选。历史字段，当前控制器不保存。
- `notification_email`: 字符串，可选。接收通知的邮箱；非空时必须包含 `@`。
- `bark_url`: 字符串，可选。历史字段，当前控制器不保存。
- `gotify_url`: 字符串，可选。历史字段，当前控制器不保存。
- `gotify_token`: 字符串，可选。历史字段，当前控制器不保存。
- `gotify_priority`: 整数，可选。历史字段，当前控制器不保存。
- `accept_unset_model_ratio_model`: 布尔值。是否接受未设置模型倍率的模型。
- `record_ip_log`: 布尔值。是否记录 IP 日志。
- `allow_training_data_groups`: 布尔值。是否允许使用需要训练数据采集同意的分组。

## 成功响应字段

- `success`: `true`。
- `message`: `设置已更新`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`当前仅支持邮件通知`、`预警阈值必须大于0`、`无效的邮箱地址`、`更新设置失败: ...`，或用户查询错误。

