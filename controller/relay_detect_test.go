package controller

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/types"
)

func TestShouldRejectDetectMagicString(t *testing.T) {
	cases := []struct {
		format types.RelayFormat
		want   bool
	}{
		{types.RelayFormatOpenAI, true},
		{types.RelayFormatOpenAIResponses, true},
		{types.RelayFormatClaude, true},
		{types.RelayFormatGemini, true},
		{types.RelayFormatEmbedding, false},
		{types.RelayFormatRerank, false},
		{types.RelayFormatOpenAIAudio, false},
		{types.RelayFormatOpenAIImage, false},
		{types.RelayFormatOpenAIRealtime, false},
	}

	for _, tc := range cases {
		if got := shouldRejectDetectMagicString(tc.format); got != tc.want {
			t.Fatalf("shouldRejectDetectMagicString(%q) = %v, want %v", tc.format, got, tc.want)
		}
	}
}

func TestRequestContainsDetectMagicString_CombineText(t *testing.T) {
	req := &dto.GeneralOpenAIRequest{
		Model: "gpt-4o-mini",
		Messages: []dto.Message{
			{Role: "user", Content: "hi " + constant.PrivnodeDetectMagicString},
		},
	}

	meta := req.GetTokenCountMeta()
	if !requestContainsDetectMagicString(req, meta) {
		t.Fatalf("expected magic string to be detected in CombineText")
	}
}

func TestRequestContainsDetectMagicString_MarshalFallback(t *testing.T) {
	req := &dto.GeminiChatRequest{
		Contents: []dto.GeminiChatContent{
			{Role: "user", Parts: []dto.GeminiPart{{Text: "hello"}}},
		},
		SystemInstructions: &dto.GeminiChatContent{
			Parts: []dto.GeminiPart{{Text: "sys " + constant.PrivnodeDetectMagicString}},
		},
	}

	meta := req.GetTokenCountMeta()
	if meta != nil && strings.Contains(meta.CombineText, constant.PrivnodeDetectMagicString) {
		t.Fatalf("test setup invalid: expected CombineText to not contain the magic string")
	}
	if !requestContainsDetectMagicString(req, meta) {
		t.Fatalf("expected magic string to be detected via marshal fallback")
	}
}
