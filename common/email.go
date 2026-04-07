package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

const (
	postmarkAPIBaseURL           = "https://api.postmarkapp.com"
	postmarkSendEmailEndpoint    = postmarkAPIBaseURL + "/email"
	postmarkBatchEmailEndpoint   = postmarkAPIBaseURL + "/email/batch"
	postmarkBulkEmailEndpoint    = postmarkAPIBaseURL + "/email/bulk"
	postmarkDefaultMessageStream = "outbound"
	postmarkDirectBatchLimit     = 50
	postmarkChunkSize            = 50
	postmarkChunkInterval        = 2 * time.Minute
	postmarkMaxRetryAttempts     = 3
)

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

type postmarkEmailRequest struct {
	From          string            `json:"From"`
	To            string            `json:"To"`
	Subject       string            `json:"Subject"`
	HtmlBody      string            `json:"HtmlBody,omitempty"`
	TextBody      string            `json:"TextBody,omitempty"`
	MessageStream string            `json:"MessageStream,omitempty"`
	Metadata      map[string]string `json:"Metadata,omitempty"`
}

type postmarkEmailResponse struct {
	MessageID string `json:"MessageID"`
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
}

type postmarkBatchResponseItem struct {
	To        string `json:"To"`
	MessageID string `json:"MessageID"`
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
}

type postmarkBulkRequest struct {
	From          string                `json:"From"`
	Subject       string                `json:"Subject"`
	HtmlBody      string                `json:"HtmlBody,omitempty"`
	TextBody      string                `json:"TextBody,omitempty"`
	MessageStream string                `json:"MessageStream,omitempty"`
	Messages      []postmarkBulkMessage `json:"Messages"`
}

type postmarkBulkMessage struct {
	To            string            `json:"To"`
	TemplateModel map[string]string `json:"TemplateModel,omitempty"`
	Metadata      map[string]string `json:"Metadata,omitempty"`
}

type postmarkBulkResponse struct {
	ID      string `json:"ID"`
	Status  string `json:"Status"`
	Message string `json:"Message"`
}

type postmarkAPIError struct {
	StatusCode int
	ErrorCode  int
	Message    string
	Body       string
}

func (e *postmarkAPIError) Error() string {
	if e == nil {
		return "Postmark 请求失败"
	}
	if e.ErrorCode > 0 {
		return fmt.Sprintf("Postmark 请求失败(%d/%d): %s", e.StatusCode, e.ErrorCode, e.Message)
	}
	return fmt.Sprintf("Postmark 请求失败(%d): %s", e.StatusCode, e.Message)
}

func splitEmailReceivers(receiver string) []string {
	fields := strings.FieldsFunc(receiver, func(r rune) bool {
		return r == ';' || r == ','
	})
	receivers := make([]string, 0, len(fields))
	for _, field := range fields {
		trimmed := strings.TrimSpace(field)
		if trimmed != "" {
			receivers = append(receivers, trimmed)
		}
	}
	return receivers
}

func htmlToPlainText(content string) string {
	replacer := strings.NewReplacer("<br>", "\n", "<br/>", "\n", "<br />", "\n", "</p>", "\n", "</div>", "\n", "</li>", "\n")
	text := replacer.Replace(content)
	text = htmlTagPattern.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = strings.TrimSpace(text)
	if text == "" {
		return "This email contains HTML content."
	}
	return text
}

func shortEmailKeyHash(values ...string) string {
	hasher := sha256.New()
	for _, value := range values {
		_, _ = hasher.Write([]byte(strings.TrimSpace(value)))
		_, _ = hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil)[:16])
}

func normalizeEmailIdempotencyKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if len(key) <= 256 {
		return key
	}
	return fmt.Sprintf("system-email/%s", shortEmailKeyHash(key))
}

func scopeEmailIdempotencyKey(baseKey string, recipient string) string {
	if strings.TrimSpace(baseKey) == "" {
		return ""
	}
	return normalizeEmailIdempotencyKey(fmt.Sprintf("%s/%s", baseKey, shortEmailKeyHash(recipient)))
}

func formatEmailSender(name string, address string) string {
	sender := mail.Address{
		Name:    strings.TrimSpace(name),
		Address: strings.TrimSpace(address),
	}
	return sender.String()
}

func GenerateEmailIdempotencyKey(eventType string, values ...string) string {
	eventType = strings.TrimSpace(eventType)
	if eventType == "" {
		eventType = "system-email"
	}
	return normalizeEmailIdempotencyKey(fmt.Sprintf("%s/%s", eventType, shortEmailKeyHash(values...)))
}

