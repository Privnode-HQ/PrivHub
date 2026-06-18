---
method: POST
path: /api/r2s/suppliers/:id/enable
auth: admin
handler: controller.EnableR2SSupplier
source: router/api-router.go:298
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/r2s/suppliers/:id/enable`

管理员重新启用已停用的 R2S 供应商。启用后，该供应商会重新进入
供应商余额统计和余额提醒统计。

## 路径参数字段

- `id`: R2S 供应商 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 状态已变为 `active` 的 R2S 供应商对象，字段同
  GET `/api/r2s/suppliers` 的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: ID 解析错误、供应商不存在、字段校验失败或数据库错误。
