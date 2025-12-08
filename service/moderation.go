package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/types"

	"gorm.io/gorm"
)

const (
	openAIModerationURL         = "https://api.oaipro.com/v1/moderations"
	openAIModerationModel       = "omni-moderation-latest"
	moderationChannelName       = "moderation-key"
	edenModerationURL           = "https://api.edenai.run/v2/text/moderation/"
	edenModerationProvider      = "openai"
	edenModerationChannelName   = "moderation-key--eden"
	edenLikelihoodFlagThreshold = 4.0
	edenScoreFlagThreshold      = 0.5
)

var (
	moderationChannelID     atomic.Int64
	edenModerationChannelID atomic.Int64
)

var errEdenModerationUnavailable = errors.New("eden moderation channel unavailable")

type openAIModerationRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type edenModerationRequest struct {
	ResponseAsDict       bool   `json:"response_as_dict"`
	AttributesAsList     bool   `json:"attributes_as_list"`
	ShowOriginalResponse bool   `json:"show_original_response"`
	Providers            string `json:"providers"`
	Language             string `json:"language"`
	Text                 string `json:"text"`
}

type edenProviderResponse struct {
	Items []edenModerationItem `json:"items"`
}

type edenModerationItem struct {
	Label           string          `json:"label"`
	Likelihood      json.RawMessage `json:"likelihood"`
	LikelihoodScore float64         `json:"likelihood_score"`
}

type openAIModerationResponse struct {
	Results []struct {
		Categories map[string]bool `json:"categories"`
	} `json:"results"`
}

type moderationDetails struct {
	CombinedText    string
	Messages        any
	LastUserMessage string
	Model           string
	RequestBody     string
	RequestDump     map[string]any
	Username        string
	Group           string
	UserID          int
	RequestID       string
}

// EnforceChatModeration sends the combined user prompt through OpenAI's moderation API
// when the requester belongs to one of the configured groups and the endpoint is chat related.
func EnforceChatModeration(c *gin.Context, relayMode int, relayFormat types.RelayFormat, request dto.Request, meta *types.TokenCountMeta) *types.NewAPIError {
	details := collectModerationDetails(c, request, meta)
	if !shouldRunModeration(&details, relayMode, relayFormat) {
		return nil
	}

	logCtx := buildLogContext(details.RequestID)
	reqCtx := c.Request.Context()
	if reqCtx == nil {
		reqCtx = context.Background()
	}

	blockedCategories, edenErr := tryEdenModeration(reqCtx, details.CombinedText)
	if edenErr != nil {
		if !errors.Is(edenErr, errEdenModerationUnavailable) {
			logger.LogWarn(logCtx, fmt.Sprintf("eden moderation request failed: %v", edenErr))
		}
		apiKey, err := getModerationAPIKey()
		if err != nil {
			logger.LogError(logCtx, fmt.Sprintf("failed to load moderation key: %v", err))
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

		resp, err := callOpenAIModeration(reqCtx, apiKey, details.CombinedText)
		if err != nil {
			logger.LogError(logCtx, fmt.Sprintf("openai moderation request failed: %v", err))
			return types.NewErrorWithStatusCode(
				err,
				types.ErrorCodePromptBlocked,
				http.StatusBadGateway,
				types.ErrOptionWithSkipRetry(),
			)
		}

		blockedCategories = extractTriggeredCategories(resp)
	}
	if len(blockedCategories) == 0 {
		return nil
	}

	sort.Strings(blockedCategories)
	reportModerationWebhook(&details, blockedCategories)
	message := fmt.Sprintf("内容审核未通过（%s）", strings.Join(blockedCategories, ", "))
	return types.NewErrorWithStatusCode(
		errors.New(message),
		types.ErrorCodePromptBlocked,
		http.StatusBadRequest,
		types.ErrOptionWithSkipRetry(),
	)
}

func shouldRunModeration(details *moderationDetails, relayMode int, relayFormat types.RelayFormat) bool {
	if details == nil {
		return false
	}
	if details.UserID == 1 {
		return false
	}
	if details.UserID == 206 {
		return false
	}
	if !common.IsModerationEnabledForGroup(details.Group) {
		return false
	}
	if strings.TrimSpace(details.CombinedText) == "" {
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

func tryEdenModeration(ctx context.Context, content string) ([]string, error) {
	apiKey, err := getEdenModerationAPIKey()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errEdenModerationUnavailable
		}
		return nil, err
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, errEdenModerationUnavailable
	}
	return callEdenModeration(ctx, apiKey, content)
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

func getEdenModerationAPIKey() (string, error) {
	if id := edenModerationChannelID.Load(); id > 0 {
		if channel, err := model.CacheGetChannel(int(id)); err == nil {
			if key := firstChannelKey(channel); key != "" {
				return key, nil
			}
			return "", errors.New("moderation channel key is empty")
		} else {
			edenModerationChannelID.Store(0)
		}
	}

	channel, err := model.GetFirstChannelByName(edenModerationChannelName)
	if err != nil {
		return "", err
	}
	edenModerationChannelID.Store(int64(channel.Id))
	if key := firstChannelKey(channel); key != "" {
		return key, nil
	}
	return "", errors.New("moderation channel key is empty")
}

func callOpenAIModeration(ctx context.Context, apiKey, content string) (*openAIModerationResponse, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		trimmed = content
	}
	body, err := json.Marshal(openAIModerationRequest{
		Input: truncateModerationText(trimmed),
		Model: openAIModerationModel,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIModerationURL, bytes.NewReader(body))
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
		return nil, fmt.Errorf("openai moderation failed with status %d: %s", resp.StatusCode, bodySnippet)
	}

	var moderationResp openAIModerationResponse
	if err := json.Unmarshal(respBody, &moderationResp); err != nil {
		return nil, err
	}
	return &moderationResp, nil
}

func callEdenModeration(ctx context.Context, apiKey, content string) ([]string, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		trimmed = content
	}
	body, err := json.Marshal(edenModerationRequest{
		ResponseAsDict:       true,
		AttributesAsList:     false,
		ShowOriginalResponse: false,
		Providers:            edenModerationProvider,
		Language:             "auto-detect",
		Text:                 truncateModerationText(trimmed),
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, edenModerationURL, bytes.NewReader(body))
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
		return nil, fmt.Errorf("eden moderation failed with status %d: %s", resp.StatusCode, bodySnippet)
	}

	triggered, err := findEdenTriggeredCategories(respBody)
	if err != nil {
		return nil, err
	}
	return triggered, nil
}

