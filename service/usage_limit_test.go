package service

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/types"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestReserveUsageEstimateUsesReservationWindowsForAccounting(t *testing.T) {
	originalDB := model.DB
	t.Cleanup(func() {
		model.DB = originalDB
	})

	db, err := gorm.Open(sqlite.Open("file:usage-limit-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.UserUsageWindow{}, &model.UserUsageReservation{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	model.DB = db

	originalPolicies := setting.UserGroupUsageLimits2JSONString()
	t.Cleanup(func() {
		if err := setting.UpdateUserGroupUsageLimitsByJSONString(originalPolicies); err != nil {
			t.Errorf("restore usage limit policies: %v", err)
		}
	})
	if err := setting.UpdateUserGroupUsageLimitsByJSONString(`{"test-group":{}}`); err != nil {
		t.Fatalf("set usage limit policies: %v", err)
	}

	now := time.Now()
	currentBounds := getUsageWindowBounds(now)
	minuteStart := currentBounds[usageWindowMinute].Start.Add(-2 * time.Minute)
	dayStart := currentBounds[usageWindowDay].Start.Add(-24 * time.Hour)
	monthStart := currentBounds[usageWindowMonth].Start.AddDate(0, -1, 0)

	const (
		userID          = 1
		reservationID   = "reservation-rollover"
		initialTokens   = int64(100)
		initialBudget   = int64(50)
		estimatedBudget = 200
	)

	seedWindows := []*model.UserUsageWindow{
		{
			UserID:          userID,
			WindowKind:      usageWindowMinute,
			WindowStart:     minuteStart.Unix(),
			WindowEnd:       minuteStart.Add(time.Minute).Unix(),
			RequestReserved: 1,
			TokenReserved:   initialTokens,
		},
		{
			UserID:          userID,
			WindowKind:      usageWindowDay,
			WindowStart:     dayStart.Unix(),
			WindowEnd:       dayStart.Add(24 * time.Hour).Unix(),
			RequestReserved: 1,
			TokenReserved:   initialTokens,
		},
		{
			UserID:         userID,
			WindowKind:     usageWindowMonth,
			WindowStart:    monthStart.Unix(),
			WindowEnd:      monthStart.AddDate(0, 1, 0).Unix(),
			BudgetReserved: initialBudget,
		},
	}
	for _, window := range seedWindows {
		if err := db.Create(window).Error; err != nil {
			t.Fatalf("seed usage window %s: %v", window.WindowKind, err)
		}
	}

	reservation := &model.UserUsageReservation{
		ReservationID:     reservationID,
		UserID:            userID,
		GroupName:         "test-group",
		MinuteWindowStart: minuteStart.Unix(),
		DayWindowStart:    dayStart.Unix(),
		MonthWindowStart:  monthStart.Unix(),
		ReservedRequests:  1,
		ReservedTokens:    initialTokens,
		ReservedBudget:    initialBudget,
		Status:            usageReservationReserved,
		ExpiresAt:         now.Add(time.Hour).Unix(),
	}
	if err := db.Create(reservation).Error; err != nil {
		t.Fatalf("seed reservation: %v", err)
	}

	meta := &types.TokenCountMeta{MaxTokens: 700}
	relayInfo := &relaycommon.RelayInfo{
		UserId:             userID,
		UserGroup:          "test-group",
		PromptTokens:       10,
		UsageReservationID: reservationID,
	}

	if apiErr := ReserveUsageEstimate(nil, relayInfo, meta, estimatedBudget); apiErr != nil {
		t.Fatalf("ReserveUsageEstimate returned error: %v", apiErr)
	}

	expectedTokens := estimateUsageTokens(relayInfo.PromptTokens, meta.MaxTokens)

	var minuteWindow model.UserUsageWindow
	if err := db.Where("user_id = ? AND window_kind = ? AND window_start = ?", userID, usageWindowMinute, minuteStart.Unix()).
		First(&minuteWindow).Error; err != nil {
		t.Fatalf("load minute window: %v", err)
	}
	if minuteWindow.TokenReserved != expectedTokens {
		t.Fatalf("expected stored minute window token_reserved=%d, got %d", expectedTokens, minuteWindow.TokenReserved)
	}

	var dayWindow model.UserUsageWindow
	if err := db.Where("user_id = ? AND window_kind = ? AND window_start = ?", userID, usageWindowDay, dayStart.Unix()).
		First(&dayWindow).Error; err != nil {
		t.Fatalf("load day window: %v", err)
	}
	if dayWindow.TokenReserved != expectedTokens {
		t.Fatalf("expected stored day window token_reserved=%d, got %d", expectedTokens, dayWindow.TokenReserved)
	}

	var monthWindow model.UserUsageWindow
	if err := db.Where("user_id = ? AND window_kind = ? AND window_start = ?", userID, usageWindowMonth, monthStart.Unix()).
		First(&monthWindow).Error; err != nil {
		t.Fatalf("load month window: %v", err)
	}
	if monthWindow.BudgetReserved != estimatedBudget {
		t.Fatalf("expected stored month window budget_reserved=%d, got %d", estimatedBudget, monthWindow.BudgetReserved)
	}

	var updatedReservation model.UserUsageReservation
	if err := db.Where("reservation_id = ?", reservationID).First(&updatedReservation).Error; err != nil {
		t.Fatalf("load updated reservation: %v", err)
	}
	if updatedReservation.ReservedTokens != expectedTokens {
		t.Fatalf("expected reservation reserved_tokens=%d, got %d", expectedTokens, updatedReservation.ReservedTokens)
	}
	if updatedReservation.ReservedBudget != estimatedBudget {
		t.Fatalf("expected reservation reserved_budget=%d, got %d", estimatedBudget, updatedReservation.ReservedBudget)
	}
}
