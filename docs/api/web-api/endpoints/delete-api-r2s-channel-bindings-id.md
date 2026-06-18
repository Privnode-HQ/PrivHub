---
method: DELETE
path: /api/r2s/channel-bindings/:id
auth: admin
handler: controller.DisableR2SChannelBinding
source: router/api-router.go:302
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# DELETE `/api/r2s/channel-bindings/:id`

管理员停用 R2S 渠道绑定。该接口不会物理删除绑定，避免破坏历史收入
识别记录的审计链路。

## 路径参数字段

- `id`: 渠道绑定 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 状态已变为 `disabled` 的渠道绑定对象。

## 失败响应

- `success`: `false`。
- `message`: ID 解析错误、绑定不存在或数据库错误。
