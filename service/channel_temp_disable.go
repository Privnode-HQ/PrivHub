package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

const defaultTemporaryChannelDisableDuration = 5 * time.Minute
const maxTemporaryDisableReasonLength = 256
const temporaryDisableCleanupInterval = 1 * time.Minute

type temporaryDisabledChannel struct {
	expireAt time.Time
	reason   string
}

var (
	temporaryDisabledChannels      = make(map[int]temporaryDisabledChannel)
	temporaryDisabledChannelsMu    sync.RWMutex
	temporaryDisabledCleanupOnce   sync.Once
	temporaryDisabledCleanupTicker *time.Ticker
)

// TemporarilyDisableChannel marks a channel as unusable for the specified duration.
// Returning the expiration time allows callers to log or inspect the disable window.
func TemporarilyDisableChannel(channelID int, duration time.Duration, reason string) time.Time {
	startTemporaryDisabledCleanupLoop()
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
	temporaryDisabledChannelsMu.Lock()
	entry, ok := temporaryDisabledChannels[channelID]
	if !ok {
		temporaryDisabledChannelsMu.Unlock()
		return time.Time{}, "", false
	}
	if time.Now().After(entry.expireAt) {
		delete(temporaryDisabledChannels, channelID)
		temporaryDisabledChannelsMu.Unlock()
		return time.Time{}, "", false
	}
	temporaryDisabledChannelsMu.Unlock()
	return entry.expireAt, entry.reason, true
}

func startTemporaryDisabledCleanupLoop() {
	temporaryDisabledCleanupOnce.Do(func() {
		temporaryDisabledCleanupTicker = time.NewTicker(temporaryDisableCleanupInterval)
		go func() {
			for range temporaryDisabledCleanupTicker.C {
				cleanupExpiredTemporaryDisabledChannels()
			}
		}()
	})
}

func cleanupExpiredTemporaryDisabledChannels() {
	now := time.Now()
	temporaryDisabledChannelsMu.Lock()
	for channelID, entry := range temporaryDisabledChannels {
		if now.After(entry.expireAt) {
			delete(temporaryDisabledChannels, channelID)
		}
	}
	temporaryDisabledChannelsMu.Unlock()
}
