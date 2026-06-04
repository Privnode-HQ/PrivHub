---
method: POST
path: /api/setup
auth: public
handler: controller.PostSetup
source: router/api-router.go:17
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/setup`

提交首次安装向导。系统未初始化时可创建 Root 用户并保存运行模式；系统已初始化后会拒绝重复执行。

## 请求体字段

- `username`: 字符串。Root 用户名；只有尚无 Root 用户时必填，最长 12 个字符。
- `password`: 字符串。Root 用户密码；只有尚无 Root 用户时必填，至少 8 个字符。
- `confirmPassword`: 字符串。确认密码，必须与 `password` 一致。
- `SelfUseModeEnabled`: 布尔值。是否启用自用模式，会保存到全局选项。
- `DemoSiteEnabled`: 布尔值。是否启用演示站点模式，会保存到全局选项。

## 成功响应字段

- `success`: `true`。
- `message`: 固定为 `系统初始化成功`。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `系统已经初始化完成`、`请求参数有误`、`用户名长度不能超过12个字符`、`两次输入的密码不一致`、`密码长度至少为8个字符`、`创建管理员账号失败: ...`、`保存自用模式设置失败: ...`、`保存演示站点模式设置失败: ...`、`系统初始化失败: ...`。

