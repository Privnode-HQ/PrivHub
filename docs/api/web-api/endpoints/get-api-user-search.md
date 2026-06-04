---
method: GET
path: /api/user/search
auth: support
handler: controller.SearchUsers
source: router/api-router.go:125
request:
  query_params:
    - keyword
    - group
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/user/search`

按关键字和分组搜索用户。

## 查询参数字段

- `keyword`: 字符串，可选。搜索关键字。
- `group`: 字符串，可选。用户分组过滤。
- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，最大 `100`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页码。
- `data.page_size`: 实际每页条数。
- `data.total`: 搜索结果总数。
- `data.items`: 用户数组，字段同 `GET /api/user/` 的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: 搜索错误。

