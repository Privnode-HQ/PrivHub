package claude

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

func TestHandleClaudeResponseDataFallsBackWhenUsageMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service.InitTokenEncoders()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	info := &relaycommon.RelayInfo{
		PromptTokens:    17,
		OriginModelName: "claude-3-7-sonnet-20250219",
		RelayFormat:     types.RelayFormatClaude,
		RequestURLPath:  "/v1/messages",
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "claude-3-7-sonnet-20250219",
		},
	}
	claudeInfo := &ClaudeResponseInfo{Usage: &dto.Usage{}}

	body := []byte(`{"id":"msg_123","type":"message","role":"assistant","model":"claude-3-7-sonnet-20250219","content":[{"type":"text","text":"hello from claude"}],"stop_reason":"end_turn"}`)

	err := HandleClaudeResponseData(ctx, info, claudeInfo, nil, body, RequestModeMessage)
	if err != nil {
		t.Fatalf("HandleClaudeResponseData returned error: %v", err)
	}
	if claudeInfo.Usage == nil {
		t.Fatal("expected usage to be populated")
	}
	if claudeInfo.Usage.PromptTokens != info.PromptTokens {
		t.Fatalf("expected prompt tokens %d, got %d", info.PromptTokens, claudeInfo.Usage.PromptTokens)
	}
	if claudeInfo.Usage.CompletionTokens == 0 {
		t.Fatal("expected local completion token estimate when usage is missing")
	}
	if claudeInfo.Usage.TotalTokens != claudeInfo.Usage.PromptTokens+claudeInfo.Usage.CompletionTokens {
		t.Fatalf("unexpected total tokens: %+v", claudeInfo.Usage)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"id":"msg_123"`) {
		t.Fatalf("expected response body to be written, got %s", recorder.Body.String())
	}
}

func TestFormatClaudeResponseInfoAllowsMissingMessageUsage(t *testing.T) {
	claudeInfo := &ClaudeResponseInfo{Usage: &dto.Usage{}}
	response := &dto.ClaudeResponse{
		Type: "message_start",
		Message: &dto.ClaudeMediaMessage{
			Id:    "msg_456",
			Model: "claude-3-7-sonnet-20250219",
		},
	}

	ok := FormatClaudeResponseInfo(RequestModeMessage, response, nil, claudeInfo)
	if !ok {
		t.Fatal("expected message_start response to be accepted")
	}
	if claudeInfo.ResponseId != "msg_456" {
		t.Fatalf("expected response id to be copied, got %q", claudeInfo.ResponseId)
	}
	if claudeInfo.Model != "claude-3-7-sonnet-20250219" {
		t.Fatalf("expected model to be copied, got %q", claudeInfo.Model)
	}
	if claudeInfo.Usage.TotalTokens != 0 {
		t.Fatalf("expected usage to remain zero when upstream usage is missing, got %+v", claudeInfo.Usage)
	}
}
