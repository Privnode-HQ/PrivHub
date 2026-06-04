---
method: GET
path: /api/pricing
auth: optional_user
handler: controller.GetPricing
source: router/api-router.go:40
request:
  body: none
response:
  success_http_status: 200
  envelope: pricing_legacy
---

# GET `/api/pricing`

获取价格、供应商、分组倍率和当前用户可用分组。该接口允许匿名访问；如果能解析用户身份，会按用户分组调整可用分组与倍率。

## 请求字段

无路径参数、查询参数或请求体。可携带用户会话或 Access Token；未携带时按匿名视角返回。

## 成功响应字段

该接口使用历史响应形状，多个字段位于顶层：

- `success`: `true`。
- `data`: 模型价格数据，来自 `model.GetPricing()`。
- `vendors`: 供应商元数据数组，来自 `model.GetVendors()`。
- `group_ratio`: 对象。键为可用分组名，值为该分组倍率；登录用户会应用 group-to-group 覆盖倍率。
- `usable_group`: 对象。键为用户可用分组名，值为显示名或分组说明。
- `supported_endpoint`: 对象。模型支持的 endpoint 类型映射。
- `auto_groups`: 数组或对象。当前用户分组可自动选择的分组配置。

## 错误响应

当前控制器没有显式错误分支。

