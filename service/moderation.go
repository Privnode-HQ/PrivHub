package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/types"

	"gorm.io/gorm"
)

const (
	mistralModerationURL   = "https://api.mistral.ai/v1/chat/moderations"
	mistralModerationModel = "mistral-moderation-latest"
	moderationChannelName  = "moderation-key"
)

var moderationChannelID atomic.Int64

type mistralModerationInput struct {
	Content string `json:"content"`
}

type mistralModerationRequest struct {
	Input []mistralModerationInput `json:"input"`
	Model string                   `json:"model"`
}

type mistralModerationResponse struct {
	Results []struct {
		Categories map[string]bool `json:"categories"`
	} `json:"results"`
}

// EnforceChatModeration sends the combined user prompt through Mistral's moderation API
// when the requester belongs to one of the configured groups and the endpoint is chat related.
func EnforceChatModeration(ctx context.Context, group string, relayMode int, relayFormat types.RelayFormat, combinedText string) *types.NewAPIError {
	if !shouldRunModeration(group, relayMode, relayFormat, combinedText) {
		return nil
	}

	apiKey, err := getModerationAPIKey()
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("failed to load moderation key: %v", err))
		statusCode := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			statusCode = http.StatusServiceUnavailable
		}
		return types.NewErrorWithStatusCode(
			err,
			types.ErrorCodePromptBlocked,
			statusCode,
			types.ErrOptionWithSkipRetry(),
		)
	}

	resp, err := callMistralModeration(ctx, apiKey, combinedText)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("mistral moderation request failed: %v", err))
		return types.NewErrorWithStatusCode(
			err,
			types.ErrorCodePromptBlocked,
			http.StatusBadGateway,
			types.ErrOptionWithSkipRetry(),
		)
	}

	blockedCategories := extractTriggeredCategories(resp)
	if len(blockedCategories) == 0 {
		return nil
	}

	sort.Strings(blockedCategories)
	message := fmt.Sprintf("内容审核未通过（%s）", strings.Join(blockedCategories, ", "))
	return types.NewErrorWithStatusCode(
		errors.New(message),
		types.ErrorCodePromptBlocked,
		http.StatusBadRequest,
		types.ErrOptionWithSkipRetry(),
	)
}

func shouldRunModeration(group string, relayMode int, relayFormat types.RelayFormat, combinedText string) bool {
	if !common.IsModerationEnabledForGroup(group) {
		return false
	}
	if strings.TrimSpace(combinedText) == "" {
		return false
	}
	switch relayMode {
	case relayconstant.RelayModeChatCompletions, relayconstant.RelayModeCompletions, relayconstant.RelayModeResponses:
		return true
	}
	// Fallback for chat endpoints without relay mode metadata, e.g., Claude messages.
	if relayMode == relayconstant.RelayModeUnknown {
		if relayFormat == types.RelayFormatClaude {
			return true
		}
	}
	return false
}

func getModerationAPIKey() (string, error) {
	if id := moderationChannelID.Load(); id > 0 {
		if channel, err := model.CacheGetChannel(int(id)); err == nil {
			if key := firstChannelKey(channel); key != "" {
				return key, nil
			}
			return "", errors.New("moderation channel key is empty")
		} else {
			moderationChannelID.Store(0)
		}
	}

	channel, err := model.GetFirstChannelByName(moderationChannelName)
	if err != nil {
		return "", err
	}
	moderationChannelID.Store(int64(channel.Id))
	if key := firstChannelKey(channel); key != "" {
		return key, nil
	}
	return "", errors.New("moderation channel key is empty")
}

func firstChannelKey(channel *model.Channel) string {
	if channel == nil {
		return ""
	}
	keys := channel.GetKeys()
	if len(keys) > 0 {
		return strings.TrimSpace(keys[0])
	}
	return strings.TrimSpace(channel.Key)
}

func callMistralModeration(ctx context.Context, apiKey, content string) (*mistralModerationResponse, error) {
	body, err := json.Marshal(mistralModerationRequest{
		Input: []mistralModerationInput{{Content: content}},
		Model: mistralModerationModel,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, mistralModerationURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := GetHttpClient()
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		bodySnippet := string(respBody)
		if len(bodySnippet) > 512 {
			bodySnippet = bodySnippet[:512]
		}
		return nil, fmt.Errorf("mistral moderation failed with status %d: %s", resp.StatusCode, bodySnippet)
	}

	var moderationResp mistralModerationResponse
	if err := json.Unmarshal(respBody, &moderationResp); err != nil {
		return nil, err
	}
	return &moderationResp, nil
}

func extractTriggeredCategories(resp *mistralModerationResponse) []string {
	if resp == nil || len(resp.Results) == 0 {
		return nil
	}
	categories := resp.Results[0].Categories
	if len(categories) == 0 {
		return nil
	}
	triggered := make([]string, 0)
	for category, flagged := range categories {
		if flagged {
			triggered = append(triggered, category)
		}
	}
	return triggered
}
