---
method: GET
path: /api/r2s/suppliers/:id
auth: admin
handler: controller.GetR2SSupplier
source: router/api-router.go:296
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/r2s/suppliers/:id`

管理员查看单个 R2S 供应商。

## 路径参数字段

- `id`: R2S 供应商 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: R2S 供应商对象，字段同 GET `/api/r2s/suppliers`
  的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: ID 解析错误、供应商不存在或数据库错误。
