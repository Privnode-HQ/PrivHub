package common

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	sendgridmail "github.com/sendgrid/sendgrid-go/helpers/mail"
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

func SendEmail(subject string, receiver string, content string) error {
	if SendGridAPIKey == "" {
		return fmt.Errorf("SendGrid API Key 未配置")
	}
	if SendGridSenderEmail == "" {
		return fmt.Errorf("SendGrid 发件人邮箱未配置")
	}

	receivers := splitEmailReceivers(receiver)
	if len(receivers) == 0 {
		return fmt.Errorf("收件人邮箱未配置")
	}

	senderName := strings.TrimSpace(SendGridSenderName)
	if senderName == "" {
		senderName = SystemName
	}
	from := sendgridmail.NewEmail(senderName, SendGridSenderEmail)
	plainTextContent := htmlToPlainText(content)
	client := sendgrid.NewSendClient(SendGridAPIKey)

	for _, recipient := range receivers {
		to := sendgridmail.NewEmail("", recipient)
		message := sendgridmail.NewSingleEmail(from, subject, to, plainTextContent, content)
		response, err := client.Send(message)
		if err != nil {
			SysError(fmt.Sprintf("failed to send email to %s via SendGrid: %v", recipient, err))
			return err
		}
		if response == nil {
			err = fmt.Errorf("SendGrid 返回空响应")
			SysError(fmt.Sprintf("failed to send email to %s via SendGrid: %v", recipient, err))
			return err
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			err = fmt.Errorf("SendGrid 邮件发送失败，状态码 %d: %s", response.StatusCode, response.Body)
			SysError(fmt.Sprintf("failed to send email to %s via SendGrid: %v", recipient, err))
			return err
		}
	}

	return nil
}
