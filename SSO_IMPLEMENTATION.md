# SSO 功能实现说明

## 概述

本系统实现了基于 JWT 的 SSO（单点登录）授权功能，允许第三方应用通过标准的 OAuth2 类似流程获取用户授权。

## 功能特性

- 支持用户登录状态检查
- 安全的授权确认页面
- 基于 JWT HS256 的 token 生成
- 自动处理未登录用户的重定向

## 实现的文件

### 后端文件

1. **`common/jwt.go`** - JWT 工具函数
   - `InitSSOJWT()` - 初始化 SSO JWT 密钥
   - `GenerateSSOToken()` - 生成 SSO JWT token

2. **`controller/sso.go`** - SSO 控制器
   - `SSOAuthRequest()` - 处理 SSO 授权请求入口
   - `SSOApprove()` - 处理用户授权确认
   - `SSOCancel()` - 处理用户取消授权

3. **`router/sso-router.go`** - SSO 路由配置
   - `/sso-beta/v1` - SSO 入口端点

4. **`router/api-router.go`** - API 路由配置（添加）
   - `/api/sso-beta/approve` - 授权确认 API
   - `/api/sso-beta/cancel` - 取消授权 API

5. **`common/init.go`** - 环境变量初始化（修改）
   - 添加了 `InitSSOJWT()` 调用

### 前端文件

1. **`web/src/pages/SSOAuthorize.jsx`** - SSO 授权页面组件
2. **`web/src/App.jsx`** - 路由配置（修改）
   - 添加了 `/sso-beta/authorize` 路由

### 配置文件

1. **`.env.example`** - 环境变量配置示例
   - 添加了 `SSO_JWT_SECRET` 配置说明

## 使用流程

### 1. 配置环境变量

在 `.env` 文件中设置 SSO JWT 密钥（可选）：

```bash
# SSO JWT 密钥（用于生成 SSO 授权 token）
# 如果未设置，将使用 SESSION_SECRET
SSO_JWT_SECRET=your_sso_jwt_secret_here
```

### 2. SSO 授权流程

#### 步骤 1: 第三方应用发起授权请求

第三方应用将用户重定向到：

```
https://your-domain.com/sso-beta/v1?protocol=i0&client_id=ticket-v1&nonce=xxx&metadata=xxx&postauth=callback-hostname
```

**参数说明：**
- `protocol`: 协议版本，固定为 `i0`
- `client_id`: 客户端 ID，固定为 `ticket-v1`
- `nonce`: 随机字符串，用于防止重放攻击
- `metadata`: 可选的元数据
- `postauth`: 回调地址的 hostname（不包含协议和路径）

#### 步骤 2: 用户登录（如果需要）

如果用户未登录，系统会自动重定向到登录页面。登录成功后，用户会被重定向回授权页面。

#### 步骤 3: 用户确认授权

授权页面会显示：

```
是否授权 Privnode 支持 访问您的账号信息？

Privnode 支持 将可以：
• 获取您的用户基本信息

他们无法代表您执行操作。

[授权] [取消]
```

- **授权**: 生成 JWT token 并重定向到回调地址
- **取消**: 返回主页

#### 步骤 4: 回调

如果用户点击"授权"，系统会生成 JWT token 并重定向到：

```
https://callback-hostname/sso/callback?nonce=xxx&metadata=xxx&token=jwt_token
```

### 3. JWT Token 结构

生成的 JWT token 包含以下 payload：

```json
{
  "uid": 123,
  "username": "user123",
  "authtk": "user_access_token",
  "metadata": {
    "group": "default"
  },
  "exp": 1234567890,
  "iat": 1234567890
}
```

**字段说明：**
- `uid`: 用户 ID
- `username`: 用户名
- `authtk`: 用户的访问令牌
- `metadata.group`: 用户组
- `exp`: 过期时间（5分钟后）
- `iat`: 签发时间

### 4. 验证 JWT Token

第三方应用收到 token 后，可以使用相同的密钥（`SSO_JWT_SECRET`）验证 JWT token：

```go
import "github.com/golang-jwt/jwt/v5"

// 验证 token
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method")
    }
    return []byte(ssoJWTSecret), nil
})

if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
    uid := claims["uid"].(float64)
    username := claims["username"].(string)
    // ... 使用用户信息
}
```

## 安全性考虑

1. **JWT 有效期**: Token 有效期为 5 分钟，防止长期滥用
2. **Nonce 验证**: 建议第三方应用验证返回的 nonce 与发送的一致
3. **HTTPS**: 生产环境必须使用 HTTPS 传输
4. **密钥管理**: `SSO_JWT_SECRET` 应该是强随机字符串，妥善保管
5. **协议和客户端验证**: 系统会验证 `protocol` 和 `client_id` 参数
6. **Session Cookie 配置**: 系统使用 `SameSite=Lax` 模式，允许跨站链接跳转时发送 Cookie（SSO 必需）

