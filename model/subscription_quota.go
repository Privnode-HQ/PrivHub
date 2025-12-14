package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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

func unmarshalFlexibleInt64(raw json.RawMessage) (int64, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return 0, nil
	}

	// Fast-path: proper JSON number.
	var asInt int64
	if err := json.Unmarshal(raw, &asInt); err == nil {
		return asInt, nil
	}

	// Accept string-encoded numbers.
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		asString = strings.TrimSpace(asString)
		if asString == "" {
			return 0, nil
		}
		if v, err := strconv.ParseInt(asString, 10, 64); err == nil {
			return v, nil
		}
		if v, err := strconv.ParseFloat(asString, 64); err == nil {
			return int64(v), nil
		}
		return 0, fmt.Errorf("invalid int64 string %q", asString)
	}

	// Fallback: decode with UseNumber to preserve integer values.
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var asNumber json.Number
	if err := decoder.Decode(&asNumber); err == nil {
		if v, err := asNumber.Int64(); err == nil {
			return v, nil
		}
		if v, err := asNumber.Float64(); err == nil {
			return int64(v), nil
		}
	}

	return 0, fmt.Errorf("invalid int64 json value: %s", string(raw))
}

func (l *SubscriptionLimit) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		return nil
	}

	type rawLimit struct {
		Total     json.RawMessage `json:"total"`
		Available json.RawMessage `json:"available"`
		ResetAt   json.RawMessage `json:"reset_at"`
	}
	var raw rawLimit
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	total, err := unmarshalFlexibleInt64(raw.Total)
	if err != nil {
		return fmt.Errorf("invalid total: %w", err)
	}
	available, err := unmarshalFlexibleInt64(raw.Available)
	if err != nil {
		return fmt.Errorf("invalid available: %w", err)
	}
	resetAt, err := unmarshalFlexibleInt64(raw.ResetAt)
	if err != nil {
		return fmt.Errorf("invalid reset_at: %w", err)
	}
	l.Total = total
	l.Available = available
	l.ResetAt = resetAt
	return nil
}

type SubscriptionDuration struct {
	StartAt          int64 `json:"start_at"`
	EndAt            int64 `json:"end_at"`
	AutoRenewEnabled bool  `json:"auto_renew_enabled"`
}

func (d *SubscriptionDuration) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		return nil
	}

	type rawDuration struct {
		StartAt          json.RawMessage `json:"start_at"`
		EndAt            json.RawMessage `json:"end_at"`
		AutoRenewEnabled bool            `json:"auto_renew_enabled"`
	}
	var raw rawDuration
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	startAt, err := unmarshalFlexibleInt64(raw.StartAt)
	if err != nil {
		return fmt.Errorf("invalid start_at: %w", err)
	}
	endAt, err := unmarshalFlexibleInt64(raw.EndAt)
	if err != nil {
		return fmt.Errorf("invalid end_at: %w", err)
	}
	d.StartAt = startAt
	d.EndAt = endAt
	d.AutoRenewEnabled = raw.AutoRenewEnabled
	return nil
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

// SubscriptionData is stored in users.subscription_data.
//
// Legacy format: a JSON array of SubscriptionItem.
// Current format: {"items": [...], "last_reset_at": <unix seconds>}.
type SubscriptionData struct {
	Items       []SubscriptionItem `json:"items"`
	LastResetAt int64              `json:"last_reset_at"`
}

func (d *SubscriptionData) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		d.Items = []SubscriptionItem{}
		d.LastResetAt = 0
		return nil
	}

	// Legacy array format.
	if len(b) > 0 && b[0] == '[' {
		var items []SubscriptionItem
		if err := common.Unmarshal(b, &items); err != nil {
			return err
		}
		if items == nil {
			items = []SubscriptionItem{}
		}
		d.Items = items
		d.LastResetAt = 0
		return nil
	}

	type rawData struct {
		Items       []SubscriptionItem `json:"items"`
		LastResetAt json.RawMessage    `json:"last_reset_at"`
	}
	var raw rawData
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	lastResetAt, err := unmarshalFlexibleInt64(raw.LastResetAt)
	if err != nil {
		return fmt.Errorf("invalid last_reset_at: %w", err)
	}
	if raw.Items == nil {
		raw.Items = []SubscriptionItem{}
	}
	d.Items = raw.Items
	d.LastResetAt = lastResetAt
	return nil
}