type EmailRecipientContext struct {
	Username      string
	RecipientName string
	CAHID         string
	Email         string
	PublishedAt   time.Time
}

type BatchEmailEntry struct {
	Recipient string
	Subject   string
	Content   string
	Context   EmailRecipientContext
}

type BatchEmailResult struct {
	Recipient string
	Success   bool
	Error     string
}

func SendEmail(subject string, receiver string, content string) error {
	return sendEmail(subject, receiver, content, "", EmailRecipientContext{})
}

func SendEmailWithIdempotencyKey(subject string, receiver string, content string, idempotencyKey string) error {
	return sendEmail(subject, receiver, content, idempotencyKey, EmailRecipientContext{})
}

func SendEmailWithContext(subject string, receiver string, content string, ctx EmailRecipientContext) error {
	return sendEmail(subject, receiver, content, "", ctx)
}

func SendEmailWithIdempotencyKeyAndContext(subject string, receiver string, content string, idempotencyKey string, ctx EmailRecipientContext) error {
	return sendEmail(subject, receiver, content, idempotencyKey, ctx)
}

func SendBatchEmailsWithIdempotencyKey(entries []BatchEmailEntry, idempotencyKey string) ([]BatchEmailResult, error) {
	if err := validatePostmarkConfiguration(); err != nil {
		return nil, err
	}
	from, err := currentPostmarkSender()
	if err != nil {
		return nil, err
	}
	return sendBatchEntriesWithStrategy(from, entries, idempotencyKey), nil
}

func sendEmail(subject string, receiver string, content string, idempotencyKey string, ctx EmailRecipientContext) error {
	if err := validatePostmarkConfiguration(); err != nil {
		return err
	}

	receivers := splitEmailReceivers(receiver)
	if len(receivers) == 0 {
		return fmt.Errorf("收件人邮箱未配置")
	}

	from, err := currentPostmarkSender()
	if err != nil {
		return err
	}

	if len(receivers) == 1 {
		recipientCtx := ctx
		recipientCtx.Email = receivers[0]
		request, err := buildPostmarkEmailRequest(from, subject, content, recipientCtx, scopeEmailIdempotencyKey(idempotencyKey, receivers[0]))
		if err != nil {
			return err
		}
		if _, err = sendPostmarkSingleWithRetry(*request); err != nil {
			err = fmt.Errorf("Postmark 邮件发送失败: %w", err)
			SysError(fmt.Sprintf("failed to send email to %s via Postmark: %v", receivers[0], err))
			return err
		}
		return nil
	}

	entries := make([]BatchEmailEntry, 0, len(receivers))
	for _, recipient := range receivers {
		recipientCtx := ctx
		recipientCtx.Email = recipient
		entries = append(entries, BatchEmailEntry{
			Recipient: recipient,
			Subject:   subject,
			Content:   content,
			Context:   recipientCtx,
		})
	}

	results := sendBatchEntriesWithStrategy(from, entries, idempotencyKey)
	for _, result := range results {
		if result.Success {
			continue
		}
		errorMessage := strings.TrimSpace(result.Error)
		if errorMessage == "" {
			errorMessage = "邮件发送失败"
		}
		err = fmt.Errorf("Postmark 邮件发送失败: %s", errorMessage)
		SysError(fmt.Sprintf("failed to send email to %s via Postmark: %v", result.Recipient, err))
		return err
	}
	return nil
}

func sendBatchEntriesWithStrategy(from string, entries []BatchEmailEntry, idempotencyKey string) []BatchEmailResult {
	results := initBatchEmailResults(entries)
	if len(entries) == 0 {
		return results
	}

	switch {
	case len(entries) == 1:
		return sendBatchChunk(from, entries, idempotencyKey)
	case len(entries) <= postmarkDirectBatchLimit:
		return sendBatchChunk(from, entries, idempotencyKey)
	case resolvePostmarkLargeBatchMode() == PostmarkLargeBatchModeBulk:
		return sendBulkBatch(from, entries, idempotencyKey)
	default:
		return sendChunkedBatches(from, entries, idempotencyKey)
	}
}

func sendChunkedBatches(from string, entries []BatchEmailEntry, idempotencyKey string) []BatchEmailResult {
	results := initBatchEmailResults(entries)
	for chunkStart := 0; chunkStart < len(entries); chunkStart += postmarkChunkSize {
		chunkEnd := chunkStart + postmarkChunkSize
		if chunkEnd > len(entries) {
			chunkEnd = len(entries)
		}

		chunkResults := sendBatchChunk(
			from,
			entries[chunkStart:chunkEnd],
			scopeEmailIdempotencyKey(idempotencyKey, fmt.Sprintf("chunk-%d", chunkStart)),
		)
		copy(results[chunkStart:chunkEnd], chunkResults)

		if chunkEnd < len(entries) {
			time.Sleep(postmarkChunkInterval)
		}
	}
	return results
}

