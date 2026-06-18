---
method: DELETE
path: /api/r2s/suppliers/:id
auth: admin
handler: controller.DisableR2SSupplier
source: router/api-router.go:298
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# DELETE `/api/r2s/suppliers/:id`

管理员停用 R2S 供应商。该接口不会物理删除历史付款、余额更新或收入识别
记录，以保证历史利润识别可以继续审计。

## 路径参数字段

- `id`: R2S 供应商 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 状态已变为 `disabled` 的 R2S 供应商对象。

## 失败响应

- `success`: `false`。
- `message`: ID 解析错误、供应商不存在或数据库错误。
