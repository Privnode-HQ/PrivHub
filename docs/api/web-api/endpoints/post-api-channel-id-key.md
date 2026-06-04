---
method: POST
path: /api/channel/:id/key
auth: root + secure verification
handler: controller.GetChannelKey
source: router/api-router.go:198
request:
  path_params:
    - id
response:
  success_http_status: 200
  envelope: raw-json
---

# POST `/api/channel/:id/key`

root 管理员在通过安全验证后查看渠道密钥。

## 路径参数字段

- `id`: 渠道 ID，必须是整数。

## 中间件要求

- 必须通过管理员认证、管理员审计、root 权限校验。
- 受关键操作限流保护。
- 禁用缓存。
- 必须满足 `SecureVerificationRequired` 安全验证要求。

## 成功响应字段

- `success`: `true`。
- `message`: `获取成功`。
- `data.key`: 渠道密钥。单密钥渠道为密钥字符串；多密钥渠道可能为换行分隔字符串或 Vertex AI JSON 数组字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `渠道ID格式错误`、`获取渠道信息失败`、`渠道不存在`，或安全验证/权限中间件返回的错误。