func sendBatchChunk(from string, entries []BatchEmailEntry, idempotencyKey string) []BatchEmailResult {
	results := initBatchEmailResults(entries)
	if len(entries) == 0 {
		return results
	}

	if len(entries) == 1 {
		entry := entries[0]
		recipientCtx := entry.Context
		if strings.TrimSpace(entry.Recipient) != "" {
			recipientCtx.Email = strings.TrimSpace(entry.Recipient)
		}
		request, err := buildPostmarkEmailRequest(
			from,
			entry.Subject,
			entry.Content,
			recipientCtx,
			scopeEmailIdempotencyKey(idempotencyKey, entry.Recipient),
		)
		if err != nil {
			results[0].Error = err.Error()
			return results
		}
		if _, err = sendPostmarkSingleWithRetry(*request); err != nil {
			results[0].Error = err.Error()
			return results
		}
		results[0].Success = true
		results[0].Error = ""
		return results
	}

	requests := make([]postmarkEmailRequest, 0, len(entries))
	requestIndexToEntryIndex := make([]int, 0, len(entries))

	for entryIndex, entry := range entries {
		recipientCtx := entry.Context
		if strings.TrimSpace(entry.Recipient) != "" {
			recipientCtx.Email = strings.TrimSpace(entry.Recipient)
		}
		request, err := buildPostmarkEmailRequest(
			from,
			entry.Subject,
			entry.Content,
			recipientCtx,
			scopeEmailIdempotencyKey(idempotencyKey, entry.Recipient),
		)
		if err != nil {
			results[entryIndex].Error = err.Error()
			continue
		}
		requests = append(requests, *request)
		requestIndexToEntryIndex = append(requestIndexToEntryIndex, entryIndex)
	}

	if len(requests) == 0 {
		return results
	}

	responseItems, err := sendPostmarkBatchWithRetry(requests)
	if err != nil {
		for _, entryIndex := range requestIndexToEntryIndex {
			results[entryIndex].Error = err.Error()
		}
		return results
	}
	if len(responseItems) != len(requests) {
		err = fmt.Errorf("Postmark 批量发送返回数量异常: 期望 %d，实际 %d", len(requests), len(responseItems))
		for _, entryIndex := range requestIndexToEntryIndex {
			results[entryIndex].Error = err.Error()
		}
		return results
	}

	for requestIndex, entryIndex := range requestIndexToEntryIndex {
		item := responseItems[requestIndex]
		if item.ErrorCode == 0 {
			results[entryIndex].Success = true
			results[entryIndex].Error = ""
			continue
		}

		if isRetryablePostmarkResultCode(item.ErrorCode) {
			if _, retryErr := sendPostmarkSingleWithRetry(requests[requestIndex]); retryErr == nil {
				results[entryIndex].Success = true
				results[entryIndex].Error = ""
				continue
			} else {
				results[entryIndex].Error = retryErr.Error()
				continue
			}
		}

		errorMessage := strings.TrimSpace(item.Message)
		if errorMessage == "" {
			errorMessage = fmt.Sprintf("Postmark ErrorCode %d", item.ErrorCode)
		}
		results[entryIndex].Error = errorMessage
	}

	return results
}

