---
method: GET
path: /api/vendors/
auth: support
handler: controller.GetAllVendors
source: router/api-router.go:336
request:
  query_params:
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/vendors/`

支持人员或管理员分页查看模型供应商。

## 查询参数字段

- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页。
- `data.page_size`: 每页条数。
- `data.total`: 总数。
- `data.items`: 供应商数组。
- `data.items[].id`: 供应商 ID。
- `data.items[].name`: 供应商名称，软删除范围内唯一。
- `data.items[].description`: 供应商说明。
- `data.items[].icon`: 图标名，前端可按图标库渲染。
- `data.items[].status`: 状态，通常 `1` 表示启用。
- `data.items[].created_time`: 创建时间 Unix 秒。
- `data.items[].updated_time`: 更新时间 Unix 秒。

## 失败响应

- `success`: `false`。
- `message`: 查询失败原因。

