package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v81"
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
			operation_setting.GetPaymentSetting().AmountDiscount = map[int]operation_setting.AmountDiscountRule{}
			if tt.config > 0 {
				operation_setting.GetPaymentSetting().AmountDiscount[int(tt.originalAmount)] = operation_setting.AmountDiscountRule{
					DiscountAmount: tt.config,
				}
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
		name     string
		payMoney string
		currency stripe.Currency
		expected int64
	}{
		{
			name:     "two-decimal currency",
			payMoney: "19.45",
			currency: stripe.CurrencyUSD,
			expected: 1945,
		},
		{
			name:     "zero-decimal currency",
			payMoney: "795",
			currency: stripe.CurrencyJPY,
			expected: 795,
		},
		{
			name:     "round half up",
			payMoney: "12.345",
			currency: stripe.CurrencyUSD,
			expected: 1235,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStripeMinorUnitAmount(decimal.RequireFromString(tt.payMoney), tt.currency)
			if got != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestGetStripeMajorUnitAmount(t *testing.T) {
	tests := []struct {
		name        string
		minorAmount int64
		currency    stripe.Currency
		expected    string
	}{
		{
			name:        "two-decimal currency",
			minorAmount: 9500,
			currency:    stripe.CurrencyUSD,
			expected:    "95",
		},
		{
			name:        "zero-decimal currency",
			minorAmount: 1898,
			currency:    stripe.CurrencyJPY,
			expected:    "1898",
		},
		{
			name:        "invalid amount returns zero",
			minorAmount: 0,
			currency:    stripe.CurrencyUSD,
			expected:    "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStripeMajorUnitAmount(tt.minorAmount, tt.currency)
			if got.String() != tt.expected {
				t.Fatalf("expected major amount %s, got %s", tt.expected, got.String())
			}
		})
	}
}

func TestGetStripeCheckoutQuantity(t *testing.T) {
	originalDisplayType := operation_setting.GetQuotaDisplayType()
	originalQuotaPerUnit := common.QuotaPerUnit
	defer func() {
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalDisplayType
		common.QuotaPerUnit = originalQuotaPerUnit
	}()

	tests := []struct {
		name         string
		displayType  string
		quotaPerUnit float64
		amount       int64
		expected     int64
		expectError  bool
	}{
		{
			name:         "currency display keeps integer amount",
			displayType:  operation_setting.QuotaDisplayTypeUSD,
			quotaPerUnit: 500000,
			amount:       20,
			expected:     20,
		},
		{
			name:         "tokens display converts to stripe quantity",
			displayType:  operation_setting.QuotaDisplayTypeTokens,
			quotaPerUnit: 500000,
			amount:       1000000,
			expected:     2,
		},
		{
			name:         "tokens display requires whole stripe unit",
			displayType:  operation_setting.QuotaDisplayTypeTokens,
			quotaPerUnit: 500000,
			amount:       750000,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation_setting.GetGeneralSetting().QuotaDisplayType = tt.displayType
			common.QuotaPerUnit = tt.quotaPerUnit

			got, err := getStripeCheckoutQuantity(tt.amount)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("expected quantity %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestResolveTopUpBasePayable(t *testing.T) {
	tests := []struct {
		name                  string
		original              string
		discounted            string
		platformDiscount      string
		hasSelectedUserCoupon bool
		expectedBase          string
		expectedPlatform      string
	}{
		{
			name:             "keep platform discount when no eligible user coupon",
			original:         "100",
			discounted:       "95",
			platformDiscount: "5",
			expectedBase:     "95",
			expectedPlatform: "5",
		},
		{
			name:                  "selected user coupon overrides platform discount",
			original:              "100",
			discounted:            "95",
			platformDiscount:      "5",
			hasSelectedUserCoupon: true,
			expectedBase:          "100",
			expectedPlatform:      "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBase, gotPlatform := resolveTopUpBasePayable(
				decimal.RequireFromString(tt.original),
				decimal.RequireFromString(tt.discounted),
				decimal.RequireFromString(tt.platformDiscount),
				tt.hasSelectedUserCoupon,
			)
			if gotBase.String() != tt.expectedBase {
				t.Fatalf("expected base %s, got %s", tt.expectedBase, gotBase.String())
			}
			if gotPlatform.String() != tt.expectedPlatform {
				t.Fatalf("expected platform %s, got %s", tt.expectedPlatform, gotPlatform.String())
			}
		})
	}
}
