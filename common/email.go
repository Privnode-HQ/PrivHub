package common

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/resend/resend-go/v3"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

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
	Username    string
	Email       string
	PublishedAt time.Time
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
	if ResendAPIKey == "" {
		return nil, fmt.Errorf("Resend API Key 未配置")
	}
	if ResendSenderEmail == "" {
		return nil, fmt.Errorf("Resend 发件人邮箱未配置")
	}

	senderName := strings.TrimSpace(ResendSenderName)
	if senderName == "" {
		senderName = SystemName
	}
	from := formatEmailSender(senderName, ResendSenderEmail)
	client := resend.NewClient(ResendAPIKey)

	results := make([]BatchEmailResult, len(entries))
	for i, entry := range entries {
		results[i].Recipient = strings.TrimSpace(entry.Recipient)
	}

	for chunkStart := 0; chunkStart < len(entries); chunkStart += 100 {
		chunkEnd := chunkStart + 100
		if chunkEnd > len(entries) {
			chunkEnd = len(entries)
		}
		chunkResults, err := sendBatchChunk(
			client,
			from,
			entries[chunkStart:chunkEnd],
			scopeEmailIdempotencyKey(idempotencyKey, fmt.Sprintf("batch-%d", chunkStart)),
		)
		if err != nil {
			return nil, err
		}
		copy(results[chunkStart:chunkEnd], chunkResults)
	}
	return results, nil
}

func sendEmail(subject string, receiver string, content string, idempotencyKey string, ctx EmailRecipientContext) error {
	if ResendAPIKey == "" {
		return fmt.Errorf("Resend API Key 未配置")
	}
	if ResendSenderEmail == "" {
		return fmt.Errorf("Resend 发件人邮箱未配置")
	}

	receivers := splitEmailReceivers(receiver)
	if len(receivers) == 0 {
		return fmt.Errorf("收件人邮箱未配置")
	}

	senderName := strings.TrimSpace(ResendSenderName)
	if senderName == "" {
		senderName = SystemName
	}
	from := formatEmailSender(senderName, ResendSenderEmail)
	client := resend.NewClient(ResendAPIKey)

	for _, recipient := range receivers {
		recipientCtx := ctx
		if strings.TrimSpace(recipientCtx.Email) == "" {
			recipientCtx.Email = recipient
		}

		request, err := buildEmailRequest(from, subject, content, recipientCtx)
		if err != nil {
			return err
		}

		options := &resend.SendEmailOptions{}
		if scopedKey := scopeEmailIdempotencyKey(idempotencyKey, recipient); scopedKey != "" {
			options.IdempotencyKey = scopedKey
		}
		response, err := client.Emails.SendWithOptions(context.Background(), request, options)
		if err != nil {
			err = fmt.Errorf("Resend 邮件发送失败: %w", err)
			SysError(fmt.Sprintf("failed to send email to %s via Resend: %v", recipient, err))
			return err
		}
		if response == nil || response.Id == "" {
			err = fmt.Errorf("Resend 返回空响应")
			SysError(fmt.Sprintf("failed to send email to %s via Resend: %v", recipient, err))
			return err
		}
	}

	return nil
}

