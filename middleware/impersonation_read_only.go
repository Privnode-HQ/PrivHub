package middleware

import (
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func isReadOnlyImpersonationForbiddenGET(path string) bool {
	switch {
	case path == "/api/user/token":
		return true
	case strings.HasPrefix(path, "/api/oauth/") && strings.HasSuffix(path, "/bind"):
		return true
	default:
		return false
	}
}

func ReadOnlyImpersonationGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if !service.IsReadOnlyImpersonation(session) {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		if service.IsImpersonationStopPath(path) {
			c.Next()
			return
		}

		method := strings.ToUpper(strings.TrimSpace(c.Request.Method))
		isSafeMethod := method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
		if !isSafeMethod || isReadOnlyImpersonationForbiddenGET(path) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "当前为只读仿冒会话，无法执行写操作",
				"code":    "READ_ONLY_IMPERSONATION",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
