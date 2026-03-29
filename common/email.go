package common

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"net/mail"
	"regexp"
	"strings"

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

func SendEmail(subject string, receiver string, content string) error {
	return sendEmail(subject, receiver, content, "")
}

func SendEmailWithIdempotencyKey(subject string, receiver string, content string, idempotencyKey string) error {
	return sendEmail(subject, receiver, content, idempotencyKey)
}

func sendEmail(subject string, receiver string, content string, idempotencyKey string) error {
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
	plainTextContent := htmlToPlainText(content)
	client := resend.NewClient(ResendAPIKey)

	for _, recipient := range receivers {
		message := &resend.SendEmailRequest{
			From:    from,
			To:      []string{recipient},
			Subject: subject,
			Html:    content,
			Text:    plainTextContent,
		}
		options := &resend.SendEmailOptions{}
		if scopedKey := scopeEmailIdempotencyKey(idempotencyKey, recipient); scopedKey != "" {
			options.IdempotencyKey = scopedKey
		}
		response, err := client.Emails.SendWithOptions(context.Background(), message, options)
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
