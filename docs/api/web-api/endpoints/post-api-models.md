---
method: POST
path: /api/models/
auth: admin
handler: controller.CreateModelMeta
source: router/api-router.go:361
request:
  body: model.Model
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/models/`

管理员创建模型元数据。

## 请求体字段

- `model_name`: 模型名或规则片段，必填，软删除范围内唯一。
- `description`: 模型说明。
- `icon`: 图标名。
- `tags`: 标签字符串。
- `vendor_id`: 供应商 ID。
- `endpoints`: 支持端点类型 JSON 字符串。
- `status`: 状态，通常 `1` 启用。
- `sync_official`: 官方同步开关，`0` 表示不同步。
- `name_rule`: 名称规则，`0` 精确，`1` 前缀，`2` 包含，`3` 后缀。
- `id`: 创建时忽略。
- `created_time`: 创建时由后端设置。
- `updated_time`: 创建时由后端设置。
- `bound_channels`: 视图字段，创建时忽略。
- `enable_groups`: 视图字段，创建时忽略。
- `quota_types`: 视图字段，创建时忽略。
- `matched_models`: 视图字段，创建时忽略。
- `matched_count`: 视图字段，创建时忽略。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 新模型元数据 ID。
- `data.model_name`: 模型名。
- `data.description`: 说明。
- `data.icon`: 图标。
- `data.tags`: 标签。
- `data.vendor_id`: 供应商 ID。
- `data.endpoints`: 端点 JSON。
- `data.status`: 状态。
- `data.sync_official`: 同步开关。
- `data.created_time`: 创建时间。
- `data.updated_time`: 更新时间。
- `data.name_rule`: 名称规则。

## 失败响应

- `success`: `false`。
- `message`: JSON 绑定错误、`模型名称不能为空`、`模型名称已存在` 或数据库错误。

