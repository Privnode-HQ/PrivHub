package middleware

import (
	"strings"

	"github.com/QuantumNous/new-api/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func shouldSkipImpersonationAudit(path string) bool {
	switch path {
	case "/api/status", "/api/message/self/unread", "/api/user/self", "/api/user/impersonation/history":
		return true
	default:
		return false
	}
}

func ImpersonationAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			return
		}

		session := sessions.Default(c)
		state := service.GetImpersonationSessionState(session)
		if !state.Active || !state.BreakGlass || state.GrantID == 0 {
			return
		}

		path := c.Request.URL.Path
		if service.IsImpersonationStopPath(path) || shouldSkipImpersonationAudit(path) {
			return
		}

		route := c.FullPath()
		if route == "" {
			route = path
		}

		_ = service.RecordBreakGlassAction(
			state.GrantID,
			state.OriginalID,
			state.OriginalUsername,
			sessionValueToInt(session.Get("id")),
			c.Request.Method,
			path,
			route,
			c.Writer.Status(),
		)
	}
}
