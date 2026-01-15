package controller

import (
	"net/http"
	"net/url"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// SSOAuthRequest 处理 SSO 授权请求
func SSOAuthRequest(c *gin.Context) {
	// 获取查询参数
	protocol := c.Query("protocol")
	clientID := c.Query("client_id")
	nonce := c.Query("nonce")
	metadata := c.Query("metadata")
	postauth := c.Query("postauth")

	// 验证必需参数
	if protocol == "" || clientID == "" || nonce == "" || postauth == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "缺少必需参数",
		})
		return
	}

	// 验证 protocol
	if protocol != "i0" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不支持的协议",
		})
		return
	}

	// 验证 client_id
	if clientID != "ticket-v1" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的客户端ID",
		})
		return
	}

	// 检查用户是否已登录
	session := sessions.Default(c)
	userID := session.Get("id")

	if userID == nil {
		// 用户未登录，保存原始请求参数到 session
		session.Set("sso_protocol", protocol)
		session.Set("sso_client_id", clientID)
		session.Set("sso_nonce", nonce)
		session.Set("sso_metadata", metadata)
		session.Set("sso_postauth", postauth)
		session.Save()

		// 重定向到前端登录页面，登录后返回到授权页面
		c.Redirect(http.StatusFound, "/login?redirect=/sso-beta/authorize?protocol="+url.QueryEscape(protocol)+
			"&client_id="+url.QueryEscape(clientID)+
			"&nonce="+url.QueryEscape(nonce)+
			"&metadata="+url.QueryEscape(metadata)+
			"&postauth="+url.QueryEscape(postauth))
		return
	}

	// 用户已登录，重定向到授权页面
	c.Redirect(http.StatusFound, "/sso-beta/authorize?protocol="+url.QueryEscape(protocol)+
		"&client_id="+url.QueryEscape(clientID)+
		"&nonce="+url.QueryEscape(nonce)+
		"&metadata="+url.QueryEscape(metadata)+
		"&postauth="+url.QueryEscape(postauth))
}

// SSOApprove 处理用户授权确认
func SSOApprove(c *gin.Context) {
	// 检查用户是否已登录
	session := sessions.Default(c)
	userID := session.Get("id")

	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未登录",
		})
		return
	}

	// 获取请求参数
	var request struct {
		Nonce    string `json:"nonce" binding:"required"`
		Metadata string `json:"metadata"`
		Postauth string `json:"postauth" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数",
		})
		return
	}

	// 获取用户信息
	uid := userID.(int)
	username := session.Get("username").(string)
	group := session.Get("group").(string)

	// 获取用户的 AccessToken 作为 authtk
	user, err := model.GetUserById(uid, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户信息失败",
		})
		return
	}

	// AccessToken 可能为 nil，需要检查
	authtk := ""
	if user.AccessToken != nil {
		authtk = *user.AccessToken
	}

	// 生成 JWT token
	token, err := common.GenerateSSOToken(uid, username, authtk, group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "生成 token 失败",
		})
		return
	}

	// 构建回调 URL
	callbackURL := "https://" + request.Postauth + "/sso/callback?nonce=" + url.QueryEscape(request.Nonce) +
		"&token=" + url.QueryEscape(token)

	// 如果有 metadata，添加到 URL
	if request.Metadata != "" {
		callbackURL += "&metadata=" + url.QueryEscape(request.Metadata)
	}

	// 返回回调 URL
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"redirect_url": callbackURL,
		},
	})
}

// SSOCancel 处理用户取消授权
func SSOCancel(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"redirect_url": "/",
		},
	})
}
