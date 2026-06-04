---
method: GET
path: /api/user/topup
auth: support
handler: controller.GetAllTopUps
source: router/api-router.go:124
request:
  query_params:
    - keyword
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
---

# GET `/api/user/topup`

支持人员或管理员分页查看全平台充值订单。

## 查询参数字段

- `keyword`: 字符串，可选。存在时按订单搜索。
- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，最大 `100`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页码。
- `data.page_size`: 实际每页条数。
- `data.total`: 总条数。
- `data.items`: 充值订单数组；字段同 `GET /api/user/topup/self` 的 `data.items[]`。

## 失败响应

- `success`: `false`。
- `message`: 查询错误。

