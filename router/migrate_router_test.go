package router

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetMigrateRouterDoesNotPanicAndUsesStaticPrefixes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("SetMigrateRouter panicked: %v", recovered)
		}
	}()
	SetMigrateRouter(engine)

	routes := map[string]bool{}
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	required := []string{
		"POST /migrate/api/migrations/:migrate_id/verify",
		"POST /migrate/api/migrations/:migrate_id/login",
		"POST /migrate/api/migrations/:migrate_id/capture",
		"POST /migrate/api/imports/setup/verify",
		"POST /migrate/api/imports/setup/password",
		"GET /migrate/api/admin/expression-docs",
	}
	for _, key := range required {
		if !routes[key] {
			t.Fatalf("expected route %s to be registered", key)
		}
	}
	if routes["POST /migrate/api/:migrate_id/verify"] {
		t.Fatal("old wildcard route should not be registered")
	}
}