const subscriptionSelectionPrefixFingerprint = "fp:"

func subscriptionItemFingerprint(item SubscriptionItem) string {
	stable := fmt.Sprintf(
		"plan_name=%s|plan_id=%s|subscription_id=%s|owner=%d|start=%d|end=%d",
		item.PlanName,
		item.PlanID,
		item.SubscriptionID,
		item.Owner,
		item.Duration.StartAt,
		item.Duration.EndAt,
	)
	return common.Sha1([]byte(stable))
}

func encodeSubscriptionSelectionToken(item SubscriptionItem, index int) string {
	fp := subscriptionItemFingerprint(item)
	return fmt.Sprintf("%s%s|i:%d", subscriptionSelectionPrefixFingerprint, fp, index)
}

func parseSelectionToken(token string) (fingerprint string, hintIndex *int, ok bool) {
	if !strings.HasPrefix(token, subscriptionSelectionPrefixFingerprint) {
		return "", nil, false
	}
	payload := strings.TrimPrefix(token, subscriptionSelectionPrefixFingerprint)
	parts := strings.Split(payload, "|")
	if len(parts) == 0 || parts[0] == "" {
		return "", nil, false
	}
	fingerprint = parts[0]
	for _, p := range parts[1:] {
		if !strings.HasPrefix(p, "i:") {
			continue
		}
		v := strings.TrimPrefix(p, "i:")
		idx, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		hintIndex = &idx
		break
	}
	return fingerprint, hintIndex, true
}

func findSubscriptionItemIndexByToken(items []SubscriptionItem, token string) (int, bool) {
	fp, hintIdx, ok := parseSelectionToken(token)
	if !ok {
		return -1, false
	}
	if hintIdx != nil {
		idx := *hintIdx
		if idx >= 0 && idx < len(items) {
			if subscriptionItemFingerprint(items[idx]) == fp {
				return idx, true
			}
		}
	}
	for i := range items {
		if subscriptionItemFingerprint(items[i]) == fp {
			return i, true
		}
	}
	return -1, false
}

func parseSubscriptionData(raw string) (SubscriptionData, bool, error) {
	if raw == "" || raw == "null" {
		return SubscriptionData{Items: []SubscriptionItem{}}, false, nil
	}
	trimmed := strings.TrimSpace(raw)
	legacy := strings.HasPrefix(trimmed, "[")
	var data SubscriptionData
	if err := common.Unmarshal([]byte(raw), &data); err != nil {
		return SubscriptionData{}, legacy, err
	}
	if data.Items == nil {
		data.Items = []SubscriptionItem{}
	}
	return data, legacy, nil
}

