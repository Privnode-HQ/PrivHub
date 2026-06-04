---
method: GET
path: /api/token/search
auth: user
handler: controller.SearchTokens
source: router/api-router.go:223
request:
  query_params:
    - keyword
    - token
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/token/search`

搜索当前用户 Token。

## 查询参数字段

- `keyword`: 字符串，可选。按名称搜索。
- `token`: 字符串，可选。按 key 片段搜索，可带 `sk-` 前缀。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: Token 数组，字段同 `GET /api/token/` 的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: 搜索错误。

