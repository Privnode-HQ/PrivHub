---
method: GET
path: /api/redemption/search
auth: support
handler: controller.SearchRedemptions
source: router/api-router.go:247
request:
  query_params:
    - keyword
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/redemption/search`

按 ID 或名称前缀搜索兑换码。

## 查询参数字段

- `keyword`: 字符串，可选。可为数字 ID 或名称前缀。
- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，最大 `100`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 分页对象，`items[]` 字段同 `GET /api/redemption/`。

## 失败响应

- `success`: `false`。
- `message`: 搜索错误。

