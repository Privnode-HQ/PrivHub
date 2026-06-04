---
method: GET
path: /api/verification
auth: public_turnstile_rate_limited
handler: controller.SendEmailVerification
source: router/api-router.go:41
request:
  query_params:
    - email
response:
  success_http_status: 200
  envelope: custom_success
---

# GET `/api/verification`

向指定邮箱发送注册/绑定用验证码。接口受邮箱验证限流和 Turnstile 校验影响。

## 查询参数字段

- `email`: 字符串，必填。目标邮箱地址，必须满足邮箱格式。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。

## 失败响应

- `success`: `false`。
- `message`: 可能为 `无效的参数`、`无效的邮箱地址`、邮箱域名/别名限制提示、`邮箱地址已被占用`，或邮件发送错误。

## 业务规则

- 开启邮箱域名白名单时，邮箱域名必须在白名单内。
- 开启邮箱别名限制时，本地部分不能包含 `+` 或 `.`。
- 已被占用的邮箱不会再次发送验证码。

