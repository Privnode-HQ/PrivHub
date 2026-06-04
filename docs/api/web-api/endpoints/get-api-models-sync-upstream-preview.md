---
method: GET
path: /api/models/sync_upstream/preview
auth: support
handler: controller.SyncUpstreamPreview
source: router/api-router.go:351
request:
  query_params:
    - locale
response:
  success_http_status: 200
  envelope: raw-json
---

# GET `/api/models/sync_upstream/preview`

预览官方上游模型/供应商元数据与本地数据之间的缺失项和冲突项。

## 查询参数字段

- `locale`: 可选语言。支持 `en`、`zh`、`ja`；其他值使用默认上游路径。

## 成功响应字段

- `success`: `true`。
- `data.missing`: 本地渠道引用但本地模型元数据缺失、且上游存在的模型名数组。
- `data.conflicts`: 本地已有模型与上游存在差异的数组。
- `data.conflicts[].model_name`: 模型名。
- `data.conflicts[].fields`: 冲突字段数组。
- `data.conflicts[].fields[].field`: 字段名，可能为 `description`、`icon`、`tags`、`vendor`、`name_rule`、`status`。
- `data.conflicts[].fields[].local`: 本地值。
- `data.conflicts[].fields[].upstream`: 上游值。
- `data.source.locale`: 本次预览使用的 locale。
- `data.source.models_url`: 上游 models JSON URL。
- `data.source.vendors_url`: 上游 vendors JSON URL。

## 失败响应

- `success`: `false`。
- `message`: `获取上游模型失败: ...`。
- `locale`: 请求的 locale。
- `source_urls.models_url`: models 上游 URL。
- `source_urls.vendors_url`: vendors 上游 URL。