func marshalSubscriptionData(data SubscriptionData) (string, error) {
	if data.Items == nil {
		data.Items = []SubscriptionItem{}
	}
	payload, err := common.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(payload), nil
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

func isUsableSubscriptionItem(item SubscriptionItem, nowSec int64, amount int64) bool {
	if item.Status != "deployed" {
		return false
	}
	if item.Limit5H.Available <= 0 || item.Limit7D.Available <= 0 {
		return false
	}
	if amount > 0 {
		if item.Limit5H.Available < amount || item.Limit7D.Available < amount {
			return false
		}
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

func resetAndPruneSubscriptionData(data SubscriptionData, nowSec int64) (SubscriptionData, bool) {
	if len(data.Items) == 0 {
		if data.Items == nil {
			data.Items = []SubscriptionItem{}
		}
		return data, false
	}
	originalLen := len(data.Items)
	changed := false
	resetOccurred := false
	pruned := make([]SubscriptionItem, 0, len(data.Items))
	for _, item := range data.Items {
		if item.Status != "deployed" {
			changed = true
			continue
		}

		now5h := nowInSameUnit(item.Limit5H.ResetAt, nowSec)
		if now5h > item.Limit5H.ResetAt {
			resetOccurred = true
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
			resetOccurred = true
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
	data.Items = pruned
	if resetOccurred {
		data.LastResetAt = nowSec
		changed = true
	}
	return data, changed || len(pruned) != originalLen
}

func consumeSubscriptionQuotaFromFirstUsableItem(items []SubscriptionItem, nowSec int64, amount int64) ([]SubscriptionItem, string, bool) {
	if amount < 0 {
		return items, "", false
	}
	for i := range items {
		if !isUsableSubscriptionItem(items[i], nowSec, amount) {
			continue
		}
		if amount > 0 {
			items[i].Limit5H.Available -= amount
			items[i].Limit7D.Available -= amount
		}
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
	if amount < 0 {
		return "", fmt.Errorf("invalid amount: %d", amount)
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

		data, legacy, err := parseSubscriptionData(user.SubscriptionData)
		if err != nil {
			return err
		}
		updated, changed := resetAndPruneSubscriptionData(data, nowSec)
		data = updated

		items, token, ok := consumeSubscriptionQuotaFromFirstUsableItem(data.Items, nowSec, amount)
		if !ok {
			return ErrSubscriptionQuotaExhausted
		}
		selectionToken = token
		data.Items = items

		raw, err := marshalSubscriptionData(data)
		if err != nil {
			return err
		}
		if legacy || changed || amount > 0 {
			if err := tx.Model(&User{}).
				Where("id = ?", userID).
				Update("subscription_data", raw).Error; err != nil {
				return err
			}
		}
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

		data, _, err := parseSubscriptionData(user.SubscriptionData)
		if err != nil {
			return err
		}
		items := data.Items

		idx, ok := findSubscriptionItemIndexByToken(items, selectionToken)
		if !ok {
			return fmt.Errorf("subscription item not found for token: %s", selectionToken)
		}
		item := items[idx]
		new5h := item.Limit5H.Available - delta
		new7d := item.Limit7D.Available - delta
		if delta > 0 {
			if new5h < 0 || new7d < 0 {
				return ErrSubscriptionQuotaExhausted
			}
		}
		if delta < 0 {
			if new5h > item.Limit5H.Total {
				new5h = item.Limit5H.Total
			}
			if new7d > item.Limit7D.Total {
				new7d = item.Limit7D.Total
			}
		}
		items[idx].Limit5H.Available = new5h
		items[idx].Limit7D.Available = new7d

		data.Items = items
		raw, err := marshalSubscriptionData(data)
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
	var userIDs []int
	if err := DB.Model(&User{}).
		Where(commonGroupCol+" = ?", "subscription").
		Pluck("id", &userIDs).Error; err != nil {
		return err
	}

	for _, userID := range userIDs {
		_ = DB.Transaction(func(tx *gorm.DB) error {
			var user User
			err := tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Select("id", "subscription_data").
				Where("id = ?", userID).
				First(&user).Error
			if err != nil {
				common.SysLog(fmt.Sprintf("failed to load subscription_data for user %d: %v", userID, err))
				return nil
			}
			data, legacy, err := parseSubscriptionData(user.SubscriptionData)
			if err != nil {
				common.SysLog(fmt.Sprintf("failed to parse subscription_data for user %d: %v", userID, err))
				return nil
			}
			updated, changed := resetAndPruneSubscriptionData(data, nowSec)
			if !legacy && !changed {
				return nil
			}
			raw, err := marshalSubscriptionData(updated)
			if err != nil {
				common.SysLog(fmt.Sprintf("failed to marshal subscription_data for user %d: %v", userID, err))
				return nil
			}
			if err := tx.Model(&User{}).
				Where("id = ?", userID).
				Update("subscription_data", raw).Error; err != nil {
				common.SysLog(fmt.Sprintf("failed to update subscription_data for user %d: %v", userID, err))
				return nil
			}
			return nil
		})
	}
	return nil
}
