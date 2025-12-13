package service

import (
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
)

const SubscriptionQuotaSelectionTokenKey = "subscription_quota_selection_token"

// ShouldUseSubscriptionQuota returns true only for subscription users
// and only for the specified endpoints:
// - POST /v1/messages?beta=true (Claude)
// - POST /v1/responses (OpenAI Responses)
func ShouldUseSubscriptionQuota(relayInfo *relaycommon.RelayInfo) bool {
	if relayInfo == nil {
		return false
	}
	if relayInfo.UserGroup != "subscription" {
		return false
	}
	if relayInfo.IsClaudeBetaQuery {
		return true
	}
	return relayInfo.RelayMode == relayconstant.RelayModeResponses
}
