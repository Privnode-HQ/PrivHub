package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/gin-gonic/gin"
)

func SetSSORouter(router *gin.Engine) {
	ssoRoute := router.Group("/sso-beta")
	{
		// SSO 入口端点，检查登录状态并重定向
		ssoRoute.GET("/v1", controller.SSOAuthRequest)
	}
}
