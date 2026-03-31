package channel

import (
	"net/http"
	"net/http/httptest"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

func newHeaderOverrideTestContext(headers map[string]string) *gin.Context {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	ctx.Request = req
	return ctx
}

func TestProcessHeaderOverride_PassthroughAndPlaceholders(t *testing.T) {
	ctx := newHeaderOverrideTestContext(map[string]string{
		"X-Client":        "client-value",
		"X-Extra":         "extra-value",
		"Authorization":   "Bearer user-token",
		"Accept-Encoding": "gzip",
		"Cookie":          "a=b",
	})

	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ApiKey: "channel-secret",
			HeadersOverride: map[string]interface{}{
				"*":             true,
				"X-Extra":       "override-extra",
				"X-From-Client": "{client_header:X-Client}",
				"Authorization": "Bearer {api_key}",
				"X-Missing":     "{client_header:X-Missing}",
			},
		},
	}

	got, err := processHeaderOverride(info, ctx)
	if err != nil {
		t.Fatalf("processHeaderOverride returned error: %v", err)
	}

	if got["x-client"] != "client-value" {
		t.Fatalf("expected wildcard passthrough for x-client, got %#v", got["x-client"])
	}
	if got["x-extra"] != "override-extra" {
		t.Fatalf("expected explicit override to win for x-extra, got %#v", got["x-extra"])
	}
	if got["x-from-client"] != "client-value" {
		t.Fatalf("expected client header placeholder to resolve, got %#v", got["x-from-client"])
	}
	if got["authorization"] != "Bearer channel-secret" {
		t.Fatalf("expected explicit authorization override, got %#v", got["authorization"])
	}
	if _, exists := got["cookie"]; exists {
		t.Fatalf("expected cookie to be excluded from passthrough, got %#v", got["cookie"])
	}
	if _, exists := got["accept-encoding"]; exists {
		t.Fatalf("expected accept-encoding to be excluded from passthrough, got %#v", got["accept-encoding"])
	}
	if _, exists := got["x-missing"]; exists {
		t.Fatalf("expected missing client header placeholder to be skipped, got %#v", got["x-missing"])
	}
}

func TestProcessHeaderOverride_RegexPassthrough(t *testing.T) {
	ctx := newHeaderOverrideTestContext(map[string]string{
		"X-Trace-Id": "trace-123",
		"X-Other":    "other-value",
	})

	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			HeadersOverride: map[string]interface{}{
				"regex:^X-Trace-": true,
			},
		},
	}

	got, err := processHeaderOverride(info, ctx)
	if err != nil {
		t.Fatalf("processHeaderOverride returned error: %v", err)
	}

	if got["x-trace-id"] != "trace-123" {
		t.Fatalf("expected regex passthrough for x-trace-id, got %#v", got["x-trace-id"])
	}
	if _, exists := got["x-other"]; exists {
		t.Fatalf("expected x-other to be excluded, got %#v", got["x-other"])
	}
}

func TestProcessHeaderOverride_InvalidRegex(t *testing.T) {
	ctx := newHeaderOverrideTestContext(nil)
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			HeadersOverride: map[string]interface{}{
				"re:(": true,
			},
		},
	}

	_, err := processHeaderOverride(info, ctx)
	if err == nil {
		t.Fatal("expected invalid regex error, got nil")
	}

	newAPIErr, ok := err.(*types.NewAPIError)
	if !ok {
		t.Fatalf("expected *types.NewAPIError, got %T", err)
	}
	if newAPIErr.GetErrorCode() != types.ErrorCodeChannelHeaderOverrideInvalid {
		t.Fatalf("expected error code %q, got %q", types.ErrorCodeChannelHeaderOverrideInvalid, newAPIErr.GetErrorCode())
	}
}

func TestApplyHeaderOverrideToRequest_SetsHost(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "https://example.com/v1/chat/completions", nil)

	applyHeaderOverrideToRequest(req, map[string]string{
		"host":  "upstream.example.com",
		"x-foo": "bar",
	})

	if req.Host != "upstream.example.com" {
		t.Fatalf("expected req.Host to be overridden, got %q", req.Host)
	}
	if req.Header.Get("X-Foo") != "bar" {
		t.Fatalf("expected X-Foo header to be set, got %q", req.Header.Get("X-Foo"))
	}
}
