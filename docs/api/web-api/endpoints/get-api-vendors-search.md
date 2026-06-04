---
method: GET
path: /api/vendors/search
auth: support
handler: controller.SearchVendors
source: router/api-router.go:337
request:
  query_params:
    - keyword
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/vendors/search`

按关键字搜索模型供应商。

## 查询参数字段

- `keyword`: 搜索关键字，匹配 `name` 或 `description`。
- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 命中总数。
- `data.items`: 供应商数组。
- `data.items[].id`: 供应商 ID。
- `data.items[].name`: 名称。
- `data.items[].description`: 说明。
- `data.items[].icon`: 图标名。
- `data.items[].status`: 状态。
- `data.items[].created_time`: 创建时间。
- `data.items[].updated_time`: 更新时间。

## 失败响应

- `success`: `false`。
- `message`: 搜索失败原因。