func sendBulkBatch(from string, entries []BatchEmailEntry, idempotencyKey string) []BatchEmailResult {
	results := initBatchEmailResults(entries)
	messages := make([]postmarkBulkMessage, 0, len(entries))
	requestIndexToEntryIndex := make([]int, 0, len(entries))

	for entryIndex, entry := range entries {
		recipientCtx := entry.Context
		if strings.TrimSpace(entry.Recipient) != "" {
			recipientCtx.Email = strings.TrimSpace(entry.Recipient)
		}
		request, err := buildPostmarkEmailRequest(
			from,
			entry.Subject,
			entry.Content,
			recipientCtx,
			scopeEmailIdempotencyKey(idempotencyKey, entry.Recipient),
		)
		if err != nil {
			results[entryIndex].Error = err.Error()
			continue
		}

		messages = append(messages, postmarkBulkMessage{
			To: request.To,
			TemplateModel: map[string]string{
				"Subject":  request.Subject,
				"TextBody": request.TextBody,
				"HtmlBody": request.HtmlBody,
			},
			Metadata: request.Metadata,
		})
		requestIndexToEntryIndex = append(requestIndexToEntryIndex, entryIndex)
	}

	if len(messages) == 0 {
		return results
	}

	request := postmarkBulkRequest{
		From:          from,
		Subject:       "{{Subject}}",
		HtmlBody:      "{{{HtmlBody}}}",
		TextBody:      "{{TextBody}}",
		MessageStream: postmarkDefaultMessageStream,
		Messages:      messages,
	}

	response, err := sendPostmarkBulkWithRetry(request)
	if err != nil {
		for _, entryIndex := range requestIndexToEntryIndex {
			results[entryIndex].Error = err.Error()
		}
		return results
	}

	if !strings.EqualFold(strings.TrimSpace(response.Status), "Accepted") {
		errorMessage := strings.TrimSpace(response.Message)
		if errorMessage == "" {
			errorMessage = "Postmark bulk 请求未被接受"
		}
		for _, entryIndex := range requestIndexToEntryIndex {
			results[entryIndex].Error = errorMessage
		}
		return results
	}

	for _, entryIndex := range requestIndexToEntryIndex {
		results[entryIndex].Success = true
		results[entryIndex].Error = ""
	}

	return results
}

func buildPostmarkEmailRequest(from string, subject string, content string, ctx EmailRecipientContext, idempotencyKey string) (*postmarkEmailRequest, error) {
	recipientEmail := strings.TrimSpace(ctx.Email)
	if recipientEmail == "" {
		return nil, fmt.Errorf("收件人邮箱未配置")
	}

	renderedHTML, err := RenderMessageHTML(subject, content, MessageTemplateContext{
		RecipientName:  resolveEmailRecipientName(ctx),
		RecipientEmail: recipientEmail,
		PublishedAt:    ctx.PublishedAt,
	})
	if err != nil {
		return nil, err
	}

	request := &postmarkEmailRequest{
		From:          from,
		To:            recipientEmail,
		Subject:       formatEmailSubject(subject, ctx),
		HtmlBody:      renderedHTML,
		TextBody:      htmlToPlainText(renderedHTML),
		MessageStream: postmarkDefaultMessageStream,
	}
	if scopedKey := normalizeEmailIdempotencyKey(idempotencyKey); scopedKey != "" {
		request.Metadata = map[string]string{
			"idempotency_key": scopedKey,
		}
	}

	return request, nil
}

func sendPostmarkSingleWithRetry(request postmarkEmailRequest) (*postmarkEmailResponse, error) {
	response := &postmarkEmailResponse{}
	if err := sendPostmarkJSONWithRetry(postmarkSendEmailEndpoint, request, response); err != nil {
		return nil, err
	}
	if response.ErrorCode != 0 {
		return nil, &postmarkAPIError{
			StatusCode: http.StatusOK,
			ErrorCode:  response.ErrorCode,
			Message:    strings.TrimSpace(response.Message),
		}
	}
	if strings.TrimSpace(response.MessageID) == "" {
		return nil, fmt.Errorf("Postmark 返回空响应")
	}
	return response, nil
}

func sendPostmarkBatchWithRetry(requests []postmarkEmailRequest) ([]postmarkBatchResponseItem, error) {
	responseItems := make([]postmarkBatchResponseItem, 0, len(requests))
	if err := sendPostmarkJSONWithRetry(postmarkBatchEmailEndpoint, requests, &responseItems); err != nil {
		return nil, err
	}
	return responseItems, nil
}

func sendPostmarkBulkWithRetry(request postmarkBulkRequest) (*postmarkBulkResponse, error) {
	response := &postmarkBulkResponse{}
	if err := sendPostmarkJSONWithRetry(postmarkBulkEmailEndpoint, request, response); err != nil {
		return nil, err
	}
	return response, nil
}

func sendPostmarkJSONWithRetry(endpoint string, payload any, response any) error {
	body, err := Marshal(payload)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt < postmarkMaxRetryAttempts; attempt++ {
		responseBody, statusCode, reqErr := doPostmarkJSONRequest(endpoint, body)
		if reqErr == nil {
			if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
				if response == nil || len(bytes.TrimSpace(responseBody)) == 0 {
					return nil
				}
				if err = Unmarshal(responseBody, response); err != nil {
					return fmt.Errorf("解析 Postmark 响应失败: %w", err)
				}
				return nil
			}
			reqErr = newPostmarkAPIError(statusCode, responseBody)
		}

		lastErr = reqErr
		if attempt == postmarkMaxRetryAttempts-1 || !isRetryablePostmarkError(reqErr) {
			break
		}
		time.Sleep(calculatePostmarkRetryDelay(attempt))
	}

	return lastErr
}

