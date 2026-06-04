---
method: GET
path: /api/models/:id
auth: support
handler: controller.GetModelMeta
source: router/api-router.go:355
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/models/:id`

按 ID 查看模型元数据。

## 路径参数字段

- `id`: 模型元数据 ID，必须是整数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.id`: 模型元数据 ID。
- `data.model_name`: 模型名或规则片段。
- `data.description`: 模型说明。
- `data.icon`: 图标名。
- `data.tags`: 标签字符串。
- `data.vendor_id`: 供应商 ID。
- `data.endpoints`: 支持端点类型 JSON 字符串。
- `data.status`: 状态，通常 `1` 启用。
- `data.sync_official`: 官方同步开关。
- `data.created_time`: 创建时间 Unix 秒。
- `data.updated_time`: 更新时间 Unix 秒。
- `data.bound_channels`: 绑定渠道数组。
- `data.bound_channels[].name`: 渠道名称。
- `data.bound_channels[].type`: 渠道类型。
- `data.enable_groups`: 启用分组数组。
- `data.quota_types`: 配额类型数组。
- `data.name_rule`: `0` 精确，`1` 前缀，`2` 包含，`3` 后缀。
- `data.matched_models`: 规则模型匹配到的实际模型名数组。
- `data.matched_count`: 匹配数量。

## 失败响应

- `success`: `false`。
- `message`: `id` 非整数、模型不存在或数据库错误。

