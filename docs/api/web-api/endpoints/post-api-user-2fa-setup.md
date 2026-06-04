---
method: POST
path: /api/user/2fa/setup
auth: user
handler: controller.Setup2FA
source: router/api-router.go:114
request:
  body: none
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/2fa/setup`

初始化当前用户 2FA 设置，生成 TOTP secret、二维码数据和备用码；此时尚未启用，需要继续调用启用接口。

## 请求字段

无路径参数、查询参数或请求体。

## 成功响应字段

- `success`: `true`。
- `message`: `2FA设置初始化成功，请使用认证器扫描二维码并输入验证码完成设置`。
- `data.secret`: TOTP secret。
- `data.qr_code_data`: 二维码 otpauth 数据。
- `data.backup_codes`: 新生成的备用码数组。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `用户已启用2FA，请先禁用后重新设置`、`生成2FA密钥失败`、`生成备用码失败`、`保存备用码失败`，或用户/2FA 记录错误。

