package model

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupR2STestDB(t *testing.T) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	oldDB := DB
	oldLogDB := LOG_DB
	oldQuotaPerUnit := common.QuotaPerUnit
	DB = db
	LOG_DB = db
	common.QuotaPerUnit = common.DefaultQuotaPerUnit

	common.OptionMapRWMutex.Lock()
	oldOptionMap := common.OptionMap
	common.OptionMap = map[string]string{
		R2SReceiptRequiredOptionKey:     "false",
		R2SDefaultCurrencyCodeOptionKey: "USD",
		R2SBalanceReminderDaysOptionKey: "30",
	}
	common.OptionMapRWMutex.Unlock()

	if err := db.AutoMigrate(
		&Option{},
		&R2SSupplier{},
		&R2SChannelBinding{},
		&R2SPayment{},
		&R2SBalanceUpdate{},
		&R2SRecognitionRecord{},
		&Log{},
		&Channel{},
		&TopUpPromotionCampaign{},
		&TopUpPromotionRedemption{},
	); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		DB = oldDB
		LOG_DB = oldLogDB
		common.QuotaPerUnit = oldQuotaPerUnit
		common.OptionMapRWMutex.Lock()
		common.OptionMap = oldOptionMap
		common.OptionMapRWMutex.Unlock()
	})
}

func setR2STestOption(key string, value string) {
	common.OptionMapRWMutex.Lock()
	common.OptionMap[key] = value
	common.OptionMapRWMutex.Unlock()
}

func requireR2SFloat(t *testing.T, got float64, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.000001 {
		t.Fatalf("expected %.6f, got %.6f", want, got)
	}
}

func TestCreateR2SPaymentRequiresReceiptAndUpdatesBalance(t *testing.T) {
	setupR2STestDB(t)
	setR2STestOption(R2SReceiptRequiredOptionKey, strconv.FormatBool(true))

	supplier := &R2SSupplier{
		Name:                "Upstream A",
		DefaultCurrencyCode: "USDT",
		DefaultExchangeRate: 7,
		BalanceAmount:       10,
		BalanceCurrencyCode: "USDT",
		BalanceReminderDays: 15,
	}
	if err := supplier.Insert(); err != nil {
		t.Fatal(err)
	}

	_, _, err := CreateR2SPaymentWithBalance(&R2SPayment{
		SupplierId:   supplier.Id,
		PaymentType:  R2SPaymentTypePrepaid,
		Amount:       100,
		CurrencyCode: "usdt",
		ExchangeRate: 7,
	}, nil, 100)
	if err == nil {
		t.Fatal("expected missing receipt to fail when receipt setting is required")
	}

	payment, balanceUpdate, err := CreateR2SPaymentWithBalance(&R2SPayment{
		SupplierId:   supplier.Id,
		PaymentType:  R2SPaymentTypePrepaid,
		Amount:       100,
		CurrencyCode: "usdt",
		ExchangeRate: 7,
		ReceiptURL:   "https://example.com/receipt.png",
	}, nil, 100)
	if err != nil {
		t.Fatal(err)
	}
	if balanceUpdate == nil {
		t.Fatal("expected prepaid payment to create a balance update")
	}
	if !payment.ReceiptRequired {
		t.Fatal("expected payment to snapshot receipt requirement")
	}
	requireR2SFloat(t, payment.SystemAmount, 700)
	requireR2SFloat(t, payment.BalanceBefore, 10)
	requireR2SFloat(t, payment.BalanceAfter, 110)
	requireR2SFloat(t, balanceUpdate.DeltaAmount, 100)
	requireR2SFloat(t, balanceUpdate.SystemDeltaAmount, 700)

	var refreshed R2SSupplier
	if err := DB.First(&refreshed, supplier.Id).Error; err != nil {
		t.Fatal(err)
	}
	requireR2SFloat(t, refreshed.BalanceAmount, 110)
	requireR2SFloat(t, refreshed.SystemBalanceAmount, 770)
	if refreshed.NextBalanceReminderAt == 0 {
		t.Fatal("expected balance update to schedule next reminder")
	}
}

