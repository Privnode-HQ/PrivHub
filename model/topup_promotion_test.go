package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"

	"github.com/shopspring/decimal"
)

func TestCalculateTopUpPromotionDiscountUsesBestMatchingTier(t *testing.T) {
	rules := []TopUpPromotionRule{
		{
			MinAmount:     100,
			DiscountType:  common.TopUpPromotionDiscountTypeFixed,
			DiscountValue: 20,
		},
		{
			MinAmount:     50,
			DiscountType:  common.TopUpPromotionDiscountTypeFixed,
			DiscountValue: 5,
		},
	}

	discount, rule, ok := CalculateTopUpPromotionDiscount(rules, decimal.RequireFromString("120"))
	if !ok {
		t.Fatal("expected promotion rule to match")
	}
	if !discount.Equal(decimal.RequireFromString("20")) {
		t.Fatalf("expected fixed discount 20, got %s", discount.String())
	}
	if rule.MinAmount != 100 {
		t.Fatalf("expected highest matching tier, got min %.2f", rule.MinAmount)
	}
}

func TestCalculateTopUpPromotionDiscountSupportsPercent(t *testing.T) {
	rules := []TopUpPromotionRule{
		{
			MinAmount:     80,
			DiscountType:  common.TopUpPromotionDiscountTypePercent,
			DiscountValue: 12.5,
		},
	}

	discount, _, ok := CalculateTopUpPromotionDiscount(rules, decimal.RequireFromString("88"))
	if !ok {
		t.Fatal("expected promotion rule to match")
	}
	if !discount.Equal(decimal.RequireFromString("11")) {
		t.Fatalf("expected percent discount 11, got %s", discount.String())
	}
}
