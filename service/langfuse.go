package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	langfuseDefaultBaseURL = "https://cloud.langfuse.com"
	langfuseMaxStringChars = 32768
)

var langfuseMissingConfigOnce sync.Once

type LangfuseCaptureRequest struct {
	UserID              int
	TokenID             int
	TokenName           string
	Group               string
	Model               string
	Method              string
	Path                string
	RequestID           string
	CaptureRate         float64
	StatusCode          int
	StartedAt           time.Time
	EndedAt             time.Time
	RequestBody         []byte
	ResponseBody        []byte
	RequestContentType  string
	ResponseContentType string
}

type langfuseConfig struct {
	PublicKey   string
	SecretKey   string
	BaseURL     string
	Environment string
}

func loadLangfuseConfig() (langfuseConfig, bool) {
	cfg := langfuseConfig{
		PublicKey:   strings.TrimSpace(os.Getenv("LANGFUSE_PUBLIC_KEY")),
		SecretKey:   strings.TrimSpace(os.Getenv("LANGFUSE_SECRET_KEY")),
		BaseURL:     strings.TrimSpace(os.Getenv("LANGFUSE_BASE_URL")),
		Environment: strings.TrimSpace(os.Getenv("LANGFUSE_TRACING_ENVIRONMENT")),
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = strings.TrimSpace(os.Getenv("LANGFUSE_HOST"))
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = langfuseDefaultBaseURL
	}
	if cfg.Environment == "" {
		cfg.Environment = strings.TrimSpace(os.Getenv("LANGFUSE_ENVIRONMENT"))
	}
	if cfg.Environment == "" {
		cfg.Environment = "production"
	}
	return cfg, cfg.PublicKey != "" && cfg.SecretKey != ""
}

func (cfg langfuseConfig) ingestionURL() string {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if strings.HasSuffix(baseURL, "/api/public") {
		return baseURL + "/ingestion"
	}
	return baseURL + "/api/public/ingestion"
}

func shouldSampleLangfuseCapture(rate float64) bool {
	if rate <= 0 {
		return false
	}
	if rate >= 1 {
		return true
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return false
	}
	return float64(n.Int64())/1_000_000 < rate
}

func CaptureRelayToLangfuse(capture LangfuseCaptureRequest) {
	if capture.CaptureRate <= 0 || capture.StatusCode >= http.StatusBadRequest {
		return
	}
	if !shouldSampleLangfuseCapture(capture.CaptureRate) {
		return
	}
	cfg, ok := loadLangfuseConfig()
	if !ok {
		langfuseMissingConfigOnce.Do(func() {
			common.SysLog("Langfuse capture skipped: LANGFUSE_PUBLIC_KEY or LANGFUSE_SECRET_KEY is not configured")
		})
		return
	}

	payload, err := buildLangfusePayload(cfg, capture)
	if err != nil {
		common.SysLog("Langfuse capture skipped: " + err.Error())
		return
	}

	gopool.Go(func() {
		if err := sendLangfusePayload(cfg, payload); err != nil {
			common.SysLog("Langfuse capture failed: " + err.Error())
		}
	})
}

func buildLangfusePayload(cfg langfuseConfig, capture LangfuseCaptureRequest) (map[string]any, error) {
	traceID := common.GetUUID()
	generationID := common.GetUUID()
	now := time.Now().UTC()
	startedAt := capture.StartedAt.UTC()
	if startedAt.IsZero() {
		startedAt = now
	}
	endedAt := capture.EndedAt.UTC()
	if endedAt.IsZero() {
		endedAt = now
	}
	input := extractLangfuseInput(capture.RequestBody, capture.RequestContentType)
	output := extractLangfuseOutput(capture.ResponseBody, capture.ResponseContentType)
	usageDetails := extractUsageDetails(capture.ResponseBody)
	metadata := map[string]any{
		"request_id":   capture.RequestID,
		"group":        capture.Group,
		"capture_rate": capture.CaptureRate,
		"token_id":     capture.TokenID,
		"token_name":   capture.TokenName,
		"method":       capture.Method,
		"path":         capture.Path,
		"status_code":  capture.StatusCode,
		"source":       "privhub",
	}
	name := "PrivHub relay"
	if capture.Model != "" {
		name = "PrivHub relay " + capture.Model
	}

	traceBody := map[string]any{
		"id":          traceID,
		"timestamp":   startedAt.Format(time.RFC3339Nano),
		"environment": cfg.Environment,
		"name":        name,
		"userId":      strconv.Itoa(capture.UserID),
		"input":       input,
		"output":      output,
		"metadata":    metadata,
		"tags":        []string{"privhub", "group:" + capture.Group},
	}
	generationBody := map[string]any{
		"id":          generationID,
		"traceId":     traceID,
		"name":        name,
		"startTime":   startedAt.Format(time.RFC3339Nano),
		"endTime":     endedAt.Format(time.RFC3339Nano),
		"model":       capture.Model,
		"input":       input,
		"output":      output,
		"metadata":    metadata,
		"environment": cfg.Environment,
	}
	if len(usageDetails) > 0 {
		generationBody["usageDetails"] = usageDetails
	}

	return map[string]any{
		"batch": []map[string]any{
			{
				"id":        common.GetUUID(),
				"timestamp": now.Format(time.RFC3339Nano),
				"type":      "trace-create",
				"body":      traceBody,
			},
			{
				"id":        common.GetUUID(),
				"timestamp": now.Format(time.RFC3339Nano),
				"type":      "generation-create",
				"body":      generationBody,
			},
		},
		"metadata": map[string]any{
			"sdk_name":        "privhub",
			"sdk_integration": "privhub-langfuse-capture",
		},
	}, nil
}

