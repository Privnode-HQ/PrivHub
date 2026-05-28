package model

import (
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

func TestNormalizeOptionValueMigratesLegacyCurrencyOptions(t *testing.T) {
	if got := normalizeOptionValue("general_setting.quota_display_type", "CUSTOM"); got != operation_setting.QuotaDisplayTypeUSD {
		t.Fatalf("expected CUSTOM to migrate to USD, got %s", got)
	}
	if got := normalizeOptionValue("general_setting.quota_display_type", "cny"); got != operation_setting.QuotaDisplayTypeCNY {
		t.Fatalf("expected cny to normalize to CNY, got %s", got)
	}
	if got := normalizeOptionValue("DisplayInCurrencyEnabled", "false"); got != "true" {
		t.Fatalf("expected legacy display toggle to stay currency enabled, got %s", got)
	}
}

func TestNormalizeOptionValueMigratesLegacyDefaultQuotaPerUnit(t *testing.T) {
	got := normalizeOptionValue("QuotaPerUnit", strconv.FormatFloat(common.LegacyDefaultQuotaPerUnit, 'f', -1, 64))
	expected := strconv.FormatFloat(common.DefaultQuotaPerUnit, 'f', -1, 64)
	if got != expected {
		t.Fatalf("expected legacy quotaPerUnit to migrate to %s, got %s", expected, got)
	}

	custom := "2000000"
	if got = normalizeOptionValue("QuotaPerUnit", custom); got != custom {
		t.Fatalf("expected custom quotaPerUnit to be preserved, got %s", got)
	}
}
