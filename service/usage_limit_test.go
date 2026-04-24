package service

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
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
	weekStart := currentBounds[usageWindowWeek].Start.AddDate(0, 0, -7)
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
			WindowKind:     usageWindowWeek,
			WindowStart:    weekStart.Unix(),
			WindowEnd:      weekStart.AddDate(0, 0, 7).Unix(),
			BudgetReserved: initialBudget,
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
		WeekWindowStart:   weekStart.Unix(),
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

	var weekWindow model.UserUsageWindow
	if err := db.Where("user_id = ? AND window_kind = ? AND window_start = ?", userID, usageWindowWeek, weekStart.Unix()).
		First(&weekWindow).Error; err != nil {
		t.Fatalf("load week window: %v", err)
	}
	if weekWindow.BudgetReserved != estimatedBudget {
		t.Fatalf("expected stored week window budget_reserved=%d, got %d", estimatedBudget, weekWindow.BudgetReserved)
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

func TestGetUserUsageLimitSnapshotIncludesWeeklyBudgetLimits(t *testing.T) {
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
	if err := setting.UpdateUserGroupUsageLimitsByJSONString(`{"test-group":{"hourly":100,"daily":200,"weekly":350,"monthly":500}}`); err != nil {
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
			WindowKind:     usageWindowWeek,
			WindowStart:    currentBounds[usageWindowWeek].Start.Unix(),
			WindowEnd:      currentBounds[usageWindowWeek].End.Unix(),
			BudgetUsed:     90,
			BudgetReserved: 15,
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
	if snapshot.Metrics.Hourly.Used == nil || snapshot.Metrics.Hourly.Pending == nil || *snapshot.Metrics.Hourly.Used != 25 || *snapshot.Metrics.Hourly.Pending != 10 {
		t.Fatalf("unexpected hourly metrics: used=%#v pending=%#v", snapshot.Metrics.Hourly.Used, snapshot.Metrics.Hourly.Pending)
	}
	if snapshot.Metrics.Daily.Limit == nil || policy.Daily == nil || *snapshot.Metrics.Daily.Limit != *policy.Daily {
		t.Fatalf("expected daily limit=%#v, got %#v", policy.Daily, snapshot.Metrics.Daily.Limit)
	}
	if snapshot.Metrics.Daily.Used == nil || snapshot.Metrics.Daily.Pending == nil || *snapshot.Metrics.Daily.Used != 60 || *snapshot.Metrics.Daily.Pending != 20 {
		t.Fatalf("unexpected daily metrics: used=%#v pending=%#v", snapshot.Metrics.Daily.Used, snapshot.Metrics.Daily.Pending)
	}
	if snapshot.Metrics.Weekly.Limit == nil || policy.Weekly == nil || *snapshot.Metrics.Weekly.Limit != *policy.Weekly {
		t.Fatalf("expected weekly limit=%#v, got %#v", policy.Weekly, snapshot.Metrics.Weekly.Limit)
	}
	if snapshot.Metrics.Weekly.Used == nil || snapshot.Metrics.Weekly.Pending == nil || *snapshot.Metrics.Weekly.Used != 90 || *snapshot.Metrics.Weekly.Pending != 15 {
		t.Fatalf("unexpected weekly metrics: used=%#v pending=%#v", snapshot.Metrics.Weekly.Used, snapshot.Metrics.Weekly.Pending)
	}
	if snapshot.Metrics.Monthly.Limit == nil || policy.Monthly == nil || *snapshot.Metrics.Monthly.Limit != *policy.Monthly {
		t.Fatalf("expected monthly limit=%#v, got %#v", policy.Monthly, snapshot.Metrics.Monthly.Limit)
	}
	if snapshot.Metrics.Monthly.Used == nil || snapshot.Metrics.Monthly.Pending == nil || *snapshot.Metrics.Monthly.Used != 120 || *snapshot.Metrics.Monthly.Pending != 30 {
		t.Fatalf("unexpected monthly metrics: used=%#v pending=%#v", snapshot.Metrics.Monthly.Used, snapshot.Metrics.Monthly.Pending)
	}
}

func TestGetUserUsageLimitSnapshotHidesMetricDetailsWhenConfigured(t *testing.T) {
	originalDB := model.DB
	t.Cleanup(func() {
		model.DB = originalDB
	})

	db, err := gorm.Open(sqlite.Open("file:usage-limit-hidden-test?mode=memory&cache=shared"), &gorm.Config{})
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
	if err := setting.UpdateUserGroupUsageLimitsByJSONString(`{"test-group":{"rpm":10,"rpm_hide_details":true,"weekly":200,"weekly_hide_details":true}}`); err != nil {
		t.Fatalf("set usage limit policies: %v", err)
	}
	policy, found := setting.GetUserGroupUsageLimit("test-group")
	if !found {
		t.Fatalf("expected test-group policy to exist")
	}

	now := time.Now()
	currentBounds := getUsageWindowBounds(now)

	const userID = 3

	seedWindows := []*model.UserUsageWindow{
		{
			UserID:          userID,
			WindowKind:      usageWindowMinute,
			WindowStart:     currentBounds[usageWindowMinute].Start.Unix(),
			WindowEnd:       currentBounds[usageWindowMinute].End.Unix(),
			RequestUsed:     3,
			RequestReserved: 2,
		},
		{
			UserID:         userID,
			WindowKind:     usageWindowWeek,
			WindowStart:    currentBounds[usageWindowWeek].Start.Unix(),
			WindowEnd:      currentBounds[usageWindowWeek].End.Unix(),
			BudgetUsed:     50,
			BudgetReserved: 10,
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

	if !snapshot.Metrics.RPM.HideDetails {
		t.Fatalf("expected rpm hide_details to be enabled")
	}
	if snapshot.Metrics.RPM.Limit != nil || snapshot.Metrics.RPM.Used != nil || snapshot.Metrics.RPM.Pending != nil || snapshot.Metrics.RPM.Remaining != nil {
		t.Fatalf("expected rpm numeric details to be hidden, got %#v", snapshot.Metrics.RPM)
	}
	expectedRPMPercent := computeConsumptionPercent(policy.RPM, 3, 2)
	if snapshot.Metrics.RPM.ConsumptionPercent == nil || expectedRPMPercent == nil || *snapshot.Metrics.RPM.ConsumptionPercent != *expectedRPMPercent {
		t.Fatalf("expected rpm consumption percent=%#v, got %#v", expectedRPMPercent, snapshot.Metrics.RPM.ConsumptionPercent)
	}

	if !snapshot.Metrics.Weekly.HideDetails {
		t.Fatalf("expected weekly hide_details to be enabled")
	}
	if snapshot.Metrics.Weekly.Limit != nil || snapshot.Metrics.Weekly.Used != nil || snapshot.Metrics.Weekly.Pending != nil || snapshot.Metrics.Weekly.Remaining != nil {
		t.Fatalf("expected weekly numeric details to be hidden, got %#v", snapshot.Metrics.Weekly)
	}
	expectedWeeklyPercent := computeConsumptionPercent(policy.Weekly, 50, 10)
	if snapshot.Metrics.Weekly.ConsumptionPercent == nil || expectedWeeklyPercent == nil || *snapshot.Metrics.Weekly.ConsumptionPercent != *expectedWeeklyPercent {
		t.Fatalf("expected weekly consumption percent=%#v, got %#v", expectedWeeklyPercent, snapshot.Metrics.Weekly.ConsumptionPercent)
	}
}

func TestGetUserUsageLimitSnapshotAppliesMultiplierRules(t *testing.T) {
	originalDB := model.DB
	t.Cleanup(func() {
		model.DB = originalDB
	})

	db, err := gorm.Open(sqlite.Open("file:usage-limit-multiplier-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.UserUsageWindow{}, &model.UserUsageReservation{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	model.DB = db

	originalPolicies := setting.UserGroupUsageLimits2JSONString()
	originalMultiplierRules := setting.UserUsageLimitMultiplierRules2JSONString()
	t.Cleanup(func() {
		if err := setting.UpdateUserGroupUsageLimitsByJSONString(originalPolicies); err != nil {
			t.Errorf("restore usage limit policies: %v", err)
		}
		if err := setting.UpdateUserUsageLimitMultiplierRulesByJSONString(originalMultiplierRules); err != nil {
			t.Errorf("restore usage limit multiplier rules: %v", err)
		}
	})

	if err := setting.UpdateUserGroupUsageLimitsByJSONString(`{"vip":{"daily":100,"weekly":200}}`); err != nil {
		t.Fatalf("set usage limit policies: %v", err)
	}
	if err := setting.UpdateUserUsageLimitMultiplierRulesByJSONString(`[
		{"scope":"all","metrics":["daily"],"multiplier":1.05},
		{"scope":"groups","group_names":["vip"],"metrics":["daily","weekly"],"multiplier":1.1},
		{"scope":"users","user_ids":[9],"metrics":["weekly"],"multiplier":1.25}
	]`); err != nil {
		t.Fatalf("set usage limit multiplier rules: %v", err)
	}

	snapshot, err := GetUserUsageLimitSnapshot(9, "vip")
	if err != nil {
		t.Fatalf("GetUserUsageLimitSnapshot returned error: %v", err)
	}

	policy, found := resolveUserUsageLimitPolicy(9, "vip")
	if !found {
		t.Fatalf("expected effective usage policy to exist")
	}
	if snapshot.Metrics.Daily.Limit == nil || policy.Daily == nil || *snapshot.Metrics.Daily.Limit != *policy.Daily {
		t.Fatalf("expected daily limit=%#v, got %#v", policy.Daily, snapshot.Metrics.Daily.Limit)
	}
	if snapshot.Metrics.Weekly.Limit == nil || policy.Weekly == nil || *snapshot.Metrics.Weekly.Limit != *policy.Weekly {
		t.Fatalf("expected weekly limit=%#v, got %#v", policy.Weekly, snapshot.Metrics.Weekly.Limit)
	}
}

func TestResetUserUsageLimitsByScope(t *testing.T) {
	originalDB := model.DB
	t.Cleanup(func() {
		model.DB = originalDB
	})

	db, err := gorm.Open(sqlite.Open("file:usage-limit-reset-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserUsageWindow{}, &model.UserUsageReservation{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	model.DB = db

	users := []*model.User{
		{Username: "reset_all_enabled", Password: "password", Group: "default", Status: common.UserStatusEnabled, AffCode: "reset_aff_1"},
		{Username: "reset_group_disabled", Password: "password", Group: "vip", Status: common.UserStatusDisabled, AffCode: "reset_aff_2"},
		{Username: "reset_group_enabled", Password: "password", Group: "vip", Status: common.UserStatusEnabled, AffCode: "reset_aff_3"},
	}
	for _, user := range users {
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("create user %s: %v", user.Username, err)
		}
	}

	seedUsage := func(userID int) {
		window := &model.UserUsageWindow{
			UserID:      userID,
			WindowKind:  usageWindowDay,
			WindowStart: 1,
			WindowEnd:   2,
			RequestUsed: 1,
		}
		if err := db.Create(window).Error; err != nil {
			t.Fatalf("seed usage window for user %d: %v", userID, err)
		}
		reservation := &model.UserUsageReservation{
			ReservationID:     time.Now().Add(time.Duration(userID) * time.Millisecond).Format("20060102150405.000000000"),
			UserID:            userID,
			GroupName:         "test",
			MinuteWindowStart: 1,
			HourWindowStart:   1,
			DayWindowStart:    1,
			WeekWindowStart:   1,
			MonthWindowStart:  1,
			ReservedRequests:  1,
			Status:            usageReservationReserved,
			ExpiresAt:         time.Now().Add(time.Hour).Unix(),
		}
		if err := db.Create(reservation).Error; err != nil {
			t.Fatalf("seed reservation for user %d: %v", userID, err)
		}
	}

	seedUsage(users[0].Id)
	seedUsage(users[1].Id)
	seedUsage(users[2].Id)

	allResult, err := ResetUserUsageLimits(setting.UsageLimitTargetAll, nil, nil)
	if err != nil {
		t.Fatalf("ResetUserUsageLimits all returned error: %v", err)
	}
	if allResult.TargetedUsers != 2 {
		t.Fatalf("expected 2 enabled users to be reset, got %d", allResult.TargetedUsers)
	}
	if allResult.SkippedBannedUsers != 1 {
		t.Fatalf("expected 1 banned user to be skipped, got %d", allResult.SkippedBannedUsers)
	}

	var remainingEnabledWindows int64
	if err := db.Model(&model.UserUsageWindow{}).Where("user_id IN ?", []int{users[0].Id, users[2].Id}).Count(&remainingEnabledWindows).Error; err != nil {
		t.Fatalf("count remaining enabled windows: %v", err)
	}
	if remainingEnabledWindows != 0 {
		t.Fatalf("expected enabled users windows to be cleared, got %d", remainingEnabledWindows)
	}

	var disabledWindows int64
	if err := db.Model(&model.UserUsageWindow{}).Where("user_id = ?", users[1].Id).Count(&disabledWindows).Error; err != nil {
		t.Fatalf("count disabled windows: %v", err)
	}
	if disabledWindows != 1 {
		t.Fatalf("expected disabled user window to remain after all reset, got %d", disabledWindows)
	}

	groupResult, err := ResetUserUsageLimits(setting.UsageLimitTargetGroups, []string{"vip"}, nil)
	if err != nil {
		t.Fatalf("ResetUserUsageLimits groups returned error: %v", err)
	}
	if groupResult.TargetedUsers != 2 {
		t.Fatalf("expected both vip users to be targeted, got %d", groupResult.TargetedUsers)
	}

	var remainingVIPReservations int64
	if err := db.Model(&model.UserUsageReservation{}).Where("user_id IN ?", []int{users[1].Id, users[2].Id}).Count(&remainingVIPReservations).Error; err != nil {
		t.Fatalf("count vip reservations: %v", err)
	}
	if remainingVIPReservations != 0 {
		t.Fatalf("expected vip reservations to be cleared, got %d", remainingVIPReservations)
	}
}
