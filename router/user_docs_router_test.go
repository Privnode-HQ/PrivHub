package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
)

func TestUserDocsRoutesServeRuntimeConfiguredDocs(t *testing.T) {
	staticDir := t.TempDir()
	writeTestFile(t, filepath.Join(staticDir, "index.html"), "51API https://51-api.com https://51-api.com/v1")
	writeTestFile(t, filepath.Join(staticDir, "quickstart", "codex", "index.html"), "Codex uses https://51-api.com/v1")
	writeTestFile(t, filepath.Join(staticDir, "RSC", "R", "quickstart", "codex.txt"), "51API https://51-api.com/v1")
	writeTestFile(t, filepath.Join(staticDir, "assets", "app.css"), "body{color:#111}")

	router := setupUserDocsTestRouter(t, staticDir)

	t.Run("docs root redirects to slash base path", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/docs?from=nav", nil)
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusTemporaryRedirect {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusTemporaryRedirect)
		}
		if got := recorder.Header().Get("Location"); got != "/docs/?from=nav" {
			t.Fatalf("Location = %q, want %q", got, "/docs/?from=nav")
		}
	})

	tests := []struct {
		name         string
		path         string
		wantBody     string
		wantCache    string
		wantContent  string
		wantVaryPart string
	}{
		{
			name:         "docs root slash",
			path:         "/docs/",
			wantBody:     "Demo Hub https://demo.example https://demo.example/v1",
			wantCache:    userDocsDynamicCacheControl,
			wantContent:  "text/html; charset=utf-8",
			wantVaryPart: "X-Forwarded-Host",
		},
		{
			name:         "nested docs page",
			path:         "/docs/quickstart/codex",
			wantBody:     "Codex uses https://demo.example/v1",
			wantCache:    userDocsDynamicCacheControl,
			wantContent:  "text/html; charset=utf-8",
			wantVaryPart: "X-Forwarded-Proto",
		},
		{
			name:         "rsc payload",
			path:         "/docs/RSC/R/quickstart/codex.txt",
			wantBody:     "Demo Hub https://demo.example/v1",
			wantCache:    userDocsDynamicCacheControl,
			wantContent:  "text/plain; charset=utf-8",
			wantVaryPart: "Host",
		},
		{
			name:         "hashed asset",
			path:         "/docs/assets/app.css",
			wantBody:     "body{color:#111}",
			wantCache:    userDocsImmutableCacheControl,
			wantContent:  "text/css; charset=utf-8",
			wantVaryPart: "Accept-Encoding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d, body = %q", recorder.Code, http.StatusOK, recorder.Body.String())
			}
			if got := recorder.Body.String(); got != tt.wantBody {
				t.Fatalf("body = %q, want %q", got, tt.wantBody)
			}
			if got := recorder.Header().Get("Cache-Control"); got != tt.wantCache {
				t.Fatalf("Cache-Control = %q, want %q", got, tt.wantCache)
			}
			if got := recorder.Header().Get("Content-Type"); got != tt.wantContent {
				t.Fatalf("Content-Type = %q, want %q", got, tt.wantContent)
			}
			if got := recorder.Header().Get("Vary"); !strings.Contains(got, tt.wantVaryPart) {
				t.Fatalf("Vary = %q, want it to contain %q", got, tt.wantVaryPart)
			}
		})
	}
}

func TestInstallScriptsUseConfiguredBaseURL(t *testing.T) {
	router := setupUserDocsTestRouter(t, t.TempDir())

	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "shell",
			path: "/install.sh",
			want: []string{
				"SYSTEM_NAME='Demo Hub'",
				"ANTHROPIC_BASE_URL='https://demo.example'",
				"OPENAI_BASE_URL='https://demo.example/v1'",
				"https://claude.ai/install.sh",
				"https://chatgpt.com/codex/install.sh",
				"@anthropic-ai/claude-code",
				"@openai/codex",
			},
		},
		{
			name: "powershell",
			path: "/install.ps1",
			want: []string{
				"$SystemName = 'Demo Hub'",
				"$AnthropicBaseUrl = 'https://demo.example'",
				"$OpenAIBaseUrl = 'https://demo.example/v1'",
				"Anthropic.ClaudeCode",
				"https://claude.ai/install.ps1",
				"https://chatgpt.com/codex/install.ps1",
				"@anthropic-ai/claude-code",
				"@openai/codex",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			if got := recorder.Header().Get("Cache-Control"); got != userDocsDynamicCacheControl {
				t.Fatalf("Cache-Control = %q, want %q", got, userDocsDynamicCacheControl)
			}
			body := recorder.Body.String()
			for _, want := range tt.want {
				if !strings.Contains(body, want) {
					t.Fatalf("script body does not contain %q", want)
				}
			}
		})
	}
}

func setupUserDocsTestRouter(t *testing.T, staticDir string) *gin.Engine {
	t.Helper()

	t.Setenv(userDocsStaticDirEnv, staticDir)
	previousSystemName := common.SystemName
	previousServerAddress := system_setting.ServerAddress
	common.SystemName = "Demo Hub"
	system_setting.ServerAddress = "https://demo.example"
	t.Cleanup(func() {
		common.SystemName = previousSystemName
		system_setting.ServerAddress = previousServerAddress
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetUserDocsRouter(router)
	return router
}

func writeTestFile(t *testing.T, filePath string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("create test dir: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
