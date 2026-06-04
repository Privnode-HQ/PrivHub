---
method: POST
path: /api/user/aff_transfer
auth: user
handler: controller.TransferAffQuota
source: router/api-router.go:109
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/aff_transfer`

把当前用户的邀请码奖励额度划转到普通可用额度。

## 请求体字段

- `quota`: 整数，必填。要划转的奖励额度数量。

## 成功响应字段

- `success`: `true`。
- `message`: `划转成功`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `划转失败 ...`，或参数/用户查询错误。

