---
method: POST
path: /api/user/2fa/backup_codes
auth: user
handler: controller.RegenerateBackupCodes
source: router/api-router.go:117
request:
  content_type: application/json
response:
  success_http_status: 200
  envelope: custom_success
---

# POST `/api/user/2fa/backup_codes`

验证 TOTP 后重新生成当前用户 2FA 备用码。

## 请求体字段

- `code`: 字符串，必填。TOTP 数字验证码。

## 成功响应字段

- `success`: `true`。
- `message`: `备用码重新生成成功`。
- `data.backup_codes`: 新备用码数组。旧备用码会被新记录替代。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `参数错误`、`用户未启用2FA`、`验证码或备用码错误，请重试`、`生成备用码失败`、`保存备用码失败`，或验证码格式/数据库错误。

