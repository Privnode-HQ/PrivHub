package operation_setting

import "testing"

func TestNormalizeQuotaDisplayTypeAllowsOnlyUSDAndCNY(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty defaults to USD", input: "", expected: QuotaDisplayTypeUSD},
		{name: "usd normalized", input: " usd ", expected: QuotaDisplayTypeUSD},
		{name: "cny normalized", input: "cny", expected: QuotaDisplayTypeCNY},
		{name: "tokens migrates to USD", input: "TOKENS", expected: QuotaDisplayTypeUSD},
		{name: "custom migrates to USD", input: "CUSTOM", expected: QuotaDisplayTypeUSD},
		{name: "unknown migrates to USD", input: "EUR", expected: QuotaDisplayTypeUSD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeQuotaDisplayType(tt.input); got != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}