func TestR2SManualBalanceUpdatePreservesReminderWhenOmitted(t *testing.T) {
	setupR2STestDB(t)

	supplier := &R2SSupplier{
		Name:                "Upstream B",
		DefaultCurrencyCode: "USD",
		DefaultExchangeRate: 1,
		BalanceAmount:       20,
		BalanceCurrencyCode: "USD",
		BalanceReminderDays: 14,
	}
	if err := supplier.Insert(); err != nil {
		t.Fatal(err)
	}

	update, err := CreateR2SManualBalanceUpdate(
		supplier.Id,
		50,
		"USD",
		1,
		nil,
		"monthly statement",
		100,
	)
	if err != nil {
		t.Fatal(err)
	}
	if update.ReminderDaysSnapshot != 14 {
		t.Fatalf("expected reminder snapshot 14, got %d", update.ReminderDaysSnapshot)
	}

	var refreshed R2SSupplier
	if err := DB.First(&refreshed, supplier.Id).Error; err != nil {
		t.Fatal(err)
	}
	if refreshed.BalanceReminderDays != 14 {
		t.Fatalf("expected reminder days to remain 14, got %d", refreshed.BalanceReminderDays)
	}

	zeroDays := 0
	update, err = CreateR2SManualBalanceUpdate(
		supplier.Id,
		60,
		"USD",
		1,
		&zeroDays,
		"disable reminders",
		100,
	)
	if err != nil {
		t.Fatal(err)
	}
	if update.ReminderDaysSnapshot != 0 || update.NextReminderAt != 0 {
		t.Fatalf("expected reminders disabled, got snapshot=%d next=%d", update.ReminderDaysSnapshot, update.NextReminderAt)
	}
	if err := DB.First(&refreshed, supplier.Id).Error; err != nil {
		t.Fatal(err)
	}
	if refreshed.BalanceReminderDays != 0 {
		t.Fatalf("expected reminder days to be 0, got %d", refreshed.BalanceReminderDays)
	}
}

func TestR2SSupplierUpdateCanDisableBalanceReminder(t *testing.T) {
	setupR2STestDB(t)

	supplier := &R2SSupplier{
		Name:                "Upstream Reminder",
		DefaultCurrencyCode: "USD",
		DefaultExchangeRate: 1,
	}
	if err := supplier.Insert(); err != nil {
		t.Fatal(err)
	}
	if supplier.BalanceReminderDays != 30 {
		t.Fatalf("expected create to use default reminder, got %d", supplier.BalanceReminderDays)
	}
	if supplier.NextBalanceReminderAt == 0 {
		t.Fatal("expected default reminder to schedule next reminder")
	}

	supplier.BalanceReminderDays = 0
	if err := supplier.Update(); err != nil {
		t.Fatal(err)
	}

	var refreshed R2SSupplier
	if err := DB.First(&refreshed, supplier.Id).Error; err != nil {
		t.Fatal(err)
	}
	if refreshed.BalanceReminderDays != 0 {
		t.Fatalf("expected reminder days to be 0, got %d", refreshed.BalanceReminderDays)
	}
	if refreshed.NextBalanceReminderAt != 0 {
		t.Fatalf("expected next reminder to be cleared, got %d", refreshed.NextBalanceReminderAt)
	}
}

