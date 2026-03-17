package operation_setting

import (
	"encoding/json"
	"testing"
)

func TestAmountDiscountRuleUnmarshalJSON(t *testing.T) {
	t.Run("array format", func(t *testing.T) {
		var rule AmountDiscountRule
		if err := json.Unmarshal([]byte(`[5,"coupon_123"]`), &rule); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rule.DiscountAmount != 5 || rule.CouponID != "coupon_123" {
			t.Fatalf("unexpected rule: %+v", rule)
		}
	})

	t.Run("legacy number format", func(t *testing.T) {
		var rule AmountDiscountRule
		if err := json.Unmarshal([]byte(`5`), &rule); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rule.DiscountAmount != 5 || rule.CouponID != "" {
			t.Fatalf("unexpected rule: %+v", rule)
		}
	})
}

func TestAmountDiscountRuleMarshalJSON(t *testing.T) {
	data, err := json.Marshal(AmountDiscountRule{
		DiscountAmount: 5,
		CouponID:       "coupon_123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `[5,"coupon_123"]` {
		t.Fatalf("unexpected json: %s", string(data))
	}
}
