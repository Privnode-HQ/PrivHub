---
method: POST
path: /api/option/migrate_console_setting
auth: root
handler: controller.MigrateConsoleSetting
source: router/api-router.go:152
request:
  body: none
response:
  success_http_status: 200
  envelope: common
---

# POST `/api/option/migrate_console_setting`

迁移旧版控制台配置键。路由注释标记为迁移检测用旧键，后续版本可能删除。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串或迁移成功消息。
- `data`: 迁移结果，具体结构由 `controller.MigrateConsoleSetting` 返回。

## 失败响应

- `success`: `false`。
- `message`: 迁移或配置更新错误。