func sendLangfusePayload(cfg langfuseConfig, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.ingestionURL(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	auth := base64.StdEncoding.EncodeToString([]byte(cfg.PublicKey + ":" + cfg.SecretKey))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusMultiStatus {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

func extractLangfuseInput(body []byte, contentType string) any {
	if len(body) == 0 {
		return nil
	}
	if !isJSONContent(contentType, body) {
		return truncateString(string(body), langfuseMaxStringChars)
	}
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return truncateString(string(body), langfuseMaxStringChars)
	}
	if m, ok := value.(map[string]any); ok {
		input := map[string]any{}
		for _, key := range []string{"model", "messages", "prompt", "input", "contents", "system", "instructions", "tools"} {
			if v, exists := m[key]; exists {
				input[key] = v
			}
		}
		if len(input) > 0 {
			return input
		}
	}
	return value
}

func extractLangfuseOutput(body []byte, contentType string) any {
	if len(body) == 0 {
		return nil
	}
	if strings.Contains(contentType, "text/event-stream") || bytes.Contains(body, []byte("\ndata:")) || bytes.HasPrefix(bytes.TrimSpace(body), []byte("data:")) {
		if output := extractSSEOutput(body); output != "" {
			return output
		}
		return truncateString(string(body), langfuseMaxStringChars)
	}
	if !isJSONContent(contentType, body) {
		return truncateString(string(body), langfuseMaxStringChars)
	}
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return truncateString(string(body), langfuseMaxStringChars)
	}
	if output := extractOutputFromJSON(value); output != nil {
		return output
	}
	return value
}

func extractSSEOutput(body []byte) string {
	var builder strings.Builder
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		var value any
		if err := json.Unmarshal([]byte(payload), &value); err != nil {
			continue
		}
		if text := extractTextFromJSON(value); text != "" {
			builder.WriteString(text)
		}
		if builder.Len() >= langfuseMaxStringChars {
			return truncateString(builder.String(), langfuseMaxStringChars)
		}
	}
	return builder.String()
}

func extractOutputFromJSON(value any) any {
	if text := extractTextFromJSON(value); text != "" {
		return text
	}
	if m, ok := value.(map[string]any); ok {
		for _, key := range []string{"output", "data", "content", "completion"} {
			if v, exists := m[key]; exists {
				return v
			}
		}
	}
	return nil
}

func extractTextFromJSON(value any) string {
	switch v := value.(type) {
	case map[string]any:
		var parts []string
		if choices, ok := v["choices"].([]any); ok {
			for _, choice := range choices {
				if text := extractTextFromJSON(choice); text != "" {
					parts = append(parts, text)
				}
			}
		}
		if message, ok := v["message"]; ok {
			if text := extractTextFromJSON(message); text != "" {
				parts = append(parts, text)
			}
		}
		if delta, ok := v["delta"]; ok {
			if text := extractTextFromJSON(delta); text != "" {
				parts = append(parts, text)
			}
		}
		if candidates, ok := v["candidates"].([]any); ok {
			for _, candidate := range candidates {
				if text := extractTextFromJSON(candidate); text != "" {
					parts = append(parts, text)
				}
			}
		}
		if content, ok := v["content"]; ok {
			if text := extractContentText(content); text != "" {
				parts = append(parts, text)
			}
		}
		if partsValue, ok := v["parts"].([]any); ok {
			for _, part := range partsValue {
				if text := extractTextFromJSON(part); text != "" {
					parts = append(parts, text)
				}
			}
		}
		for _, key := range []string{"text", "output_text", "completion"} {
			if text, ok := v[key].(string); ok {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "")
	case []any:
		var parts []string
		for _, item := range v {
			if text := extractTextFromJSON(item); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "")
	case string:
		return v
	default:
		return ""
	}
}

func extractContentText(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var parts []string
		for _, item := range v {
			if text := extractTextFromJSON(item); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "")
	default:
		return extractTextFromJSON(v)
	}
}

func extractUsageDetails(body []byte) map[string]float64 {
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return nil
	}
	usage := findUsageMap(value)
	if len(usage) == 0 {
		return nil
	}
	out := make(map[string]float64)
	for key, raw := range usage {
		if n, ok := numberFromAny(raw); ok {
			out[key] = n
		}
	}
	return out
}

func findUsageMap(value any) map[string]any {
	switch v := value.(type) {
	case map[string]any:
		if usage, ok := v["usage"].(map[string]any); ok {
			return usage
		}
		for _, child := range v {
			if usage := findUsageMap(child); len(usage) > 0 {
				return usage
			}
		}
	case []any:
		for _, child := range v {
			if usage := findUsageMap(child); len(usage) > 0 {
				return usage
			}
		}
	}
	return nil
}

func numberFromAny(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		n, err := v.Float64()
		return n, err == nil
	default:
		return 0, false
	}
}

func isJSONContent(contentType string, body []byte) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "json") || json.Valid(body)
}

func truncateString(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit] + "...[truncated]"
}
