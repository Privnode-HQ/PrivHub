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
