---
method: GET
path: /api/models/search
auth: support
handler: controller.SearchModelsMeta
source: router/api-router.go:354
request:
  query_params:
    - keyword
    - vendor
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/models/search`

搜索模型元数据。

## 查询参数字段

- `keyword`: 搜索关键字，匹配 `model_name`、`description` 或 `tags`。
- `vendor`: 供应商过滤。整数时按 `vendor_id`；非整数时按供应商名称模糊匹配。
- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 命中总数。
- `data.items`: 模型元数据数组。
- `data.items[].id`: 模型元数据 ID。
- `data.items[].model_name`: 模型名或规则片段。
- `data.items[].description`: 说明。
- `data.items[].icon`: 图标名。
- `data.items[].tags`: 标签。
- `data.items[].vendor_id`: 供应商 ID。
- `data.items[].endpoints`: 支持端点类型 JSON 字符串。
- `data.items[].status`: 状态。
- `data.items[].sync_official`: 官方同步开关。
- `data.items[].created_time`: 创建时间。
- `data.items[].updated_time`: 更新时间。
- `data.items[].bound_channels`: 绑定渠道数组。
- `data.items[].enable_groups`: 启用分组数组。
- `data.items[].quota_types`: 配额类型数组。
- `data.items[].name_rule`: `0` 精确，`1` 前缀，`2` 包含，`3` 后缀。
- `data.items[].matched_models`: 匹配模型数组。
- `data.items[].matched_count`: 匹配数量。

## 失败响应

- `success`: `false`。
- `message`: 搜索失败原因。

