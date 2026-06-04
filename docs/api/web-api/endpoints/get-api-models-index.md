---
method: GET
path: /api/models/
auth: support
handler: controller.GetAllModelsMeta
source: router/api-router.go:353
request:
  query_params:
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/models/`

支持人员或管理员分页查看模型元数据。注意该接口不同于用户侧 `GET /api/models`。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.items`: 模型元数据数组。
- `data.items[].id`: 模型元数据 ID。
- `data.items[].model_name`: 模型名或规则匹配片段。
- `data.items[].description`: 模型说明。
- `data.items[].icon`: 图标名。
- `data.items[].tags`: 标签字符串。
- `data.items[].vendor_id`: 供应商 ID。
- `data.items[].endpoints`: 支持端点类型 JSON 字符串。
- `data.items[].status`: 状态，通常 `1` 启用。
- `data.items[].sync_official`: 是否参与官方同步，`0` 表示不同步。
- `data.items[].created_time`: 创建时间 Unix 秒。
- `data.items[].updated_time`: 更新时间 Unix 秒。
- `data.items[].bound_channels`: 绑定渠道数组。
- `data.items[].bound_channels[].name`: 渠道名称。
- `data.items[].bound_channels[].type`: 渠道类型。
- `data.items[].enable_groups`: 启用分组数组。
- `data.items[].quota_types`: 配额类型数组。
- `data.items[].name_rule`: 名称规则，`0` 精确，`1` 前缀，`2` 包含，`3` 后缀。
- `data.items[].matched_models`: 规则模型匹配到的实际模型名数组。
- `data.items[].matched_count`: 规则模型匹配数量。
- `data.total`: 总数。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.vendor_counts`: 按供应商 ID 统计的模型数量。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。

