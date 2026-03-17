package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
)

func TestApplyTopupDiscount(t *testing.T) {
	tests := []struct {
		name           string
		base           string
		originalAmount int64
		config         float64
		expected       string
	}{
		{
			name:           "no discount config",
			base:           "100",
			originalAmount: 100,
			config:         0,
			expected:       "100",
		},
		{
			name:           "subtract fixed discount amount",
			base:           "100",
			originalAmount: 100,
			config:         15,
			expected:       "85",
		},
		{
			name:           "clamp below zero",
			base:           "10",
			originalAmount: 100,
			config:         20,
			expected:       "0",
		},
	}

	originalDiscount := operation_setting.GetPaymentSetting().AmountDiscount
	defer func() {
		operation_setting.GetPaymentSetting().AmountDiscount = originalDiscount
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation_setting.GetPaymentSetting().AmountDiscount = map[int]float64{}
			if tt.config > 0 {
				operation_setting.GetPaymentSetting().AmountDiscount[int(tt.originalAmount)] = tt.config
			}

			got := applyTopupDiscount(decimal.RequireFromString(tt.base), tt.originalAmount)
			if got.String() != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, got.String())
			}
		})
	}
}

func TestGetStripeMinorUnitAmount(t *testing.T) {
	tests := []struct {
		name             string
		payMoney         string
		stripeUnitPrice  float64
		stripeUnitAmount int64
		expected         int64
	}{
		{
			name:             "usd-like multiplier",
			payMoney:         "19",
			stripeUnitPrice:  8,
			stripeUnitAmount: 800,
			expected:         1900,
		},
		{
			name:             "zero-decimal-like multiplier",
			payMoney:         "795",
			stripeUnitPrice:  800,
			stripeUnitAmount: 800,
			expected:         795,
		},
		{
			name:             "fallback multiplier",
			payMoney:         "12.34",
			stripeUnitPrice:  0,
			stripeUnitAmount: 0,
			expected:         1234,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStripeMinorUnitAmount(
				decimal.RequireFromString(tt.payMoney),
				tt.stripeUnitPrice,
				tt.stripeUnitAmount,
			)
			if got != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}
