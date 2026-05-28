package model

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openQuotaPerUnitMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&Option{},
		&User{},
		&Token{},
		&Redemption{},
		&Channel{},
		&Midjourney{},
		&Task{},
		&QuotaData{},
		&UserUsageWindow{},
		&UserUsageReservation{},
		&AffRebateLog{},
		&TopUp{},
		&Log{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

func mustOptionValue(t *testing.T, db *gorm.DB, key string) string {
	t.Helper()
	var option Option
	if err := db.Where("key = ?", key).First(&option).Error; err != nil {
		t.Fatalf("get option %s: %v", key, err)
	}
	return option.Value
}

func TestMigrateLegacyDefaultQuotaPerUnitDataScalesRawQuotaFields(t *testing.T) {
	db := openQuotaPerUnitMigrationTestDB(t)

	options := []Option{
		{Key: "QuotaPerUnit", Value: strconv.FormatFloat(common.LegacyDefaultQuotaPerUnit, 'f', -1, 64)},
		{Key: "QuotaForNewUser", Value: "500000"},
		{Key: "QuotaForInviter", Value: "250000"},
		{Key: "QuotaForInvitee", Value: "125000"},
		{Key: "QuotaRemindThreshold", Value: "500000"},
		{Key: "PreConsumedQuota", Value: "1000"},
		{Key: "CreemProducts", Value: `[{"productId":"p1","quota":500000,"name":"Basic"}]`},
		{Key: "UserGroupUsageLimits", Value: `{"default":{"hourly":1}}`},
	}
	if err := db.Create(&options).Error; err != nil {
		t.Fatalf("seed options: %v", err)
	}

	subscriptionData := `{"items":[{"plan_name":"pro","5h_limit":{"total":500000,"available":250000,"reset_at":1},"7d_limit":{"total":1000000,"available":500000,"reset_at":2},"duration":{"start_at":1,"end_at":2},"status":"active"}],"last_reset_at":3}`
	user := User{
		Username:         "u",
		Password:         "password",
		Quota:            500000,
		UsedQuota:        250000,
		AffQuota:         125000,
		AffHistoryQuota:  62500,
		Setting:          `{"quota_warning_threshold":500000,"record_ip_log":true}`,
		SubscriptionData: subscriptionData,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&Token{UserId: user.Id, Key: "k", RemainQuota: 500000, UsedQuota: 250000}).Error; err != nil {
		t.Fatalf("seed token: %v", err)
	}
	if err := db.Create(&Redemption{Quota: 500000}).Error; err != nil {
		t.Fatalf("seed redemption: %v", err)
	}
	if err := db.Create(&Channel{UsedQuota: 500000}).Error; err != nil {
		t.Fatalf("seed channel: %v", err)
	}
	if err := db.Create(&Midjourney{Quota: 500000}).Error; err != nil {
		t.Fatalf("seed midjourney: %v", err)
	}
	if err := db.Create(&Task{Quota: 500000}).Error; err != nil {
		t.Fatalf("seed task: %v", err)
	}
	if err := db.Create(&QuotaData{Quota: 500000}).Error; err != nil {
		t.Fatalf("seed quota data: %v", err)
	}
	if err := db.Create(&UserUsageWindow{BudgetUsed: 500000, BudgetReserved: 250000}).Error; err != nil {
		t.Fatalf("seed usage window: %v", err)
	}
	if err := db.Create(&UserUsageReservation{ReservationID: "r", ReservedBudget: 500000}).Error; err != nil {
		t.Fatalf("seed usage reservation: %v", err)
	}
	if err := db.Create(&AffRebateLog{RewardQuota: 500000}).Error; err != nil {
		t.Fatalf("seed rebate log: %v", err)
	}
	if err := db.Create(&[]TopUp{
		{Amount: 500000, PaymentMethod: "creem", TradeNo: "creem"},
		{Amount: 500000, PaymentMethod: "stripe", TradeNo: "stripe"},
	}).Error; err != nil {
		t.Fatalf("seed topups: %v", err)
	}

	if err := migrateLegacyDefaultQuotaPerUnitData(db); err != nil {
		t.Fatalf("migrate quota data: %v", err)
	}

	var migratedUser User
	if err := db.First(&migratedUser, user.Id).Error; err != nil {
		t.Fatalf("get migrated user: %v", err)
	}
	if migratedUser.Quota != 1000000 || migratedUser.UsedQuota != 500000 || migratedUser.AffQuota != 250000 || migratedUser.AffHistoryQuota != 125000 {
		t.Fatalf("unexpected migrated user quotas: %#v", migratedUser)
	}
	var userSetting map[string]any
	if err := json.Unmarshal([]byte(migratedUser.Setting), &userSetting); err != nil {
		t.Fatalf("unmarshal migrated user setting: %v", err)
	}
	if userSetting["quota_warning_threshold"].(float64) != 1000000 {
		t.Fatalf("expected user warning threshold to double, got %v", userSetting["quota_warning_threshold"])
	}
	var migratedSubscription SubscriptionData
	if err := json.Unmarshal([]byte(migratedUser.SubscriptionData), &migratedSubscription); err != nil {
		t.Fatalf("unmarshal migrated subscription data: %v", err)
	}
	if migratedSubscription.Items[0].Limit5H.Total != 1000000 || migratedSubscription.Items[0].Limit7D.Available != 1000000 {
		t.Fatalf("unexpected migrated subscription data: %#v", migratedSubscription.Items[0])
	}

	var token Token
	if err := db.First(&token).Error; err != nil {
		t.Fatalf("get migrated token: %v", err)
	}
	if token.RemainQuota != 1000000 || token.UsedQuota != 500000 {
		t.Fatalf("unexpected migrated token: %#v", token)
	}

	var redemption Redemption
	_ = db.First(&redemption).Error
	if redemption.Quota != 1000000 {
		t.Fatalf("expected redemption quota to double, got %d", redemption.Quota)
	}
	var channel Channel
	_ = db.First(&channel).Error
	if channel.UsedQuota != 1000000 {
		t.Fatalf("expected channel used quota to double, got %d", channel.UsedQuota)
	}
	var usageWindow UserUsageWindow
	_ = db.First(&usageWindow).Error
	if usageWindow.BudgetUsed != 1000000 || usageWindow.BudgetReserved != 500000 {
		t.Fatalf("unexpected migrated usage window: %#v", usageWindow)
	}
	var reservation UserUsageReservation
	_ = db.First(&reservation).Error
	if reservation.ReservedBudget != 1000000 {
		t.Fatalf("expected reserved budget to double, got %d", reservation.ReservedBudget)
	}
	var rebate AffRebateLog
	_ = db.First(&rebate).Error
	if rebate.RewardQuota != 1000000 {
		t.Fatalf("expected rebate reward to double, got %d", rebate.RewardQuota)
	}

	var topups []TopUp
	if err := db.Order("trade_no asc").Find(&topups).Error; err != nil {
		t.Fatalf("get migrated topups: %v", err)
	}
	if topups[0].PaymentMethod != "creem" || topups[0].Amount != 1000000 {
		t.Fatalf("expected creem topup amount to double, got %#v", topups[0])
	}
	if topups[1].PaymentMethod != "stripe" || topups[1].Amount != 500000 {
		t.Fatalf("expected stripe amount to remain a display amount, got %#v", topups[1])
	}

	if got := mustOptionValue(t, db, "QuotaPerUnit"); got != strconv.FormatFloat(common.DefaultQuotaPerUnit, 'f', -1, 64) {
		t.Fatalf("expected quotaPerUnit option to migrate, got %s", got)
	}
	if got := mustOptionValue(t, db, "QuotaForNewUser"); got != "1000000" {
		t.Fatalf("expected QuotaForNewUser to double, got %s", got)
	}
	if got := mustOptionValue(t, db, "UserGroupUsageLimits"); got != `{"default":{"hourly":1}}` {
		t.Fatalf("UserGroupUsageLimits stores display values and should not be scaled, got %s", got)
	}
	var products []map[string]any
	if err := json.Unmarshal([]byte(mustOptionValue(t, db, "CreemProducts")), &products); err != nil {
		t.Fatalf("unmarshal migrated creem products: %v", err)
	}
	if products[0]["quota"].(float64) != 1000000 {
		t.Fatalf("expected Creem product quota to double, got %v", products[0]["quota"])
	}

	if err := migrateLegacyDefaultQuotaPerUnitData(db); err != nil {
		t.Fatalf("rerun quota data migration: %v", err)
	}
	var rerunUser User
	_ = db.First(&rerunUser, user.Id).Error
	if rerunUser.Quota != 1000000 {
		t.Fatalf("migration must be idempotent, got user quota %d", rerunUser.Quota)
	}
}

func TestMigrateLegacyDefaultQuotaPerUnitLogDataScalesLogsOnce(t *testing.T) {
	db := openQuotaPerUnitMigrationTestDB(t)
	originalDB := DB
	t.Cleanup(func() {
		DB = originalDB
	})
	DB = db

	if err := db.Create(&[]Option{
		{Key: quotaPerUnitMigrationSourceKey, Value: strconv.FormatFloat(common.LegacyDefaultQuotaPerUnit, 'f', -1, 64)},
	}).Error; err != nil {
		t.Fatalf("seed options: %v", err)
	}
	if err := db.Create(&Log{Quota: 500000}).Error; err != nil {
		t.Fatalf("seed log: %v", err)
	}

	if err := migrateLegacyDefaultQuotaPerUnitLogData(db); err != nil {
		t.Fatalf("migrate log data: %v", err)
	}
	var log Log
	if err := db.First(&log).Error; err != nil {
		t.Fatalf("get migrated log: %v", err)
	}
	if log.Quota != 1000000 {
		t.Fatalf("expected log quota to double, got %d", log.Quota)
	}

	if err := migrateLegacyDefaultQuotaPerUnitLogData(db); err != nil {
		t.Fatalf("rerun log migration: %v", err)
	}
	_ = db.First(&log).Error
	if log.Quota != 1000000 {
		t.Fatalf("log migration must be idempotent, got %d", log.Quota)
	}
}
