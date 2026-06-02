package router

import (
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
)

const (
	userDocsStaticDirEnv     = "USER_DOCS_STATIC_DIR"
	defaultUserDocsStaticDir = "user-docs/dist/public"

	userDocsDynamicCacheControl   = "no-cache, max-age=0, must-revalidate"
	userDocsStaticCacheControl    = "public, max-age=604800"
	userDocsImmutableCacheControl = "public, max-age=31536000, immutable"
)

type userDocsHandler struct {
	staticDir string
}

func SetUserDocsRouter(router *gin.Engine) {
	handler := &userDocsHandler{staticDir: resolveUserDocsStaticDir()}

	router.GET("/install.sh", serveInstallShell)
	router.GET("/install.ps1", serveInstallPowerShell)
	router.GET("/docs", handler.Serve)
	router.GET("/docs/*filepath", handler.Serve)
}

func resolveUserDocsStaticDir() string {
	if configured := strings.TrimSpace(os.Getenv(userDocsStaticDirEnv)); configured != "" {
		return configured
	}
	for _, candidate := range []string{
		defaultUserDocsStaticDir,
		filepath.Join("/app", defaultUserDocsStaticDir),
	} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return defaultUserDocsStaticDir
}

func serveInstallShell(c *gin.Context) {
	c.Header("Content-Type", "text/x-shellscript; charset=utf-8")
	c.Header("Cache-Control", userDocsDynamicCacheControl)
	c.Header("Vary", "Host, X-Forwarded-Host, X-Forwarded-Proto")
	c.String(http.StatusOK, buildInstallShell(c))
}

func serveInstallPowerShell(c *gin.Context) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Cache-Control", userDocsDynamicCacheControl)
	c.Header("Vary", "Host, X-Forwarded-Host, X-Forwarded-Proto")
	c.String(http.StatusOK, buildInstallPowerShell(c))
}

func (h *userDocsHandler) Serve(c *gin.Context) {
	filePath, err := h.resolveFile(c.Request.URL.Path)
	if err != nil {
		c.String(http.StatusNotFound, "documentation file not found")
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.String(http.StatusServiceUnavailable, "user-docs static files are not built; run `pnpm --dir user-docs run build` first")
			return
		}
		c.String(http.StatusInternalServerError, "failed to read documentation file")
		return
	}

	relPath := filepath.ToSlash(strings.TrimPrefix(filePath, h.staticDir))
	relPath = strings.TrimPrefix(relPath, "/")
	contentType := userDocsContentType(relPath, content)
	asset := strings.HasPrefix(relPath, "assets/")
	dynamic := !asset && shouldTransformUserDocsFile(relPath, contentType)

	if dynamic {
		content = replaceUserDocsRuntimeConfig(c, content)
		c.Header("Cache-Control", userDocsDynamicCacheControl)
		c.Header("Vary", "Accept-Encoding, Host, X-Forwarded-Host, X-Forwarded-Proto")
	} else if asset {
		c.Header("Cache-Control", userDocsImmutableCacheControl)
		c.Header("Vary", "Accept-Encoding")
	} else {
		c.Header("Cache-Control", userDocsStaticCacheControl)
		c.Header("Vary", "Accept-Encoding")
	}

	c.Data(http.StatusOK, contentType, content)
}

func (h *userDocsHandler) resolveFile(requestPath string) (string, error) {
	cleanPath := path.Clean("/" + strings.TrimPrefix(requestPath, "/docs"))
	if cleanPath == "/" || cleanPath == "." {
		return filepath.Join(h.staticDir, "index.html"), nil
	}

	rel := strings.TrimPrefix(cleanPath, "/")
	candidate := filepath.Join(h.staticDir, filepath.FromSlash(rel))
	if info, err := os.Stat(candidate); err == nil {
		if info.IsDir() {
			return filepath.Join(candidate, "index.html"), nil
		}
		return candidate, nil
	}

	if path.Ext(rel) == "" {
		return filepath.Join(candidate, "index.html"), nil
	}
	return "", os.ErrNotExist
}