func sendBatchChunk(client *resend.Client, from string, entries []BatchEmailEntry, idempotencyKey string) ([]BatchEmailResult, error) {
	results := make([]BatchEmailResult, len(entries))
	requests := make([]*resend.SendEmailRequest, 0, len(entries))
	requestIndexToEntryIndex := make([]int, 0, len(entries))

	for entryIndex, entry := range entries {
		results[entryIndex].Recipient = strings.TrimSpace(entry.Recipient)

		recipientCtx := entry.Context
		if strings.TrimSpace(recipientCtx.Email) == "" {
			recipientCtx.Email = entry.Recipient
		}
		if strings.TrimSpace(recipientCtx.Email) == "" {
			results[entryIndex].Error = "收件人邮箱未配置"
			continue
		}

		request, err := buildEmailRequest(from, entry.Subject, entry.Content, recipientCtx)
		if err != nil {
			results[entryIndex].Error = err.Error()
			continue
		}
		requests = append(requests, request)
		requestIndexToEntryIndex = append(requestIndexToEntryIndex, entryIndex)
	}

	if len(requests) == 0 {
		return results, nil
	}

	options := &resend.BatchSendEmailOptions{
		IdempotencyKey:  idempotencyKey,
		BatchValidation: resend.BatchValidationPermissive,
	}
	response, err := sendBatchWithRetry(client, requests, options)
	if err != nil {
		for _, entryIndex := range requestIndexToEntryIndex {
			results[entryIndex].Error = err.Error()
		}
		return results, nil
	}

	failedRequestIndexes := make(map[int]string, len(response.Errors))
	for _, batchError := range response.Errors {
		failedRequestIndexes[batchError.Index] = batchError.Message
	}

	for requestIndex, entryIndex := range requestIndexToEntryIndex {
		if errMessage, exists := failedRequestIndexes[requestIndex]; exists {
			results[entryIndex].Error = errMessage
			continue
		}
		results[entryIndex].Success = true
		results[entryIndex].Error = ""
	}

	return results, nil
}

func sendBatchWithRetry(client *resend.Client, requests []*resend.SendEmailRequest, options *resend.BatchSendEmailOptions) (*resend.BatchEmailResponse, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		response, err := client.Batch.SendWithOptions(context.Background(), requests, options)
		if err == nil {
			return response, nil
		}
		lastErr = err
		if attempt == 2 || !isRetryableResendError(err) {
			break
		}
		time.Sleep(calculateBatchRetryDelay(err, attempt))
	}
	return nil, lastErr
}

func buildEmailRequest(from string, subject string, content string, ctx EmailRecipientContext) (*resend.SendEmailRequest, error) {
	renderedHTML, err := RenderMessageHTML(subject, content, MessageTemplateContext{
		RecipientName:  resolveEmailRecipientName(ctx.Username),
		RecipientEmail: ctx.Email,
		PublishedAt:    ctx.PublishedAt,
	})
	if err != nil {
		return nil, err
	}

	return &resend.SendEmailRequest{
		From:    from,
		To:      []string{ctx.Email},
		Subject: formatEmailSubject(subject, ctx),
		Html:    renderedHTML,
		Text:    htmlToPlainText(renderedHTML),
	}, nil
}

func isRetryableResendError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, resend.ErrRateLimit) {
		return true
	}

	lowerMessage := strings.ToLower(err.Error())
	retryableMarkers := []string{
		"too many requests",
		"internal server error",
		"service unavailable",
		"bad gateway",
		"gateway timeout",
		"timeout",
		"temporarily",
	}
	for _, marker := range retryableMarkers {
		if strings.Contains(lowerMessage, marker) {
			return true
		}
	}
	return false
}

func calculateBatchRetryDelay(err error, attempt int) time.Duration {
	var rateLimitErr *resend.RateLimitError
	if errors.As(err, &rateLimitErr) {
		if retryAfter, parseErr := strconv.Atoi(rateLimitErr.RetryAfter); parseErr == nil && retryAfter > 0 {
			return time.Duration(retryAfter)*time.Second + batchRetryJitter(attempt)
		}
	}

	baseDelay := time.Second
	delay := baseDelay * time.Duration(1<<attempt)
	return delay + batchRetryJitter(attempt)
}

func batchRetryJitter(attempt int) time.Duration {
	raw := time.Now().UnixNano()%250 + int64(attempt*50)
	return time.Duration(raw) * time.Millisecond
}

func resolveEmailRecipientName(username string) string {
	username = strings.TrimSpace(username)
	if username != "" {
		return username
	}
	return "there"
}

func formatEmailSubject(subject string, ctx EmailRecipientContext) string {
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
