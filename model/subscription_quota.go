package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrSubscriptionQuotaExhausted = errors.New("SUBSCRIPTION_QUOTA_EXHAUSTED")

type SubscriptionLimit struct {
	Total     int64 `json:"total"`
	Available int64 `json:"available"`
	ResetAt   int64 `json:"reset_at"`
}

type SubscriptionDuration struct {
	StartAt          int64 `json:"start_at"`
	EndAt            int64 `json:"end_at"`
	AutoRenewEnabled bool  `json:"auto_renew_enabled"`
}

type SubscriptionItem struct {
	PlanName       string               `json:"plan_name"`
	PlanID         string               `json:"plan_id"`
	SubscriptionID string               `json:"subscription_id"`
	Limit5H        SubscriptionLimit    `json:"5h_limit"`
	Limit7D        SubscriptionLimit    `json:"7d_limit"`
	Duration       SubscriptionDuration `json:"duration"`
	Owner          int64                `json:"owner"`
	Status         string               `json:"status"`
}

const (
	subscriptionSelectionPrefixSubscriptionID = "sub:"
	subscriptionSelectionPrefixPlanID         = "plan:"
	subscriptionSelectionPrefixIndex          = "idx:"
)

func encodeSubscriptionSelectionToken(item SubscriptionItem, index int) string {
	if item.SubscriptionID != "" {
		return subscriptionSelectionPrefixSubscriptionID + item.SubscriptionID
	}
	if item.PlanID != "" {
		return subscriptionSelectionPrefixPlanID + item.PlanID
	}
	return fmt.Sprintf("%s%d", subscriptionSelectionPrefixIndex, index)
}

func findSubscriptionItemIndexByToken(items []SubscriptionItem, token string) (int, bool) {
	if strings.HasPrefix(token, subscriptionSelectionPrefixSubscriptionID) {
		target := strings.TrimPrefix(token, subscriptionSelectionPrefixSubscriptionID)
		for i := range items {
			if items[i].SubscriptionID == target {
				return i, true
			}
		}
		return -1, false
	}
	if strings.HasPrefix(token, subscriptionSelectionPrefixPlanID) {
		target := strings.TrimPrefix(token, subscriptionSelectionPrefixPlanID)
		for i := range items {
			if items[i].PlanID == target {
				return i, true
			}
		}
		return -1, false
	}
	if strings.HasPrefix(token, subscriptionSelectionPrefixIndex) {
		idxStr := strings.TrimPrefix(token, subscriptionSelectionPrefixIndex)
		idx := common.String2Int(idxStr)
		if idx < 0 || idx >= len(items) {
			return -1, false
		}
		return idx, true
	}
	return -1, false
}

func parseSubscriptionData(raw string) ([]SubscriptionItem, error) {
	if raw == "" || raw == "null" {
		return []SubscriptionItem{}, nil
	}
	var items []SubscriptionItem
	if err := common.Unmarshal([]byte(raw), &items); err != nil {
		return nil, err
	}
	if items == nil {
		items = []SubscriptionItem{}
	}
	return items, nil
}