func userDocsContentType(filePath string, content []byte) string {
	switch strings.ToLower(path.Ext(filePath)) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".md":
		return "text/markdown; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".webp":
		return "image/webp"
	}
	if contentType := mime.TypeByExtension(path.Ext(filePath)); contentType != "" {
		return contentType
	}
	return http.DetectContentType(content)
}

func shouldTransformUserDocsFile(filePath string, contentType string) bool {
	ext := strings.ToLower(path.Ext(filePath))
	switch ext {
	case ".html", ".js", ".txt", ".md", ".json":
		return true
	default:
		return strings.HasPrefix(contentType, "text/")
	}
}

func replaceUserDocsRuntimeConfig(c *gin.Context, content []byte) []byte {
	baseURL := configuredPublicBaseURL(c)
	replacements := []string{
		"https://51-api.com/v1", baseURL + "/v1",
		"https://51-api.com", baseURL,
		"51API", sanitizeRuntimeText(common.SystemName),
	}
	replacer := strings.NewReplacer(replacements...)
	return []byte(replacer.Replace(string(content)))
}

func configuredPublicBaseURL(c *gin.Context) string {
	if configured := strings.TrimSpace(system_setting.ServerAddress); configured != "" {
		return strings.TrimRight(configured, "/")
	}
	return requestOrigin(c)
}

func requestOrigin(c *gin.Context) string {
	proto := firstForwardedValue(c.GetHeader("X-Forwarded-Proto"))
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	host := firstForwardedValue(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = c.Request.Host
	}
	if host == "" {
		return "https://51-api.com"
	}
	return strings.TrimRight(proto+"://"+host, "/")
}

func firstForwardedValue(value string) string {
	if value == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(value, ",")[0])
}

func sanitizeRuntimeText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "51API"
	}
	return strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"&", " ",
		"<", " ",
		">", " ",
		"`", " ",
		"\r", " ",
		"\n", " ",
	).Replace(value)
}

func buildInstallShell(c *gin.Context) string {
	baseURL := configuredPublicBaseURL(c)
	replacer := strings.NewReplacer(
		"{{SYSTEM_NAME}}", shellQuote(sanitizeRuntimeText(common.SystemName)),
		"{{ANTHROPIC_BASE_URL}}", shellQuote(baseURL),
		"{{OPENAI_BASE_URL}}", shellQuote(baseURL+"/v1"),
	)
	return replacer.Replace(installShellTemplate)
}