func findEdenTriggeredCategories(payload []byte) ([]string, error) {
	if len(payload) == 0 {
		return nil, errors.New("eden moderation response is empty")
	}
	var providers map[string]json.RawMessage
	if err := json.Unmarshal(payload, &providers); err != nil {
		return nil, err
	}
	triggeredSet := make(map[string]struct{})
	for _, raw := range providers {
		items, err := parseEdenProviderItems(raw)
		if err != nil {
			continue
		}
		for idx := range items {
			if isEdenItemFlagged(&items[idx]) {
				label := strings.TrimSpace(items[idx].Label)
				if label != "" {
					triggeredSet[label] = struct{}{}
				}
			}
		}
	}
	if len(triggeredSet) == 0 {
		return nil, nil
	}
	triggered := make([]string, 0, len(triggeredSet))
	for label := range triggeredSet {
		triggered = append(triggered, label)
	}
	return triggered, nil
}

func parseEdenProviderItems(raw json.RawMessage) ([]edenModerationItem, error) {
	if len(raw) == 0 {
		return nil, errors.New("eden provider payload is empty")
	}
	var provider edenProviderResponse
	if err := json.Unmarshal(raw, &provider); err != nil {
		return nil, err
	}
	return provider.Items, nil
}

func isEdenItemFlagged(item *edenModerationItem) bool {
	if item == nil {
		return false
	}
	likelihood := parseEdenLikelihood(item.Likelihood)
	if likelihood >= edenLikelihoodFlagThreshold {
		return true
	}
	if item.LikelihoodScore >= edenScoreFlagThreshold {
		return true
	}
	return false
}

func parseEdenLikelihood(raw json.RawMessage) float64 {
	if len(raw) == 0 {
		return 0
	}
	var numeric float64
	if err := json.Unmarshal(raw, &numeric); err == nil {
		return numeric
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		text = strings.TrimSpace(text)
		if text == "" {
			return 0
		}
		if value, err := strconv.ParseFloat(text, 64); err == nil {
			return value
		}
		switch strings.ToUpper(text) {
		case "VERY_UNLIKELY":
			return 1
		case "UNLIKELY":
			return 2
		case "POSSIBLE", "POSSIBLY":
			return 3
		case "LIKELY":
			return 4
		case "VERY_LIKELY":
			return 5
		}
	}
	return 0
}

