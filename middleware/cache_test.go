package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeServeFileSystem map[string]bool

func (f fakeServeFileSystem) Exists(_ string, requestPath string) bool {
	return f[requestPath]
}

func (f fakeServeFileSystem) Open(_ string) (http.File, error) {
	return nil, os.ErrNotExist
}

func TestCacheSetsFrontendHeadersOnlyForExistingStaticFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requestPath      string
		fileExists       bool
		wantCacheControl string
		wantVary         string
	}{
		{
			name:             "hashed asset",
			requestPath:      "/assets/index-CVowx1nb.css",
			fileExists:       true,
			wantCacheControl: frontendImmutableCacheControl,
			wantVary:         "Accept-Encoding",
		},
		{
			name:             "entry html",
			requestPath:      "/index.html",
			fileExists:       true,
			wantCacheControl: frontendHTMLCacheControl,
			wantVary:         "Accept-Encoding",
		},
		{
			name:             "root static file",
			requestPath:      "/favicon.ico",
			fileExists:       true,
			wantCacheControl: frontendStaticCacheControl,
			wantVary:         "Accept-Encoding",
		},
		{
			name:        "missing asset",
			requestPath: "/assets/missing.js",
		},
		{
			name:        "missing api route",
			requestPath: "/api/missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileSystem := fakeServeFileSystem{}
			if tt.fileExists {
				fileSystem[tt.requestPath] = true
			}

			router := gin.New()
			router.Use(Cache(fileSystem))
			router.GET("/*path", func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			router.ServeHTTP(recorder, request)

			if got := recorder.Header().Get("Cache-Control"); got != tt.wantCacheControl {
				t.Fatalf("Cache-Control = %q, want %q", got, tt.wantCacheControl)
			}
			if got := recorder.Header().Get("Vary"); got != tt.wantVary {
				t.Fatalf("Vary = %q, want %q", got, tt.wantVary)
			}
		})
	}
}

func TestSetFrontendHTMLCache(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		SetFrontendHTMLCache(c)
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Cache-Control"); got != frontendHTMLCacheControl {
		t.Fatalf("Cache-Control = %q, want %q", got, frontendHTMLCacheControl)
	}
	if got := recorder.Header().Get("Vary"); got != "Accept-Encoding" {
		t.Fatalf("Vary = %q, want %q", got, "Accept-Encoding")
	}
}
