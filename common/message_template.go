package common

import (
	"bytes"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
)

const MessageTemplateOptionKey = "MessageTemplateHTML"

const defaultMessageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Title}}</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#f1f1f1;font-family:'Inter',-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif}
.wrap{max-width:600px;margin:36px auto;background:#fff;overflow:hidden}
.logo{font-size:26px;font-weight:700;color:#fff;letter-spacing:-1px;line-height:1}
.sub{font-size:12px;color:rgba(255,255,255,.55);margin-top:4px}
.hd{background:#0D126A;display:flex;align-items:stretch;justify-content:space-between;padding:0}
.hd>div{padding:22px 0 22px 28px;display:flex;flex-direction:column;justify-content:center}
.login-btn{display:flex;align-items:center;justify-content:center;padding:0 18px;background:rgba(255,255,255,.14);color:#fff;text-decoration:none;font-size:15px;font-weight:300;border-radius:0}
.body{padding:26px 28px 22px}
.greeting{font-size:13.5px;color:#222;margin-bottom:18px}
.greeting strong{color:#0D126A;font-weight:600}
.content{font-size:13.5px;line-height:1.72;color:#333}
.content h1{font-size:19px;font-weight:700;color:#0D126A;margin:0 0 10px;letter-spacing:-.3px}
.content h2{font-size:15px;font-weight:600;color:#111;margin:16px 0 7px}
.content h3{font-size:13.5px;font-weight:600;color:#222;margin:12px 0 5px}
.content p{margin:0 0 11px}
.content ul,.content ol{padding-left:20px;margin:0 0 11px}
.content li{margin-bottom:4px}
.content a{color:#0D126A;text-decoration:none;border-bottom:1px solid rgba(13,18,106,.28)}
.content strong{font-weight:600;color:#111}
.content code{font-family:monospace;font-size:12px;background:#eef0fa;color:#0D126A;padding:1px 5px;border-radius:3px}
.content pre{background:#0f1120;color:#e2e8f0;padding:14px;border-radius:4px;overflow-x:auto;margin:0 0 11px;font-family:monospace;font-size:12px;line-height:1.6}
.content pre code{background:none;color:inherit;padding:0}
.content blockquote{border-left:2px solid #0D126A;margin:0 0 11px;padding:7px 13px;background:#f4f5fb;color:#555;border-radius:0 3px 3px 0}
.content hr{border:none;border-top:1px solid #eee;margin:16px 0}
.content table{width:100%;border-collapse:collapse;margin:0 0 11px;font-size:13px}
.content th{background:#0D126A;color:#fff;padding:7px 11px;text-align:left;font-weight:500;font-size:12.5px}
.content td{padding:7px 11px;border-bottom:1px solid #eee}
.content tr:nth-child(even) td{background:#f8f9fc}
.ft{background:#0D126A;padding:18px 28px}
.notice{font-size:11px;color:rgba(255,255,255,.5);margin-bottom:12px;line-height:1.6}
.ft-bottom{display:flex;align-items:center;justify-content:space-between;gap:12px;flex-wrap:wrap}
.copy{font-size:11px;color:rgba(255,255,255,.45)}
.links{display:flex;align-items:center;gap:14px}
.links a{font-size:11px;color:rgba(255,255,255,.75);text-decoration:none;font-weight:500}
</style>
</head>
<body>
<div class="wrap">
  <div class="hd">
    <div>
      <div class="logo">{{.SystemName}}</div>
    </div>
    <a class="login-btn" href="{{.DashboardURL}}">Dash</a>
  </div>
  <div class="body">
    <div class="greeting">Hello, <strong>{{.RecipientName}}</strong>.</div>
    <div class="content">{{.ContentHTML}}</div>
  </div>
  <div class="ft">
    <div class="notice">{{.NoticeText}}</div>
    <div class="ft-bottom">
      <span class="copy">Copyright {{.SystemName}} {{.CurrentYear}}, all rights reserved.</span>
      <div class="links">
        <a href="{{.TermsURL}}">Terms of Service</a>
        <a href="{{.PrivacyURL}}">Privacy Policy</a>
      </div>
    </div>
  </div>
</div>
</body>
</html>
`

type MessageTemplateContext struct {
	RecipientName  string
	RecipientEmail string
	PublishedAt    time.Time
}

type messageTemplateData struct {
	Title          string
	SystemName     string
	RecipientName  string
	RecipientEmail string
	PublishedAt    string
	ContentHTML    htmltemplate.HTML
	DashboardURL   string
	NoticeText     string
	CurrentYear    int
	TermsURL       string
	PrivacyURL     string
}

func GetMessageTemplateHTML() string {
	OptionMapRWMutex.RLock()
	defer OptionMapRWMutex.RUnlock()

	if templateHTML := strings.TrimSpace(OptionMap[MessageTemplateOptionKey]); templateHTML != "" {
		return templateHTML
	}
	return defaultMessageTemplate
}

func ValidateMessageTemplate(templateHTML string) error {
	templateHTML = strings.TrimSpace(templateHTML)
	if templateHTML == "" {
		return errors.New("消息模板不能为空")
	}
	if !strings.Contains(templateHTML, ".ContentHTML") {
		return errors.New("消息模板必须包含 {{.ContentHTML}} 占位符")
	}
	_, err := htmltemplate.New("message_template").Parse(templateHTML)
	if err != nil {
		return fmt.Errorf("消息模板格式错误: %w", err)
	}
	return nil
}

func RenderMessageHTML(title string, markdownContent string, ctx MessageTemplateContext) (string, error) {
	if err := ValidateMessageTemplate(GetMessageTemplateHTML()); err != nil {
		return "", err
	}

	contentHTML := markdown.ToHTML([]byte(strings.TrimSpace(markdownContent)), nil, nil)
	recipientName := strings.TrimSpace(ctx.RecipientName)
	if recipientName == "" {
		recipientName = "there"
	}
	systemName := strings.TrimSpace(SystemName)
	if systemName == "" {
		systemName = "Privnode"
	}

	publishedAt := ctx.PublishedAt
	if publishedAt.IsZero() {
		publishedAt = time.Now()
	}

	data := messageTemplateData{
		Title:          strings.TrimSpace(title),
		SystemName:     systemName,
		RecipientName:  recipientName,
		RecipientEmail: strings.TrimSpace(ctx.RecipientEmail),
		PublishedAt:    publishedAt.Format("2006-01-02 15:04:05"),
		ContentHTML:    htmltemplate.HTML(contentHTML),
		DashboardURL:   getMessageDashboardURL(),
		NoticeText:     fmt.Sprintf("This message has been sent to inform you of important updates to your %s products and services.", systemName),
		CurrentYear:    time.Now().Year(),
		TermsURL:       "https://legal.privnode.com/view?path=privacy-policy",
		PrivacyURL:     "https://legal.privnode.com/view?path=terms",
	}

	tpl, err := htmltemplate.New("message_template").Parse(GetMessageTemplateHTML())
	if err != nil {
		return "", fmt.Errorf("解析消息模板失败: %w", err)
	}

	var buf bytes.Buffer
	if err = tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染消息模板失败: %w", err)
	}
	return buf.String(), nil
}

func getMessageDashboardURL() string {
	OptionMapRWMutex.RLock()
	serverAddress := strings.TrimSpace(OptionMap["ServerAddress"])
	OptionMapRWMutex.RUnlock()

	if serverAddress == "" {
		serverAddress = "https://privnode.com"
	}
	return strings.TrimRight(serverAddress, "/") + "/login"
}
