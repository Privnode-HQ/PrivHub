package model

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupPaidQuotaBreakdownTestDB(t *testing.T) {
	t.Helper()

	originalDB := DB
	originalQuotaPerUnit := common.QuotaPerUnit
	t.Cleanup(func() {
		DB = originalDB
		common.QuotaPerUnit = originalQuotaPerUnit
	})

	common.QuotaPerUnit = common.DefaultQuotaPerUnit
	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&User{}, &TopUp{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	DB = db
}

func TestCalculateTopUpActualPaidQuotaExcludesDiscountsAndFees(t *testing.T) {
	originalQuotaPerUnit := common.QuotaPerUnit
	t.Cleanup(func() {
		common.QuotaPerUnit = originalQuotaPerUnit
	})
	common.QuotaPerUnit = common.DefaultQuotaPerUnit

	got := calculateTopUpActualPaidQuota(topUpPaidQuotaRow{
		PaymentMethod: "alipay",
		Amount:        10,
		OriginalMoney: 10,
		PayMoney:      7.35,
		ProcessingFee: 0.35,
	})
	if got != 7_000_000 {
		t.Fatalf("expected paid quota 7000000, got %d", got)
	}
}

func TestCalculateRemainingPaidQuotaBreakdownConsumesPaidFirst(t *testing.T) {
	totalQuota, remainPaidQuota, remainNonPaidQuota := calculateRemainingPaidQuotaBreakdown(3_000_000, 12_000_000, 10_000_000)
	if totalQuota != 15_000_000 {
		t.Fatalf("expected total quota 15000000, got %d", totalQuota)
	}
	if remainPaidQuota != 0 {
		t.Fatalf("expected paid quota to be exhausted first, got %d", remainPaidQuota)
	}
	if remainNonPaidQuota != 3_000_000 {
		t.Fatalf("expected remaining non-paid quota 3000000, got %d", remainNonPaidQuota)
	}
}

func TestGetUserPaidQuotaBreakdownByCAHID(t *testing.T) {
	setupPaidQuotaBreakdownTestDB(t)

	user := &User{
		Username:  "paid-breakdown",
		Password:  "irrelevant",
		Quota:     11_000_000,
		UsedQuota: 4_000_000,
	}
	if err := DB.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	topUps := []TopUp{
		{
			UserId:        user.Id,
			TradeNo:       "discounted-epay",
			PaymentMethod: "alipay",
			Status:        common.TopUpStatusSuccess,
			Amount:        10,
			OriginalMoney: 10,
			PayMoney:      7.35,
			ProcessingFee: 0.35,
		},
		{
			UserId:        user.Id,
			TradeNo:       "creem-full",
			PaymentMethod: "creem",
			Status:        common.TopUpStatusSuccess,
			Amount:        2_000_000,
			OriginalMoney: 2,
			PayMoney:      2,
		},
	}
	if err := DB.Create(&topUps).Error; err != nil {
		t.Fatalf("create topups: %v", err)
	}

	breakdown, err := GetUserPaidQuotaBreakdownByCAHID(user.CAHID)
	if err != nil {
		t.Fatalf("get breakdown: %v", err)
	}
	if breakdown.TotalQuota != 15_000_000 {
		t.Fatalf("expected total quota 15000000, got %d", breakdown.TotalQuota)
	}
	if breakdown.RemainQuota != 11_000_000 {
		t.Fatalf("expected remain quota 11000000, got %d", breakdown.RemainQuota)
	}
	if breakdown.TotalPaidQuota != 9_000_000 {
		t.Fatalf("expected total paid quota 9000000, got %d", breakdown.TotalPaidQuota)
	}
	if breakdown.RemainPaidQuota != 5_000_000 {
		t.Fatalf("expected remain paid quota 5000000, got %d", breakdown.RemainPaidQuota)
	}
	if breakdown.RemainNonPaidQuota != 6_000_000 {
		t.Fatalf("expected remain non-paid quota 6000000, got %d", breakdown.RemainNonPaidQuota)
	}
}
