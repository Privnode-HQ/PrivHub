# API 鉴权文档

## 认证方式

### Access Token

对于需要鉴权的 API 接口，必须同时提供以下两个请求头来进行 Access Token 认证：

1. **请求头中的 `Authorization` 字段**

    将 Access Token 放置于 HTTP 请求头部的 `Authorization` 字段中，格式如下：

    ```
    Authorization: <your_access_token>
    ```

    其中 `<your_access_token>` 需要替换为实际的 Access Token 值。

2. **请求头中的 `New-Api-User` 字段**

    将用户 ID 放置于 HTTP 请求头部的 `New-Api-User` 字段中，格式如下：

    ```
    New-Api-User: <your_user_id>
    ```

    其中 `<your_user_id>` 需要替换为实际的用户 ID。

**注意：**

*   **必须同时提供 `Authorization` 和 `New-Api-User` 两个请求头才能通过鉴权。**
*   如果只提供其中一个请求头，或者两个请求头都未提供，则会返回 `401 Unauthorized` 错误。
*   如果 `Authorization` 中的 Access Token 无效，则会返回 `401 Unauthorized` 错误，并提示“无权进行此操作，access token 无效”。
*   如果 `New-Api-User` 中的用户 ID 与 Access Token 不匹配，则会返回 `401 Unauthorized` 错误，并提示“无权进行此操作，与登录用户不匹配，请重新登录”。
*   如果没有提供 `New-Api-User` 请求头，则会返回 `401 Unauthorized` 错误，并提示“无权进行此操作，未提供 New-Api-User”。
*   如果 `New-Api-User` 请求头格式错误，则会返回 `401 Unauthorized` 错误，并提示“无权进行此操作，New-Api-User 格式错误”。
*   如果用户已被禁用，则会返回 `403 Forbidden` 错误，并提示“用户已被封禁”（若设置了封禁原因会在提示中附带原因）。
*   如果用户权限不足，则会返回 `403 Forbidden` 错误，并提示“无权进行此操作，权限不足”。
*   如果用户信息无效，则会返回 `403 Forbidden` 错误，并提示“无权进行此操作，用户信息无效”。

### Admin Service Account (ASA)

管理员可以在 Web 控制台的 **Admin Service Account** 页面创建 ASA，并获得一次性显示的 JWT 凭据。ASA JWT 用于管理员自动化调用现有 Web API，调用时只需要提供标准 Bearer Token：

```http
Authorization: Bearer <asa_jwt>
```

使用 ASA JWT 时不需要、也不应该额外传递 `New-Api-User` 请求头。后端会从 JWT 与数据库中的 Service Account 记录解析管理员身份，并验证：

* JWT 签名、`iss`、`aud`、`exp`、`nbf`、`iat`、`jti`；
* `token_use=admin_service_account`、版本号、Service Account ID；
* Service Account 当前为启用状态、未被删除、未过期、未被轮换；
* 绑定用户仍然存在、处于启用状态，并且仍然是管理员或超级管理员；
* 如果配置了 IP 白名单，请求来源 IP 必须匹配白名单。

ASA JWT 的载荷包含管理员自动化审计所需信息，包括 Service Account ID/名称、绑定管理员的用户 ID、CAH ID、用户名、角色、创建者 ID/CAH ID/用户名、作用域、签发时间、启用时间、过期时间和唯一 `jti`。JWT 凭据只在创建或轮换时返回一次；后端仅保存凭据哈希，无法再次取回原文。

## Curl 示例

假设您的 Access Token 为 `access_token`，用户 ID 为 `123`，要访问的 API 接口为 `/api/user/self`，则可以使用以下 curl 命令：

```bash
curl -X GET \
  -H "Authorization: access_token" \
  -H "New-Api-User: 123" \
  https://your-domain.com/api/user/self
```

请将 `access_token`、`123` 和 `https://your-domain.com` 替换为实际的值。

使用 ASA JWT 调用管理员接口时：

```bash
curl -X GET \
  -H "Authorization: Bearer <asa_jwt>" \
  https://your-domain.com/api/user/self
```