## 重要配置说明

### Session Cookie SameSite 策略

为了支持 SSO 跨站授权流程，系统的 Session Cookie 已配置为 `SameSite=Lax` 模式（在 `main.go:154`）。

**SameSite 模式对比：**

| 模式 | 跨站链接跳转 | SSO 支持 | 安全性 |
|------|------------|---------|--------|
| `Strict` | ❌ 不发送 Cookie | ❌ 无法使用 | 最高 |
| `Lax` | ✅ 发送 Cookie | ✅ 正常工作 | 高 |
| `None` | ✅ 发送 Cookie | ✅ 正常工作 | 中（需要 HTTPS）|

**为什么需要 Lax 模式？**
- 当用户从第三方应用（如 `example.com`）点击链接跳转到 SSO 入口时，浏览器需要发送 Session Cookie 来验证登录状态
- `Strict` 模式会完全阻止跨站请求发送 Cookie，导致已登录用户也被认为未登录
- `Lax` 模式在保持较高安全性的同时，允许顶级导航（链接跳转）发送 Cookie

## 测试示例

### 测试 URL

```
http://localhost:3000/sso-beta/v1?protocol=i0&client_id=ticket-v1&nonce=test123&metadata=testmeta&postauth=example.com
```

### 预期行为

1. 如果未登录：重定向到 `/login?redirect=/sso-beta/authorize?...`
2. 如果已登录：重定向到 `/sso-beta/authorize?...`
3. 用户点击授权：重定向到 `https://example.com/sso/callback?nonce=test123&metadata=testmeta&token=...`
4. 用户点击取消：重定向到 `/`

## 错误处理

系统会在以下情况返回错误：

- 缺少必需参数：`{"success": false, "message": "缺少必需参数"}`
- 不支持的协议：`{"success": false, "message": "不支持的协议"}`
- 无效的客户端 ID：`{"success": false, "message": "无效的客户端ID"}`
- 未登录（API 调用）：`{"success": false, "message": "未登录"}`
- 生成 token 失败：`{"success": false, "message": "生成 token 失败"}`

## 常见问题排查

### 问题 1: 已登录用户通过链接跳转仍被要求登录 🔥

**症状：**
- ✅ 直接访问 SSO URL 正常，能看到授权页面
- ❌ 从其他网站链接跳转到 SSO URL 时，已登录用户被重定向到登录页
- 浏览器 Network 显示 302 重定向

**原因：**
Session Cookie 的 `SameSite` 设置为 `Strict` 模式，浏览器在跨站链接跳转时不发送 Cookie

**解决方案：**
✅ **已修复**：系统已将 `main.go:154` 中的 Session 配置改为 `SameSiteLaxMode`

验证配置：
```go
// main.go
store.Options(sessions.Options{
    SameSite: http.SameSiteLaxMode, // ✅ 正确配置
    // SameSite: http.SameSiteStrictMode, // ❌ 错误配置
})
```

**重启应用后生效**

### 问题 2: 缺少必需参数

**症状：**
浏览器收到 JSON 错误响应：
```json
{"success": false, "message": "缺少必需参数"}
```

**原因：**
SSO URL 缺少必需的查询参数

**解决方案：**
确保 URL 包含所有必需参数：
```
✅ 正确：https://example.com/sso-beta/v1?protocol=i0&client_id=ticket-v1&nonce=xxx&postauth=callback.com
❌ 错误：https://example.com/sso-beta/v1
```

### 问题 3: Session 过期或不存在

**症状：**
用户明明已登录，但 SSO 仍然重定向到登录页

**排查方法：**
1. 打开浏览器开发者工具 → Application/Storage → Cookies
2. 检查是否存在 `session` Cookie
3. 访问 `/api/user/self` 验证登录状态

**可能原因：**
- Session 已过期（30天后）
- Cookie 被清除
- 不同子域名的 Cookie 域设置问题

## 注意事项

1. 确保前端已经构建并部署
2. 确保 `.env` 文件中配置了必要的环境变量
3. 生产环境需要设置强随机的 `SSO_JWT_SECRET`
4. 回调地址需要支持 HTTPS 协议

## 技术栈

- **后端**: Go + Gin Framework
- **前端**: React + Semi-UI
- **JWT**: github.com/golang-jwt/jwt/v5
- **会话**: gin-contrib/sessions
