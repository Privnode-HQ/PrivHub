---
method: GET
path: /api/user/groups
auth: public
handler: controller.GetUserGroups
source: router/api-router.go:74
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/user/groups`

获取匿名视角可用分组及倍率、描述和训练数据采集提示。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 对象。键为分组名。
- `data.<group>.ratio`: 当前用户分组到目标分组的倍率；匿名时按空用户分组解析。`auto` 分组为字符串 `自动`。
- `data.<group>.desc`: 分组说明。
- `data.<group>.capture_rate`: 训练数据采集率数值。
- `data.<group>.requires_training_data_consent`: 布尔值，`capture_rate > 0` 时为 `true`。
- `data.<group>.training_data_allowed`: 布尔值。匿名视角没有用户设置时通常为 `false`。

## 失败响应

当前控制器没有显式错误分支。

