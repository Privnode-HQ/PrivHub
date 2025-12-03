package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

const defaultTemporaryChannelDisableDuration = 5 * time.Minute
const maxTemporaryDisableReasonLength = 256

type temporaryDisabledChannel struct {
	expireAt time.Time
	reason   string
}

var (
	temporaryDisabledChannels   = make(map[int]temporaryDisabledChannel)
	temporaryDisabledChannelsMu sync.RWMutex
)

// TemporarilyDisableChannel marks a channel as unusable for the specified duration.
// Returning the expiration time allows callers to log or inspect the disable window.
func TemporarilyDisableChannel(channelID int, duration time.Duration, reason string) time.Time {
	if duration <= 0 {
		duration = defaultTemporaryChannelDisableDuration
	}
	if len(reason) > maxTemporaryDisableReasonLength {
		reason = reason[:maxTemporaryDisableReasonLength]
	}
	expireAt := time.Now().Add(duration)
	temporaryDisabledChannelsMu.Lock()
	temporaryDisabledChannels[channelID] = temporaryDisabledChannel{
		expireAt: expireAt,
		reason:   reason,
	}
	temporaryDisabledChannelsMu.Unlock()
	common.SysLog(fmt.Sprintf("channel #%d temporarily disabled until %s", channelID, expireAt.Format(time.RFC3339)))
	return expireAt
}

// IsChannelTemporarilyDisabled returns true if the channel is currently in the temporary disable window.
func IsChannelTemporarilyDisabled(channelID int) bool {
	_, _, ok := GetTemporaryDisabledChannelInfo(channelID)
	return ok
}

// GetTemporaryDisabledChannelInfo returns the expiration time and reason of the temporary disable entry.
func GetTemporaryDisabledChannelInfo(channelID int) (time.Time, string, bool) {
	temporaryDisabledChannelsMu.RLock()
	entry, ok := temporaryDisabledChannels[channelID]
	temporaryDisabledChannelsMu.RUnlock()
	if !ok {
		return time.Time{}, "", false
	}
	if time.Now().After(entry.expireAt) {
		temporaryDisabledChannelsMu.Lock()
		delete(temporaryDisabledChannels, channelID)
		temporaryDisabledChannelsMu.Unlock()
		return time.Time{}, "", false
	}
	return entry.expireAt, entry.reason, true
}
