---
method: GET
path: /api/oauth/github
auth: public_oauth_callback
handler: controller.GitHubOAuth
source: router/api-router.go:44
request:
  query_params:
    - state
    - code
response:
  success_http_status: 200
  forbidden_http_status: 403
  envelope: custom_or_login
---

# GET `/api/oauth/github`

GitHub OAuth 回调。未登录时用于登录/注册；已有登录 session 时转为绑定 GitHub 账号。

## 查询参数字段

- `state`: 字符串，必填。必须等于 session 中 `/api/oauth/state` 生成的 state。
- `code`: 字符串，必填。GitHub 授权码。

## 成功响应字段

登录或注册成功后响应同 `POST /api/user/login`：

- `success`: `true`。
- `message`: 空字符串。
- `data.cah_id`: 用户 CAH ID。
- `data.username`: 用户名。
- `data.display_name`: 显示名。
- `data.role`: 用户角色。
- `data.status`: 用户状态。
- `data.group`: 用户分组。
- `data.email`: 绑定邮箱。
- `data.force_password_reset`: 是否需要强制重置密码。
- `data.force_email_bind`: 是否需要强制绑定邮箱。
- `data.required_actions`: 当前用户必须完成的动作列表。

绑定成功时可能重定向到个人控制台。

## 失败响应

- HTTP 403: `state is empty or not same`。
- HTTP 200 且 `success=false`: 可能为 `管理员未开启通过 GitHub 登录以及注册`、`用户已注销`、`管理员关闭了新用户注册`、用户被禁用提示，或 GitHub API 错误。

