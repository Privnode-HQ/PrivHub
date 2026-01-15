package common

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var SSOJWTSecret string

// InitSSOJWT 初始化 SSO JWT 密钥
func InitSSOJWT() {
	SSOJWTSecret = os.Getenv("SSO_JWT_SECRET")
	if SSOJWTSecret == "" {
		// 如果未设置，使用 SESSION_SECRET
		SSOJWTSecret = SessionSecret
	}
}

// SSOTokenPayload SSO Token 的 payload 结构
type SSOTokenPayload struct {
	UID      int               `json:"uid"`
	Username string            `json:"username"`
	AuthTK   string            `json:"authtk"`
	Metadata map[string]string `json:"metadata"`
	jwt.RegisteredClaims
}

// GenerateSSOToken 生成 SSO JWT token
func GenerateSSOToken(uid int, username string, authtk string, group string) (string, error) {
	// 创建 token payload
	claims := SSOTokenPayload{
		UID:      uid,
		Username: username,
		AuthTK:   authtk,
		Metadata: map[string]string{
			"group": group,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)), // 5分钟有效期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并返回
	tokenString, err := token.SignedString([]byte(SSOJWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
