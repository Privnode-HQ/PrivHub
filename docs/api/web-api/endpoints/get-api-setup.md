---
method: GET
path: /api/setup
auth: public
handler: controller.GetSetup
source: router/api-router.go:16
request:
  path_params: []
  query_params: []
  body: none
response:
  success_http_status: 200
  envelope: custom_success
  data: Setup
---

# GET `/api/setup`

获取首次安装向导状态。该接口公开可访问，用于判断系统是否已完成初始化以及数据库类型。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: 固定为 `true`。
- `data.status`: 布尔值，系统是否已完成初始化；对应 `constant.Setup`。
- `data.root_init`: 布尔值，是否已经存在 Root 用户；系统已初始化时可能省略或为默认 `false`。
- `data.database_type`: 字符串，当前数据库类型；可能值为 `mysql`、`postgres`、`sqlite`，初始化完成后可能为空。

## 错误响应

当前控制器没有显式错误分支。