func TestR2SSummaryExcludesDisabledSupplierBalanceAndReminder(t *testing.T) {
	setupR2STestDB(t)

	now := common.GetTimestamp()
	activeSupplier := &R2SSupplier{
		Name:                "Active Supplier",
		DefaultCurrencyCode: "USD",
		DefaultExchangeRate: 2,
		BalanceAmount:       10,
		BalanceCurrencyCode: "USD",
		BalanceUpdatedTime:  now - 2*86400,
		BalanceReminderDays: 1,
	}
	if err := activeSupplier.Insert(); err != nil {
		t.Fatal(err)
	}
	disabledSupplier := &R2SSupplier{
		Name:                "Disabled Supplier",
		Status:              R2SStatusDisabled,
		DefaultCurrencyCode: "USD",
		DefaultExchangeRate: 3,
		BalanceAmount:       100,
		BalanceCurrencyCode: "USD",
		BalanceUpdatedTime:  now - 2*86400,
		BalanceReminderDays: 1,
	}
	if err := disabledSupplier.Insert(); err != nil {
		t.Fatal(err)
	}

	summary, err := GetR2SSummary(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	requireR2SFloat(t, summary.SupplierBalanceAmount, 20)
	if summary.SupplierCount != 2 {
		t.Fatalf("expected total supplier count 2, got %d", summary.SupplierCount)
	}
	if summary.ActiveSupplierCount != 1 {
		t.Fatalf("expected active supplier count 1, got %d", summary.ActiveSupplierCount)
	}
	if summary.ReminderDueCount != 1 {
		t.Fatalf("expected only active supplier reminder due, got %d", summary.ReminderDueCount)
	}
}

func TestDeleteR2SSupplierRemovesUnusedSupplierAndBindings(t *testing.T) {
	setupR2STestDB(t)

	supplier := &R2SSupplier{
		Name:                "Unused Supplier",
		DefaultCurrencyCode: "USD",
		DefaultExchangeRate: 1,
	}
	if err := supplier.Insert(); err != nil {
		t.Fatal(err)
	}
	channel := &Channel{
		Key:    "unused-key",
		Name:   "unused-channel",
		Group:  "default",
		Status: common.ChannelStatusEnabled,
	}
	if err := DB.Create(channel).Error; err != nil {
		t.Fatal(err)
	}
	binding := &R2SChannelBinding{
		SupplierId:      supplier.Id,
		ChannelId:       channel.Id,
		GroupMultiplier: 1,
	}
	if err := binding.Insert(); err != nil {
		t.Fatal(err)
	}

	if err := DeleteR2SSupplier(supplier.Id); err != nil {
		t.Fatal(err)
	}

	var supplierCount int64
	if err := DB.Model(&R2SSupplier{}).Where("id = ?", supplier.Id).Count(&supplierCount).Error; err != nil {
		t.Fatal(err)
	}
	if supplierCount != 0 {
		t.Fatalf("expected supplier to be deleted, got count %d", supplierCount)
	}
	var bindingCount int64
	if err := DB.Model(&R2SChannelBinding{}).Where("supplier_id = ?", supplier.Id).Count(&bindingCount).Error; err != nil {
		t.Fatal(err)
	}
	if bindingCount != 0 {
		t.Fatalf("expected supplier bindings to be deleted, got count %d", bindingCount)
	}
}

func TestDeleteR2SSupplierRefusesSupplierWithHistory(t *testing.T) {
	setupR2STestDB(t)

	supplier := &R2SSupplier{
		Name:                "Historical Supplier",
		DefaultCurrencyCode: "USD",
		DefaultExchangeRate: 1,
		BalanceAmount:       10,
		BalanceCurrencyCode: "USD",
	}
	if err := supplier.Insert(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := CreateR2SPaymentWithBalance(&R2SPayment{
		SupplierId:   supplier.Id,
		PaymentType:  R2SPaymentTypePostpaid,
		Amount:       25,
		CurrencyCode: "USD",
		ExchangeRate: 1,
	}, nil, 100); err != nil {
		t.Fatal(err)
	}

	if err := DeleteR2SSupplier(supplier.Id); err == nil {
		t.Fatal("expected delete to fail for supplier with payment history")
	}

	var supplierCount int64
	if err := DB.Model(&R2SSupplier{}).Where("id = ?", supplier.Id).Count(&supplierCount).Error; err != nil {
		t.Fatal(err)
	}
	if supplierCount != 1 {
		t.Fatalf("expected supplier to remain, got count %d", supplierCount)
	}
}

func TestR2SRecognitionRecordSnapshotsBindingMultiplier(t *testing.T) {
	setupR2STestDB(t)

	supplier := &R2SSupplier{
		Name:                "Upstream C",
		DefaultCurrencyCode: "USD",
		DefaultExchangeRate: 1,
	}
	if err := supplier.Insert(); err != nil {
		t.Fatal(err)
	}
	channel := &Channel{
		Key:    "test-key",
		Name:   "channel-c",
		Group:  "vip,default",
		Status: common.ChannelStatusEnabled,
	}
	if err := DB.Create(channel).Error; err != nil {
		t.Fatal(err)
	}
	binding := &R2SChannelBinding{
		SupplierId:        supplier.Id,
		ChannelId:         channel.Id,
		UpstreamGroupName: "vip",
		GroupMultiplier:   0.42,
	}
	if err := binding.Insert(); err != nil {
		t.Fatal(err)
	}

	record := &R2SRecognitionRecord{
		SupplierId:       supplier.Id,
		ChannelBindingId: binding.Id,
		RevenueAmount:    100,
		CostAmount:       30,
		PeriodStart:      1000,
		PeriodEnd:        2000,
	}
	if err := record.Insert(); err != nil {
		t.Fatal(err)
	}
	requireR2SFloat(t, record.GroupMultiplierSnapshot, 0.42)
	requireR2SFloat(t, record.SystemProfitAmount, 70)
	requireR2SFloat(t, record.ProfitMargin, 70)

	binding.GroupMultiplier = 0.99
	if err := binding.Update(); err != nil {
		t.Fatal(err)
	}

	var stored R2SRecognitionRecord
	if err := DB.First(&stored, record.Id).Error; err != nil {
		t.Fatal(err)
	}
	requireR2SFloat(t, stored.GroupMultiplierSnapshot, 0.42)
	if stored.UpstreamGroupNameSnapshot != "vip" {
		t.Fatalf("expected upstream group snapshot vip, got %q", stored.UpstreamGroupNameSnapshot)
	}
}

func TestSyncR2SRecognitionFromUsageLogsCreatesAndUpdatesHistoricalProfit(t *testing.T) {
	setupR2STestDB(t)

	supplier := &R2SSupplier{
		Name:                "Historical Upstream",
		DefaultCurrencyCode: "USD",
		DefaultExchangeRate: 1,
	}
	if err := supplier.Insert(); err != nil {
		t.Fatal(err)
	}
	channel := &Channel{
		Key:    "history-key",
		Name:   "history-channel",
		Group:  "vip",
		Status: common.ChannelStatusEnabled,
	}
	if err := DB.Create(channel).Error; err != nil {
		t.Fatal(err)
	}
	binding := &R2SChannelBinding{
		SupplierId:        supplier.Id,
		ChannelId:         channel.Id,
		UpstreamGroupName: "vip",
		GroupMultiplier:   0.4,
	}
	if err := binding.Insert(); err != nil {
		t.Fatal(err)
	}

	usageLog := &Log{
		UserId:    1,
		CreatedAt: 1700000000,
		Type:      LogTypeConsume,
		Quota:     int(100 * common.QuotaPerUnit),
		ChannelId: channel.Id,
		Group:     "vip",
	}
	if err := LOG_DB.Create(usageLog).Error; err != nil {
		t.Fatal(err)
	}

	result, err := SyncR2SRecognitionFromUsageLogs(0, 0, 9)
	if err != nil {
		t.Fatal(err)
	}
	if result.CreatedCount != 1 || result.UpdatedCount != 0 || result.SyncedCount != 1 {
		t.Fatalf("unexpected first sync result: %#v", result)
	}

	var record R2SRecognitionRecord
	if err := DB.Where("source_type = ? AND source_reference = ?", R2SRecognitionSourceUsage, "usage_log:1").First(&record).Error; err != nil {
		t.Fatal(err)
	}
	if record.SupplierId != supplier.Id {
		t.Fatalf("expected supplier id %d, got %d", supplier.Id, record.SupplierId)
	}
	if record.ChannelBindingId != binding.Id {
		t.Fatalf("expected binding id %d, got %d", binding.Id, record.ChannelBindingId)
	}
	requireR2SFloat(t, record.RevenueAmount, 100)
	requireR2SFloat(t, record.CostAmount, 40)
	requireR2SFloat(t, record.SystemProfitAmount, 60)
	requireR2SFloat(t, record.ProfitMargin, 60)
	if record.PeriodStart != usageLog.CreatedAt || record.PeriodEnd != usageLog.CreatedAt {
		t.Fatalf("expected period to match usage log time, got %d-%d", record.PeriodStart, record.PeriodEnd)
	}

	binding.GroupMultiplier = 0.55
	if err := binding.Update(); err != nil {
		t.Fatal(err)
	}
	result, err = SyncR2SRecognitionFromUsageLogs(0, 0, 9)
	if err != nil {
		t.Fatal(err)
	}
	if result.CreatedCount != 0 || result.UpdatedCount != 1 || result.SyncedCount != 1 {
		t.Fatalf("unexpected second sync result: %#v", result)
	}

	var count int64
	if err := DB.Model(&R2SRecognitionRecord{}).Where("source_type = ?", R2SRecognitionSourceUsage).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected sync to be idempotent, got %d records", count)
	}
	if err := DB.First(&record, record.Id).Error; err != nil {
		t.Fatal(err)
	}
	requireR2SFloat(t, record.CostAmount, 55)
	requireR2SFloat(t, record.SystemProfitAmount, 45)
	requireR2SFloat(t, record.ProfitMargin, 45)
}

func TestGetR2STrendAggregatesDailyProfitMargin(t *testing.T) {
	setupR2STestDB(t)

	supplier := &R2SSupplier{
		Name:                "Trend Upstream",
		DefaultCurrencyCode: "USD",
		DefaultExchangeRate: 1,
	}
	if err := supplier.Insert(); err != nil {
		t.Fatal(err)
	}
	record := &R2SRecognitionRecord{
		SupplierId:              supplier.Id,
		RevenueAmount:           200,
		CostAmount:              50,
		PeriodStart:             1700000000,
		PeriodEnd:               1700000000,
		GroupMultiplierSnapshot: 1,
	}
	if err := record.Insert(); err != nil {
		t.Fatal(err)
	}

	rows, err := GetR2STrend(1699920000, 1700006400)
	if err != nil {
		t.Fatal(err)
	}
	var found *R2STrendPoint
	for i := range rows {
		if rows[i].RecognizedRevenueAmount > 0 {
			found = &rows[i]
			break
		}
	}
	if found == nil {
		t.Fatal("expected trend to include recognized revenue")
	}
	requireR2SFloat(t, found.RecognizedRevenueAmount, 200)
	requireR2SFloat(t, found.RecognizedCostAmount, 50)
	requireR2SFloat(t, found.RecognizedProfitAmount, 150)
	requireR2SFloat(t, found.ProfitMargin, 75)
}

func TestR2SPromotionProfitabilityUsesRedemptionsAndRecognizedCosts(t *testing.T) {
	setupR2STestDB(t)

	campaign := &TopUpPromotionCampaign{
		Name:         "Summer Profitability",
		CurrencyCode: "USD",
		Status:       common.TopUpPromotionStatusActive,
	}
	if err := DB.Create(campaign).Error; err != nil {
		t.Fatal(err)
	}
	redemption := &TopUpPromotionRedemption{
		CampaignId:         campaign.Id,
		OriginalAmount:     100,
		DiscountAmount:     15,
		FinalPayableAmount: 85,
		CurrencyCode:       "USD",
		Status:             common.TopUpPromotionRedemptionStatusUsed,
		UsedAt:             2000,
	}
	if err := DB.Create(redemption).Error; err != nil {
		t.Fatal(err)
	}

	supplier := &R2SSupplier{
		Name:                "Upstream Promo",
		DefaultCurrencyCode: "USD",
		DefaultExchangeRate: 1,
	}
	if err := supplier.Insert(); err != nil {
		t.Fatal(err)
	}
	record := &R2SRecognitionRecord{
		SupplierId:              supplier.Id,
		PromotionCampaignId:     campaign.Id,
		RevenueAmount:           85,
		CostAmount:              40,
		PeriodStart:             1000,
		PeriodEnd:               3000,
		GroupMultiplierSnapshot: 1,
	}
	if err := record.Insert(); err != nil {
		t.Fatal(err)
	}

	rows, err := GetR2SPromotionProfitability(campaign.Id, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one profitability row, got %d", len(rows))
	}
	row := rows[0]
	if row.CampaignName != campaign.Name {
		t.Fatalf("expected campaign name %q, got %q", campaign.Name, row.CampaignName)
	}
	if row.TopUpCount != 1 {
		t.Fatalf("expected one top-up redemption, got %d", row.TopUpCount)
	}
	if row.SystemCurrencyCode != "USD" {
		t.Fatalf("expected system currency USD, got %q", row.SystemCurrencyCode)
	}
	if !row.ProfitCalculated {
		t.Fatal("expected profit to be calculated for matching currency")
	}
	requireR2SFloat(t, row.GrossRevenueAmount, 100)
	requireR2SFloat(t, row.DiscountAmount, 15)
	requireR2SFloat(t, row.NetRevenueAmount, 85)
	requireR2SFloat(t, row.RecognizedCostAmount, 40)
	requireR2SFloat(t, row.ProfitAmount, 45)
	requireR2SFloat(t, row.ProfitMargin, 52.941176)

	foreignCampaign := &TopUpPromotionCampaign{
		Name:         "USDT Campaign",
		CurrencyCode: "USDT",
		Status:       common.TopUpPromotionStatusActive,
	}
	if err := DB.Create(foreignCampaign).Error; err != nil {
		t.Fatal(err)
	}
	foreignRedemption := &TopUpPromotionRedemption{
		CampaignId:         foreignCampaign.Id,
		OriginalAmount:     50,
		DiscountAmount:     0,
		FinalPayableAmount: 50,
		CurrencyCode:       "USDT",
		Status:             common.TopUpPromotionRedemptionStatusUsed,
		UsedAt:             2000,
	}
	if err := DB.Create(foreignRedemption).Error; err != nil {
		t.Fatal(err)
	}

	rows, err = GetR2SPromotionProfitability(foreignCampaign.Id, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one foreign profitability row, got %d", len(rows))
	}
	if rows[0].ProfitCalculated {
		t.Fatal("expected profit to remain uncalculated for mismatched currency")
	}
	requireR2SFloat(t, rows[0].ProfitAmount, 0)
	requireR2SFloat(t, rows[0].ProfitMargin, 0)
}