func buildInstallPowerShell(c *gin.Context) string {
	baseURL := configuredPublicBaseURL(c)
	replacer := strings.NewReplacer(
		"{{SYSTEM_NAME}}", powerShellQuote(sanitizeRuntimeText(common.SystemName)),
		"{{ANTHROPIC_BASE_URL}}", powerShellQuote(baseURL),
		"{{OPENAI_BASE_URL}}", powerShellQuote(baseURL+"/v1"),
	)
	return replacer.Replace(installPowerShellTemplate)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func powerShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

const installShellTemplate = `#!/usr/bin/env bash
set -euo pipefail

SYSTEM_NAME={{SYSTEM_NAME}}
ANTHROPIC_BASE_URL={{ANTHROPIC_BASE_URL}}
OPENAI_BASE_URL={{OPENAI_BASE_URL}}

say() { printf '%s\n' "$*"; }
has() { command -v "$1" >/dev/null 2>&1; }
ask() {
  local prompt="${1:-}"
  local answer
  if [ -r /dev/tty ]; then
    printf '%s' "$prompt" > /dev/tty
    IFS= read -r answer < /dev/tty || true
  else
    printf '%s' "$prompt"
    IFS= read -r answer || true
  fi
  printf '%s' "$answer"
}
ask_secret() {
  local prompt="${1:-}"
  local answer
  if [ -r /dev/tty ]; then
    printf '%s' "$prompt" > /dev/tty
    stty -echo < /dev/tty 2>/dev/null || true
    IFS= read -r answer < /dev/tty || true
    stty echo < /dev/tty 2>/dev/null || true
    printf '\n' > /dev/tty
  else
    printf '%s' "$prompt"
    IFS= read -r answer || true
  fi
  printf '%s' "$answer"
}
run_sudo() {
  if [ "$(id -u)" = "0" ]; then
    "$@"
  elif has sudo; then
    sudo "$@"
  else
    say "需要 root 权限执行：$*"
    exit 1
  fi
}
backup_file() {
  local file="$1"
  if [ -f "$file" ]; then
    cp "$file" "$file.bak.$(date +%s)"
  fi
}
escape_double() {
  printf '%s' "$1" | sed 's/[\"\\]/\\&/g'
}

install_node_if_needed() {
  if has node && has npm; then
    return
  fi
  local answer
  answer="$(ask "没有检测到 Node.js/npm。是否安装 Node.js LTS？[Y/n] ")"
  case "$answer" in
    n|N|no|NO) return ;;
  esac

  case "$(uname -s)" in
    Darwin)
      if has brew; then
        brew install node
      else
        say "未检测到 Homebrew。请先从 https://nodejs.org 安装 Node.js LTS，或安装 Homebrew 后重试。"
      fi
      ;;
    Linux)
      if has apt-get; then
        run_sudo apt-get update
        run_sudo apt-get install -y nodejs npm
      elif has dnf; then
        run_sudo dnf install -y nodejs npm
      elif has yum; then
        run_sudo yum install -y nodejs npm
      elif has pacman; then
        run_sudo pacman -S --noconfirm nodejs npm
      elif has apk; then
        run_sudo apk add --no-cache nodejs npm
      else
        say "未识别当前 Linux 包管理器。请从 https://nodejs.org 安装 Node.js LTS 后重试。"
      fi
      ;;
    *)
      say "请从 https://nodejs.org 安装 Node.js LTS 后重试。"
      ;;
  esac
}

install_claude_if_needed() {
  if has claude; then
    say "已检测到 Claude Code：$(claude --version 2>/dev/null || printf installed)"
    return
  fi
  say "正在安装 Claude Code..."
  if has curl; then
    if curl -fsSL https://claude.ai/install.sh | bash; then
      return
    fi
    say "Claude Code 官方安装脚本失败，尝试 npm 安装。"
  fi
  install_node_if_needed
  if has npm; then
    npm install -g @anthropic-ai/claude-code
  else
    say "未找到 npm，无法自动安装 Claude Code。"
  fi
}

install_codex_if_needed() {
  if has codex; then
    say "已检测到 Codex CLI：$(codex --version 2>/dev/null || printf installed)"
    return
  fi
  say "正在安装 Codex CLI..."
  if has curl; then
    if curl -fsSL https://chatgpt.com/codex/install.sh | sh; then
      return
    fi
    say "Codex 官方安装脚本失败，尝试 npm 安装。"
  fi
  install_node_if_needed
  if has npm; then
    npm install -g @openai/codex
  else
    say "未找到 npm，无法自动安装 Codex CLI。"
  fi
}

configure_claude() {
  local api_key="$1"
  mkdir -p "$HOME/.claude"
  local settings="$HOME/.claude/settings.json"
  backup_file "$settings"
  cat > "$settings" <<EOF
{
  "env": {
    "ANTHROPIC_BASE_URL": "$ANTHROPIC_BASE_URL",
    "ANTHROPIC_AUTH_TOKEN": "$(escape_double "$api_key")",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
EOF
  say "已写入 $settings"
}

configure_codex() {
  local api_key="$1"
  local model="$2"
  mkdir -p "$HOME/.codex"
  local config="$HOME/.codex/config.toml"
  local auth="$HOME/.codex/auth.json"
  backup_file "$config"
  backup_file "$auth"
  cat > "$config" <<EOF
model_provider = "51api"
model = "$(escape_double "$model")"
model_reasoning_effort = "high"
model_verbosity = "high"
disable_response_storage = true

[features]
web_search_request = true

[model_providers.51api]
name = "$SYSTEM_NAME"
base_url = "$OPENAI_BASE_URL"
wire_api = "responses"
requires_openai_auth = true
EOF
  cat > "$auth" <<EOF
{
  "OPENAI_API_KEY": "$(escape_double "$api_key")"
}
EOF
  say "已写入 $config 和 $auth"
}

say "$SYSTEM_NAME 51 CLI"
say "Anthropic Base URL: $ANTHROPIC_BASE_URL"
say "OpenAI Base URL: $OPENAI_BASE_URL"
say ""
say "请选择要配置的客户端："
say "1) Claude Code"
say "2) Codex CLI"
say "3) 两者都配置"
choice="$(ask "输入 1/2/3 [3]: ")"
choice="${choice:-3}"

api_key="$(ask_secret "请输入 $SYSTEM_NAME API Key: ")"
if [ -z "$api_key" ]; then
  say "API Key 不能为空。"
  exit 1
fi

case "$choice" in
  1)
    install_claude_if_needed
    configure_claude "$api_key"
    ;;
  2)
    install_codex_if_needed
    model="$(ask "请输入 Codex 使用的模型名称: ")"
    if [ -z "$model" ]; then
      say "模型名称不能为空。"
      exit 1
    fi
    configure_codex "$api_key" "$model"
    ;;
  3|"")
    install_claude_if_needed
    configure_claude "$api_key"
    install_codex_if_needed
    model="$(ask "请输入 Codex 使用的模型名称: ")"
    if [ -z "$model" ]; then
      say "模型名称不能为空。"
      exit 1
    fi
    configure_codex "$api_key" "$model"
    ;;
  *)
    say "无效选择：$choice"
    exit 1
    ;;
esac

say ""
say "配置完成。重新打开终端后可运行 claude 或 codex 验证。"
`

const installPowerShellTemplate = `$ErrorActionPreference = "Stop"

$SystemName = {{SYSTEM_NAME}}
$AnthropicBaseUrl = {{ANTHROPIC_BASE_URL}}
$OpenAIBaseUrl = {{OPENAI_BASE_URL}}

function Test-Command($Name) {
  return [bool](Get-Command $Name -ErrorAction SilentlyContinue)
}

function Backup-File($Path) {
  if (Test-Path $Path) {
    Copy-Item $Path "$Path.bak.$([DateTimeOffset]::UtcNow.ToUnixTimeSeconds())" -Force
  }
}

function Read-PlainSecret($Prompt) {
  $secure = Read-Host $Prompt -AsSecureString
  $bstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
  try {
    return [Runtime.InteropServices.Marshal]::PtrToStringBSTR($bstr)
  } finally {
    [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
  }
}

function ConvertTo-TomlString($Value) {
  return $Value.Replace('\', '\\').Replace('"', '\"')
}

function Install-NodeIfNeeded {
  if ((Test-Command node) -and (Test-Command npm)) {
    return
  }
  $answer = Read-Host "没有检测到 Node.js/npm。是否安装 Node.js LTS？[Y/n]"
  if ($answer -match "^(n|no)$") {
    return
  }
  if (Test-Command winget) {
    winget install OpenJS.NodeJS.LTS -e --source winget
  } elseif (Test-Command choco) {
    choco install nodejs-lts -y
  } else {
    Write-Host "未检测到 winget 或 Chocolatey。请从 https://nodejs.org 安装 Node.js LTS 后重试。"
  }
}

function Install-ClaudeIfNeeded {
  if (Test-Command claude) {
    Write-Host "已检测到 Claude Code。"
    return
  }
  Write-Host "正在安装 Claude Code..."
  if (Test-Command winget) {
    try {
      winget install Anthropic.ClaudeCode -e --source winget
    } catch {
      Write-Host "winget 安装 Claude Code 失败，尝试官方 PowerShell 安装脚本。"
    }
  }
  if (-not (Test-Command claude)) {
    try {
      Invoke-RestMethod https://claude.ai/install.ps1 | Invoke-Expression
    } catch {
      Write-Host "Claude Code 官方安装脚本失败，尝试 npm 安装。"
    }
  }
  if (-not (Test-Command claude)) {
    Install-NodeIfNeeded
    if (Test-Command npm) {
      npm install -g @anthropic-ai/claude-code
    } else {
      Write-Host "未找到 npm，无法自动安装 Claude Code。"
    }
  }
}

function Install-CodexIfNeeded {
  if (Test-Command codex) {
    Write-Host "已检测到 Codex CLI。"
    return
  }
  Write-Host "正在安装 Codex CLI..."
  try {
    Invoke-RestMethod https://chatgpt.com/codex/install.ps1 | Invoke-Expression
  } catch {
    Write-Host "Codex 官方安装脚本失败，尝试 npm 安装。"
  }
  if (-not (Test-Command codex)) {
    Install-NodeIfNeeded
    if (Test-Command npm) {
      npm install -g @openai/codex
    } else {
      Write-Host "未找到 npm，无法自动安装 Codex CLI。"
    }
  }
}

function Configure-Claude($ApiKey) {
  $dir = Join-Path $HOME ".claude"
  $settingsPath = Join-Path $dir "settings.json"
  New-Item -ItemType Directory -Path $dir -Force | Out-Null
  Backup-File $settingsPath
  $settings = [ordered]@{
    env = [ordered]@{
      ANTHROPIC_BASE_URL = $AnthropicBaseUrl
      ANTHROPIC_AUTH_TOKEN = $ApiKey
      CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC = "1"
    }
  }
  $settings | ConvertTo-Json -Depth 10 | Set-Content -Path $settingsPath -Encoding UTF8
  Write-Host "已写入 $settingsPath"
}

function Configure-Codex($ApiKey, $Model) {
  $dir = Join-Path $HOME ".codex"
  $configPath = Join-Path $dir "config.toml"
  $authPath = Join-Path $dir "auth.json"
  New-Item -ItemType Directory -Path $dir -Force | Out-Null
  Backup-File $configPath
  Backup-File $authPath
  $modelForToml = ConvertTo-TomlString $Model
  $systemNameForToml = ConvertTo-TomlString $SystemName
  $baseUrlForToml = ConvertTo-TomlString $OpenAIBaseUrl

  @"
model_provider = "51api"
model = "$modelForToml"
model_reasoning_effort = "high"
model_verbosity = "high"
disable_response_storage = true

[features]
web_search_request = true

[model_providers.51api]
name = "$systemNameForToml"
base_url = "$baseUrlForToml"
wire_api = "responses"
requires_openai_auth = true
"@ | Set-Content -Path $configPath -Encoding UTF8

  [ordered]@{ OPENAI_API_KEY = $ApiKey } | ConvertTo-Json -Depth 10 | Set-Content -Path $authPath -Encoding UTF8
  Write-Host "已写入 $configPath 和 $authPath"
}

Write-Host "$SystemName 51 CLI"
Write-Host "Anthropic Base URL: $AnthropicBaseUrl"
Write-Host "OpenAI Base URL: $OpenAIBaseUrl"
Write-Host ""
Write-Host "请选择要配置的客户端："
Write-Host "1) Claude Code"
Write-Host "2) Codex CLI"
Write-Host "3) 两者都配置"
$choice = Read-Host "输入 1/2/3 [3]"
if ([string]::IsNullOrWhiteSpace($choice)) { $choice = "3" }

$apiKey = Read-PlainSecret "请输入 $SystemName API Key"
if ([string]::IsNullOrWhiteSpace($apiKey)) {
  throw "API Key 不能为空。"
}

switch ($choice) {
  "1" {
    Install-ClaudeIfNeeded
    Configure-Claude $apiKey
  }
  "2" {
    Install-CodexIfNeeded
    $model = Read-Host "请输入 Codex 使用的模型名称"
    if ([string]::IsNullOrWhiteSpace($model)) { throw "模型名称不能为空。" }
    Configure-Codex $apiKey $model
  }
  "3" {
    Install-ClaudeIfNeeded
    Configure-Claude $apiKey
    Install-CodexIfNeeded
    $model = Read-Host "请输入 Codex 使用的模型名称"
    if ([string]::IsNullOrWhiteSpace($model)) { throw "模型名称不能为空。" }
    Configure-Codex $apiKey $model
  }
  default {
    throw "无效选择：$choice"
  }
}

Write-Host ""
Write-Host "配置完成。重新打开终端后可运行 claude 或 codex 验证。"
`
