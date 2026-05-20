package middleware

import (
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

const (
	frontendHTMLCacheControl      = "public, max-age=604800"
	frontendStaticCacheControl    = "public, max-age=604800"
	frontendImmutableCacheControl = "public, max-age=31536000, immutable"
)

func Cache(fileSystem static.ServeFileSystem) func(c *gin.Context) {
	return func(c *gin.Context) {
		if fileSystem.Exists("/", c.Request.URL.Path) {
			setFrontendCacheHeader(c, c.Request.URL.Path)
		}
		c.Next()
	}
}

func SetFrontendHTMLCache(c *gin.Context) {
	setCacheHeader(c, frontendHTMLCacheControl)
}

func setFrontendCacheHeader(c *gin.Context, requestPath string) {
	if requestPath == "/" || strings.HasSuffix(requestPath, "/index.html") {
		SetFrontendHTMLCache(c)
		return
	}
	if strings.HasPrefix(requestPath, "/assets/") {
		setCacheHeader(c, frontendImmutableCacheControl)
		return
	}
	setCacheHeader(c, frontendStaticCacheControl)
}

func setCacheHeader(c *gin.Context, cacheControl string) {
	c.Header("Cache-Control", cacheControl)
	c.Header("Vary", "Accept-Encoding")
}
