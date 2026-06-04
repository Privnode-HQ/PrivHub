---
method: GET
path: /api/option/
auth: root
handler: controller.GetOptions
source: router/api-router.go:149
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/option/`

Root 获取全局选项列表。含 `Token`、`Secret`、`Key` 后缀的敏感选项不会返回。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data`: 选项数组。
- `data[].key`: 选项键名。
- `data[].value`: 选项值，统一转为字符串。

## 失败响应

该控制器没有显式错误分支；鉴权失败由中间件处理。

