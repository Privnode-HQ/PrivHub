---
method: PUT
path: /api/models/
auth: admin
handler: controller.UpdateModelMeta
source: router/api-router.go:362
request:
  query_params:
    - status_only
  body: model.Model
response:
  success_http_status: 200
  envelope: common
---

# PUT `/api/models/`

管理员更新模型元数据，或仅更新模型状态。

## 查询参数字段

- `status_only`: 传 `true` 时只更新 `status` 字段，避免清空其他字段。

## 请求体字段

- `id`: 模型元数据 ID，必填。
- `status`: 新状态；`status_only=true` 时只使用该字段。
- `model_name`: 模型名或规则片段；非 `status_only` 时会检查重复。
- `description`: 说明。
- `icon`: 图标名。
- `tags`: 标签。
- `vendor_id`: 供应商 ID。
- `endpoints`: 支持端点类型 JSON 字符串。
- `sync_official`: 官方同步开关。
- `name_rule`: 名称规则，`0` 精确，`1` 前缀，`2` 包含，`3` 后缀。
- `created_time`: 更新时被模型层排除，不覆盖创建时间。
- `updated_time`: 后端更新为当前时间。
- `bound_channels`: 视图字段，更新时不应依赖。
- `enable_groups`: 视图字段，更新时不应依赖。
- `quota_types`: 视图字段，更新时不应依赖。
- `matched_models`: 视图字段，更新时不应依赖。
- `matched_count`: 视图字段，更新时不应依赖。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 模型元数据 ID。
- `data.model_name`: 模型名。
- `data.description`: 说明。
- `data.icon`: 图标。
- `data.tags`: 标签。
- `data.vendor_id`: 供应商 ID。
- `data.endpoints`: 端点 JSON。
- `data.status`: 状态。
- `data.sync_official`: 官方同步开关。
- `data.updated_time`: 更新时间。
- `data.name_rule`: 名称规则。

## 失败响应

- `success`: `false`。
- `message`: JSON 绑定错误、`缺少模型 ID`、`模型名称已存在` 或数据库错误。

