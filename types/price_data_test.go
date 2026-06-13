package types

import "testing"

func TestModelRatioPricingConversion(t *testing.T) {
	modelRatio := 2.5

	if got := ModelRatioInputPricePerMillionTokens(modelRatio); got != 5 {
		t.Fatalf("expected input price 5 per 1M tokens, got %v", got)
	}

	if got := ModelRatioTokenPrice(modelRatio); got != 0.000005 {
		t.Fatalf("expected token price 0.000005, got %v", got)
	}

	if got := ModelRatioTokenQuotaRatio(modelRatio, 1000000); got != 5 {
		t.Fatalf("expected token quota ratio 5, got %v", got)
	}

	if got := ModelRatioTokenQuotaRatio(modelRatio, 500000); got != 2.5 {
		t.Fatalf("expected custom quota unit to scale raw quota, got %v", got)
	}
}

func TestModelRatioPricingMatchesBillingBreakdown(t *testing.T) {
	effectiveTokens := 94.0 + 157939.0*1.25 + 126.0*5.0
	quota := int(effectiveTokens *
		ModelRatioTokenQuotaRatio(2.5, 1000000) *
		1.6)

	if quota != 1585182 {
		t.Fatalf("expected quota 1585182, got %d", quota)
	}
}
