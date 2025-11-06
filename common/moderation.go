package common

import (
	"strings"
	"sync"
)

var (
	moderationGroupsMu  sync.RWMutex
	moderationGroupSet  = make(map[string]struct{})
	moderationGroupList []string
)

// SetModerationEnabledGroups parses the raw comma-separated group list
// and updates the in-memory set used for quick membership checks.
func SetModerationEnabledGroups(raw string) {
	moderationGroupsMu.Lock()
	defer moderationGroupsMu.Unlock()

	moderationGroupSet = make(map[string]struct{})
	moderationGroupList = make([]string, 0)
	if raw == "" {
		return
	}

	for _, part := range strings.Split(raw, ",") {
		group := strings.TrimSpace(part)
		if group == "" {
			continue
		}
		if _, exists := moderationGroupSet[group]; exists {
			continue
		}
		moderationGroupSet[group] = struct{}{}
		moderationGroupList = append(moderationGroupList, group)
	}
}

// GetModerationEnabledGroups returns a copy of the parsed group list for diagnostics.
func GetModerationEnabledGroups() []string {
	moderationGroupsMu.RLock()
	defer moderationGroupsMu.RUnlock()

	result := make([]string, len(moderationGroupList))
	copy(result, moderationGroupList)
	return result
}

// IsModerationEnabledForGroup reports whether content moderation must run for the provided group.
func IsModerationEnabledForGroup(group string) bool {
	group = strings.TrimSpace(group)
	if group == "" {
		return false
	}
	moderationGroupsMu.RLock()
	defer moderationGroupsMu.RUnlock()
	_, ok := moderationGroupSet[group]
	return ok
}
