---
method: POST
path: /api/models/sync_upstream
auth: admin
handler: controller.SyncUpstreamModels
source: router/api-router.go:360
request:
  body: syncRequest
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/models/sync_upstream`

从官方上游同步模型和供应商元数据。默认只为本地缺失模型创建记录，可通过 `overwrite` 覆盖已有模型的指定字段。

## 请求体字段

- `locale`: 可选语言。支持 `en`、`zh`、`ja`；其他值使用默认上游路径。
- `overwrite`: 可选覆盖列表。
- `overwrite[].model_name`: 要覆盖的本地模型名。
- `overwrite[].fields`: 要覆盖的字段数组。可用值为 `description`、`icon`、`tags`、`vendor`、`name_rule`、`status`。

## 成功响应字段

- `success`: `true`。
- `data.created_models`: 新建模型数量。
- `data.created_vendors`: 新建供应商数量。
- `data.updated_models`: 覆盖更新模型数量。
- `data.skipped_models`: 跳过的模型名数组。
- `data.created_list`: 成功创建的模型名数组。
- `data.updated_list`: 成功更新的模型名数组。
- `data.source.locale`: 使用的 locale。
- `data.source.models_url`: models 上游 URL。
- `data.source.vendors_url`: vendors 上游 URL。

## 失败响应

- `success`: `false`。
- `message`: 查询缺失模型失败、`获取上游模型失败: ...` 或上游响应解析错误。
- `locale`: 上游获取失败时返回请求 locale。
- `source_urls.models_url`: 上游 models URL。
- `source_urls.vendors_url`: 上游 vendors URL。

