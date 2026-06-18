---
method: DELETE
path: /api/r2s/suppliers/:id/permanent
auth: admin
handler: controller.DeleteR2SSupplier
source: router/api-router.go:299
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# DELETE `/api/r2s/suppliers/:id/permanent`

管理员永久删除未产生财务历史的 R2S 供应商。该接口只适用于误建或
尚未使用的供应商；如果供应商已有付款历史、余额更新历史或收入识别记录，
接口会拒绝删除，管理员应改用停用来保留审计链路。

删除成功时，该供应商关联的 R2S 渠道绑定会一并删除，避免留下无法使用的
本地配置。

## 路径参数字段

- `id`: R2S 供应商 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: `null`。

## 失败响应

- `success`: `false`。
- `message`: ID 解析错误、供应商不存在、供应商已有付款历史、供应商已有
  余额更新历史、供应商已有收入识别记录，或数据库错误。
