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
	hourStart := currentBounds[usageWindowHour].Start.Add(-2 * time.Hour)
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
			UserID:         userID,
			WindowKind:     usageWindowHour,
			WindowStart:    hourStart.Unix(),
			WindowEnd:      hourStart.Add(time.Hour).Unix(),
			BudgetReserved: initialBudget,
		},
		{
			UserID:          userID,
			WindowKind:      usageWindowDay,
			WindowStart:     dayStart.Unix(),
			WindowEnd:       dayStart.Add(24 * time.Hour).Unix(),
			RequestReserved: 1,
			TokenReserved:   initialTokens,
			BudgetReserved:  initialBudget,
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
		HourWindowStart:   hourStart.Unix(),
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
	if dayWindow.BudgetReserved != estimatedBudget {
		t.Fatalf("expected stored day window budget_reserved=%d, got %d", estimatedBudget, dayWindow.BudgetReserved)
	}

	var hourWindow model.UserUsageWindow
	if err := db.Where("user_id = ? AND window_kind = ? AND window_start = ?", userID, usageWindowHour, hourStart.Unix()).
		First(&hourWindow).Error; err != nil {
		t.Fatalf("load hour window: %v", err)
	}
	if hourWindow.BudgetReserved != estimatedBudget {
		t.Fatalf("expected stored hour window budget_reserved=%d, got %d", estimatedBudget, hourWindow.BudgetReserved)
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

func TestGetUserUsageLimitSnapshotIncludesHourlyDailyBudgetLimits(t *testing.T) {
	originalDB := model.DB
	t.Cleanup(func() {
		model.DB = originalDB
	})

	db, err := gorm.Open(sqlite.Open("file:usage-limit-snapshot-test?mode=memory&cache=shared"), &gorm.Config{})
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
	if err := setting.UpdateUserGroupUsageLimitsByJSONString(`{"test-group":{"hourly":100,"daily":200,"monthly":500}}`); err != nil {
		t.Fatalf("set usage limit policies: %v", err)
	}
	policy, found := setting.GetUserGroupUsageLimit("test-group")
	if !found {
		t.Fatalf("expected test-group policy to exist")
	}

	now := time.Now()
	currentBounds := getUsageWindowBounds(now)

	const userID = 2

	seedWindows := []*model.UserUsageWindow{
		{
			UserID:         userID,
			WindowKind:     usageWindowHour,
			WindowStart:    currentBounds[usageWindowHour].Start.Unix(),
			WindowEnd:      currentBounds[usageWindowHour].End.Unix(),
			BudgetUsed:     25,
			BudgetReserved: 10,
		},
		{
			UserID:         userID,
			WindowKind:     usageWindowDay,
			WindowStart:    currentBounds[usageWindowDay].Start.Unix(),
			WindowEnd:      currentBounds[usageWindowDay].End.Unix(),
			BudgetUsed:     60,
			BudgetReserved: 20,
		},
		{
			UserID:         userID,
			WindowKind:     usageWindowMonth,
			WindowStart:    currentBounds[usageWindowMonth].Start.Unix(),
			WindowEnd:      currentBounds[usageWindowMonth].End.Unix(),
			BudgetUsed:     120,
			BudgetReserved: 30,
		},
	}
	for _, window := range seedWindows {
		if err := db.Create(window).Error; err != nil {
			t.Fatalf("seed usage window %s: %v", window.WindowKind, err)
		}
	}

	snapshot, err := GetUserUsageLimitSnapshot(userID, "test-group")
	if err != nil {
		t.Fatalf("GetUserUsageLimitSnapshot returned error: %v", err)
	}

	if snapshot.NoLimitsConfigured {
		t.Fatalf("expected limits to be configured")
	}
	if snapshot.Metrics.Hourly.Limit == nil || policy.Hourly == nil || *snapshot.Metrics.Hourly.Limit != *policy.Hourly {
		t.Fatalf("expected hourly limit=%#v, got %#v", policy.Hourly, snapshot.Metrics.Hourly.Limit)
	}
	if snapshot.Metrics.Hourly.Used != 25 || snapshot.Metrics.Hourly.Pending != 10 {
		t.Fatalf("unexpected hourly metrics: used=%d pending=%d", snapshot.Metrics.Hourly.Used, snapshot.Metrics.Hourly.Pending)
	}
	if snapshot.Metrics.Daily.Limit == nil || policy.Daily == nil || *snapshot.Metrics.Daily.Limit != *policy.Daily {
		t.Fatalf("expected daily limit=%#v, got %#v", policy.Daily, snapshot.Metrics.Daily.Limit)
	}
	if snapshot.Metrics.Daily.Used != 60 || snapshot.Metrics.Daily.Pending != 20 {
		t.Fatalf("unexpected daily metrics: used=%d pending=%d", snapshot.Metrics.Daily.Used, snapshot.Metrics.Daily.Pending)
	}
	if snapshot.Metrics.Monthly.Limit == nil || policy.Monthly == nil || *snapshot.Metrics.Monthly.Limit != *policy.Monthly {
		t.Fatalf("expected monthly limit=%#v, got %#v", policy.Monthly, snapshot.Metrics.Monthly.Limit)
	}
	if snapshot.Metrics.Monthly.Used != 120 || snapshot.Metrics.Monthly.Pending != 30 {
		t.Fatalf("unexpected monthly metrics: used=%d pending=%d", snapshot.Metrics.Monthly.Used, snapshot.Metrics.Monthly.Pending)
	}
}
