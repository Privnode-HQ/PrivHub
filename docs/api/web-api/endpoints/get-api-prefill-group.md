---
method: GET
path: /api/prefill_group/
auth: support
handler: controller.GetPrefillGroups
source: router/api-router.go:313
request:
  query_params:
    - type
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/prefill_group/`

获取预填组列表，可按类型过滤。

## 查询参数字段

- `type`: 字符串，可选。组类型，例如 `model`、`tag`、`endpoint`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 预填组数组。
- `data[].id`: 组 ID。
- `data[].name`: 组名称，软删除范围内唯一。
- `data[].type`: 组类型。
- `data[].items`: JSON 数组或对象，组内条目。
- `data[].description`: 描述。
- `data[].created_time`: 创建时间 Unix 秒。
- `data[].updated_time`: 更新时间 Unix 秒。

## 失败响应

- `success`: `false`。
- `message`: 查询错误。