func doPostmarkJSONRequest(endpoint string, body []byte) ([]byte, int, error) {
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Postmark-Server-Token", strings.TrimSpace(PostmarkServerToken))
	request.Header.Set("User-Agent", "PrivHub-Postmark/1.0")

	client := &http.Client{}
	if RelayTimeout > 0 {
		client.Timeout = time.Duration(RelayTimeout) * time.Second
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, response.StatusCode, err
	}
	return responseBody, response.StatusCode, nil
}

func newPostmarkAPIError(statusCode int, body []byte) error {
	apiErr := &postmarkAPIError{
		StatusCode: statusCode,
		Body:       strings.TrimSpace(string(body)),
		Message:    http.StatusText(statusCode),
	}

	var response struct {
		ErrorCode int    `json:"ErrorCode"`
		Message   string `json:"Message"`
	}
	if len(bytes.TrimSpace(body)) > 0 {
		if err := Unmarshal(body, &response); err == nil {
			apiErr.ErrorCode = response.ErrorCode
			if strings.TrimSpace(response.Message) != "" {
				apiErr.Message = strings.TrimSpace(response.Message)
			}
		}
	}

	if strings.TrimSpace(apiErr.Message) == "" {
		apiErr.Message = "Postmark 请求失败"
	}
	return apiErr
}

func validatePostmarkConfiguration() error {
	if strings.TrimSpace(PostmarkServerToken) == "" {
		return fmt.Errorf("Postmark Server Token 未配置")
	}
	if strings.TrimSpace(PostmarkSenderEmail) == "" {
		return fmt.Errorf("Postmark 发件人邮箱未配置")
	}
	return nil
}

func currentPostmarkSender() (string, error) {
	if err := validatePostmarkConfiguration(); err != nil {
		return "", err
	}

	senderName := strings.TrimSpace(PostmarkSenderName)
	if senderName == "" {
		senderName = SystemName
	}
	return formatEmailSender(senderName, PostmarkSenderEmail), nil
}

func resolvePostmarkLargeBatchMode() string {
	switch strings.TrimSpace(strings.ToLower(PostmarkLargeBatchMode)) {
	case PostmarkLargeBatchModeBulk:
		return PostmarkLargeBatchModeBulk
	default:
		return PostmarkLargeBatchModeChunked
	}
}

func initBatchEmailResults(entries []BatchEmailEntry) []BatchEmailResult {
	results := make([]BatchEmailResult, len(entries))
	for i, entry := range entries {
		results[i].Recipient = strings.TrimSpace(entry.Recipient)
	}
	return results
}

func isRetryablePostmarkError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *postmarkAPIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= http.StatusInternalServerError
	}

	lowerMessage := strings.ToLower(err.Error())
	retryableMarkers := []string{
		"timeout",
		"temporarily",
		"connection reset",
		"broken pipe",
		"connection refused",
	}
	for _, marker := range retryableMarkers {
		if strings.Contains(lowerMessage, marker) {
			return true
		}
	}
	return false
}

func isRetryablePostmarkResultCode(code int) bool {
	return code == http.StatusTooManyRequests || code >= http.StatusInternalServerError
}

func calculatePostmarkRetryDelay(attempt int) time.Duration {
	baseDelay := time.Second
	delay := baseDelay * time.Duration(1<<attempt)
	return delay + emailRetryJitter(attempt)
}

func emailRetryJitter(attempt int) time.Duration {
	raw := time.Now().UnixNano()%250 + int64(attempt*50)
	return time.Duration(raw) * time.Millisecond
}

func resolveEmailRecipientName(ctx EmailRecipientContext) string {
	recipientName := strings.TrimSpace(ctx.RecipientName)
	if recipientName != "" {
		return recipientName
	}

	username := strings.TrimSpace(ctx.Username)
	if username != "" {
		return username
	}

	cahID := strings.TrimSpace(strings.ToUpper(ctx.CAHID))
	if cahID != "" {
		return cahID
	}
	return "there"
}

func formatEmailSubject(subject string, ctx EmailRecipientContext) string {
	cahID := strings.TrimSpace(strings.ToUpper(ctx.CAHID))
	if cahID != "" {
		return fmt.Sprintf("[%s] %s", cahID, subject)
	}

	username := strings.TrimSpace(ctx.Username)
	if username == "" {
		username = "guest"
	}

	email := strings.TrimSpace(ctx.Email)
	if email == "" {
		return fmt.Sprintf("[%s] %s", username, subject)
	}
	return fmt.Sprintf("[%s %s] %s", username, email, subject)
}