func extractTriggeredCategories(resp *openAIModerationResponse) []string {
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

const moderationMaxTextLen = 8000

func truncateModerationText(text string) string {
	if len(text) <= moderationMaxTextLen {
		return text
	}
	runes := []rune(text)
	if len(runes) <= moderationMaxTextLen {
		return string(runes)
	}
	return string(runes[:moderationMaxTextLen])
}

func collectModerationDetails(c *gin.Context, request dto.Request, meta *types.TokenCountMeta) moderationDetails {
	details := moderationDetails{
		Username:  common.GetContextKeyString(c, constant.ContextKeyUserName),
		Group:     common.GetContextKeyString(c, constant.ContextKeyUsingGroup),
		UserID:    common.GetContextKeyInt(c, constant.ContextKeyUserId),
		Model:     common.GetContextKeyString(c, constant.ContextKeyOriginalModel),
		RequestID: c.GetString(common.RequestIdKey),
	}

	if meta != nil {
		details.CombinedText = meta.CombineText
	}

	if body, err := common.GetRequestBody(c); err == nil {
		details.RequestBody = string(body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	headers := make(map[string][]string, len(c.Request.Header))
	for k, v := range c.Request.Header {
		copied := make([]string, len(v))
		copy(copied, v)
		headers[k] = copied
	}

	requestDump := map[string]any{
		"method":      c.Request.Method,
		"url":         c.Request.URL.String(),
		"path":        c.Request.URL.Path,
		"query":       c.Request.URL.RawQuery,
		"headers":     headers,
		"client_ip":   c.ClientIP(),
		"remote_addr": c.Request.RemoteAddr,
	}
	if c.Request.Host != "" {
		requestDump["host"] = c.Request.Host
	}
	if details.RequestID != "" {
		requestDump["request_id"] = details.RequestID
	}
	details.RequestDump = requestDump

	details.Messages, details.LastUserMessage = extractMessagesFromRequest(request)
	if details.Model == "" {
		details.Model = deriveModelFromRequest(request)
	}

	if strings.TrimSpace(details.CombinedText) == "" {
		switch {
		case strings.TrimSpace(details.LastUserMessage) != "":
			details.CombinedText = details.LastUserMessage
		case strings.TrimSpace(details.RequestBody) != "":
			details.CombinedText = details.RequestBody
		}
	}

	return details
}

func extractMessagesFromRequest(request dto.Request) (any, string) {
	switch r := request.(type) {
	case *dto.GeneralOpenAIRequest:
		if len(r.Messages) == 0 {
			return nil, ""
		}
		return r.Messages, extractLastUserMessageFromOpenAI(r.Messages)
	case *dto.ClaudeRequest:
		if len(r.Messages) == 0 {
			return nil, ""
		}
		return r.Messages, extractLastUserMessageFromClaude(r.Messages)
	case *dto.OpenAIResponsesRequest:
		inputs := r.ParseInput()
		if len(inputs) == 0 {
			return nil, ""
		}
		return inputs, extractLastUserMessageFromResponses(inputs)
	default:
		return nil, ""
	}
}

func extractLastUserMessageFromOpenAI(messages []dto.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if !strings.EqualFold(messages[i].Role, "user") {
			continue
		}
		if text := joinOpenAIMessageText(&messages[i]); strings.TrimSpace(text) != "" {
			return text
		}
	}
	return ""
}

func joinOpenAIMessageText(message *dto.Message) string {
	if message == nil {
		return ""
	}
	content := message.ParseContent()
	if len(content) == 0 {
		if str, ok := message.Content.(string); ok {
			return str
		}
		return ""
	}

	var builder strings.Builder
	for _, item := range content {
		if item.Type == dto.ContentTypeText && strings.TrimSpace(item.Text) != "" {
			if builder.Len() > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString(item.Text)
		}
	}
	return builder.String()
}

func extractLastUserMessageFromClaude(messages []dto.ClaudeMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if !strings.EqualFold(messages[i].Role, "user") {
			continue
		}
		text := strings.TrimSpace(messages[i].GetStringContent())
		if text != "" {
			return text
		}
	}
	return ""
}

func extractLastUserMessageFromResponses(inputs []dto.MediaInput) string {
	for i := len(inputs) - 1; i >= 0; i-- {
		if strings.EqualFold(inputs[i].Type, "input_text") && strings.TrimSpace(inputs[i].Text) != "" {
			return inputs[i].Text
		}
	}
	return ""
}

func deriveModelFromRequest(request dto.Request) string {
	switch r := request.(type) {
	case *dto.GeneralOpenAIRequest:
		return r.Model
	case *dto.ClaudeRequest:
		return r.Model
	case *dto.OpenAIResponsesRequest:
		return r.Model
	default:
		return ""
	}
}

func reportModerationWebhook(details *moderationDetails, categories []string) {
	if details == nil {
		return
	}
	webhookURL := strings.TrimSpace(os.Getenv("MODERATION_WEBHOOK_URL"))
	if webhookURL == "" {
		return
	}

	payload := map[string]any{
		"timestamp":         time.Now().UTC().Format(time.RFC3339Nano),
		"username":          details.Username,
		"user_id":           details.UserID,
		"group":             details.Group,
		"model":             details.Model,
		"categories":        categories,
		"last_user_message": details.LastUserMessage,
		"messages":          details.Messages,
		"request_body":      details.RequestBody,
		"request":           details.RequestDump,
		"combined_text":     details.CombinedText,
	}
	if details.RequestID != "" {
		payload["request_id"] = details.RequestID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logCtx := buildLogContext(details.RequestID)
		logger.LogError(logCtx, fmt.Sprintf("failed to marshal moderation webhook payload: %v", err))
		return
	}

	gopool.Go(func() {
		logCtx := buildLogContext(details.RequestID)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
		if err != nil {
			logger.LogError(logCtx, fmt.Sprintf("failed to create moderation webhook request: %v", err))
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := GetHttpClient()
		if client == nil {
			client = http.DefaultClient
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.LogError(logCtx, fmt.Sprintf("moderation webhook request failed: %v", err))
			return
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		if resp.StatusCode >= http.StatusBadRequest {
			logger.LogWarn(logCtx, fmt.Sprintf("moderation webhook returned status %d", resp.StatusCode))
		}
	})
}

func buildLogContext(requestID string) context.Context {
	if requestID == "" {
		return context.Background()
	}
	return context.WithValue(context.Background(), common.RequestIdKey, requestID)
}
