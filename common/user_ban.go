package common

import "strings"

func UserBannedMessage(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "用户已被封禁"
	}
	reason = strings.ReplaceAll(reason, "\r", " ")
	reason = strings.ReplaceAll(reason, "\n", " ")
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "用户已被封禁"
	}
	return "用户已被封禁，原因：" + reason
}