func marshalSubscriptionData(items []SubscriptionItem) (string, error) {
	data, err := common.Marshal(items)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func nowInSameUnit(referenceTimestamp int64, nowSec int64) int64 {
	// Heuristic: if the stored timestamp looks like milliseconds, compare using now in ms.
	// - seconds since epoch in 2025 is ~1.7e9
	// - milliseconds since epoch in 2025 is ~1.7e12
	if referenceTimestamp > 1_000_000_000_000 {
		return nowSec * 1000
	}
	return nowSec
}

func durationContains(duration SubscriptionDuration, nowSec int64) bool {
	if duration.StartAt == 0 || duration.EndAt == 0 {
		return false
	}
	if duration.StartAt != 0 {
		now := nowInSameUnit(duration.StartAt, nowSec)
		if now < duration.StartAt {
			return false
		}
	}
	if duration.EndAt != 0 {
		now := nowInSameUnit(duration.EndAt, nowSec)
		if now > duration.EndAt {
			return false
		}
	}
	return true
}

func isUsableSubscriptionItem(item SubscriptionItem, nowSec int64) bool {
	if item.Status != "deployed" {
		return false
	}
	if item.Limit5H.Available <= 0 || item.Limit7D.Available <= 0 {
		return false
	}
	if !durationContains(item.Duration, nowSec) {
		return false
	}
	return true
}

func advanceResetAt(resetAt int64, nowSec int64, intervalSec int64) int64 {
	if intervalSec <= 0 {
		return resetAt
	}
	now := nowInSameUnit(resetAt, nowSec)
	interval := intervalSec
	if resetAt > 1_000_000_000_000 {
		interval = intervalSec * 1000
	}
	if resetAt <= 0 {
		return now + interval
	}
	if now <= resetAt {
		return resetAt
	}
	periods := (now-resetAt)/interval + 1
	return resetAt + periods*interval
}

func resetAndPruneSubscriptionData(items []SubscriptionItem, nowSec int64) ([]SubscriptionItem, bool) {
	if len(items) == 0 {
		return items, false
	}
	changed := false
	pruned := make([]SubscriptionItem, 0, len(items))
	for _, item := range items {
		if item.Status != "deployed" {
			changed = true
			continue
		}

		now5h := nowInSameUnit(item.Limit5H.ResetAt, nowSec)
		if now5h > item.Limit5H.ResetAt {
			if item.Limit5H.Available != item.Limit5H.Total {
				item.Limit5H.Available = item.Limit5H.Total
				changed = true
			}
			newResetAt := advanceResetAt(item.Limit5H.ResetAt, nowSec, 5*60*60)
			if newResetAt != item.Limit5H.ResetAt {
				item.Limit5H.ResetAt = newResetAt
				changed = true
			}
		}
		now7d := nowInSameUnit(item.Limit7D.ResetAt, nowSec)
		if now7d > item.Limit7D.ResetAt {
			if item.Limit7D.Available != item.Limit7D.Total {
				item.Limit7D.Available = item.Limit7D.Total
				changed = true
			}
			newResetAt := advanceResetAt(item.Limit7D.ResetAt, nowSec, 7*24*60*60)
			if newResetAt != item.Limit7D.ResetAt {
				item.Limit7D.ResetAt = newResetAt
				changed = true
			}
		}
		pruned = append(pruned, item)
	}
	return pruned, changed || len(pruned) != len(items)
}

func consumeSubscriptionQuotaFromFirstUsableItem(items []SubscriptionItem, nowSec int64, amount int64) ([]SubscriptionItem, string, bool) {
	for i := range items {
		if !isUsableSubscriptionItem(items[i], nowSec) {
			continue
		}
		items[i].Limit5H.Available -= amount
		items[i].Limit7D.Available -= amount
		return items, encodeSubscriptionSelectionToken(items[i], i), true
	}
	return items, "", false
}

// PreConsumeUserSubscriptionQuota deducts quota from the first usable subscription item and
// returns a selection token for later adjustment/refund. It only updates users.subscription_data.
func PreConsumeUserSubscriptionQuota(userID int, nowSec int64, amount int64) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf("invalid userID: %d", userID)
	}
	var selectionToken string
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "subscription_data").
			Where("id = ?", userID).
			First(&user).Error
		if err != nil {
			return err
		}

		items, err := parseSubscriptionData(user.SubscriptionData)
		if err != nil {
			return err
		}
		updated, changed := resetAndPruneSubscriptionData(items, nowSec)
		items = updated

		items, token, ok := consumeSubscriptionQuotaFromFirstUsableItem(items, nowSec, amount)
		if !ok {
			return ErrSubscriptionQuotaExhausted
		}
		selectionToken = token

		raw, err := marshalSubscriptionData(items)
		if err != nil {
			return err
		}

		if err := tx.Model(&User{}).
			Where("id = ?", userID).
			Update("subscription_data", raw).Error; err != nil {
			return err
		}
		_ = changed
		return nil
	})
	return selectionToken, err
}

// AdjustUserSubscriptionQuotaBySelectionToken applies additional consumption/refund to the previously selected subscription.
// Positive delta means extra consume; negative delta means refund. It only updates users.subscription_data.
func AdjustUserSubscriptionQuotaBySelectionToken(userID int, selectionToken string, delta int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid userID: %d", userID)
	}
	if selectionToken == "" {
		return fmt.Errorf("empty selectionToken")
	}
	if delta == 0 {
		return nil
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		var user User
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "subscription_data").
			Where("id = ?", userID).
			First(&user).Error
		if err != nil {
			return err
		}

		items, err := parseSubscriptionData(user.SubscriptionData)
		if err != nil {
			return err
		}

		idx, ok := findSubscriptionItemIndexByToken(items, selectionToken)
		if !ok {
			return fmt.Errorf("subscription item not found for token: %s", selectionToken)
		}

		items[idx].Limit5H.Available -= delta
		items[idx].Limit7D.Available -= delta

		raw, err := marshalSubscriptionData(items)
		if err != nil {
			return err
		}

		return tx.Model(&User{}).
			Where("id = ?", userID).
			Update("subscription_data", raw).Error
	})
}

// ResetSubscriptionQuotaForAllUsers iterates subscription users and:
// - prunes non-deployed subscription items
// - resets 5h/7d available quota to total when now > reset_at
// It only updates users.subscription_data.
func ResetSubscriptionQuotaForAllUsers(nowSec int64) error {
	var users []User
	if err := DB.
		Model(&User{}).
		Select("id", "subscription_data").
		Where(commonGroupCol+" = ?", "subscription").
		Find(&users).Error; err != nil {
		return err
	}

	for _, user := range users {
		items, err := parseSubscriptionData(user.SubscriptionData)
		if err != nil {
			common.SysLog(fmt.Sprintf("failed to parse subscription_data for user %d: %v", user.Id, err))
			continue
		}
		updated, changed := resetAndPruneSubscriptionData(items, nowSec)
		if !changed {
			continue
		}
		raw, err := marshalSubscriptionData(updated)
		if err != nil {
			common.SysLog(fmt.Sprintf("failed to marshal subscription_data for user %d: %v", user.Id, err))
			continue
		}
		err = DB.Model(&User{}).
			Where("id = ?", user.Id).
			Update("subscription_data", raw).Error
		if err != nil {
			common.SysLog(fmt.Sprintf("failed to update subscription_data for user %d: %v", user.Id, err))
		}
	}
	return nil
}
