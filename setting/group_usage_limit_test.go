package setting

import (
	"encoding/json"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

func TestUserGroupUsageLimitsRoundTripSupportsWeeklyAndHideDetails(t *testing.T) {
	originalPolicies := UserGroupUsageLimits2JSONString()
	t.Cleanup(func() {
		if err := UpdateUserGroupUsageLimitsByJSONString(originalPolicies); err != nil {
			t.Errorf("restore usage limit policies: %v", err)
		}
	})

	input := `{"default":{"rpm":5,"rpm_hide_details":true,"rpd":20,"tpm":2000,"tpd":10000,"hourly":300,"daily":900,"weekly":1800,"weekly_hide_details":true,"monthly":3000}}`
	if err := UpdateUserGroupUsageLimitsByJSONString(input); err != nil {
		t.Fatalf("UpdateUserGroupUsageLimitsByJSONString returned error: %v", err)
	}

	policy, found := GetUserGroupUsageLimit("default")
	if !found {
		t.Fatalf("expected default group policy to exist")
	}

	checkConfigured := func(name string, value *int64) {
		t.Helper()
		if value == nil {
			t.Fatalf("expected %s to be configured", name)
		}
	}

	checkConfigured("rpm", policy.RPM)
	checkConfigured("rpd", policy.RPD)
	checkConfigured("tpm", policy.TPM)
	checkConfigured("tpd", policy.TPD)
	checkConfigured("hourly", policy.Hourly)
	checkConfigured("daily", policy.Daily)
	checkConfigured("weekly", policy.Weekly)
	checkConfigured("monthly", policy.Monthly)
	if !policy.RPMHideDetails {
		t.Fatalf("expected rpm_hide_details to be enabled")
	}
	if !policy.WeeklyHideDetails {
		t.Fatalf("expected weekly_hide_details to be enabled")
	}

	serialized := UserGroupUsageLimits2JSONString()
	if err := CheckUserGroupUsageLimits(serialized); err != nil {
		t.Fatalf("expected serialized policy to remain valid, got error: %v", err)
	}

	var serializedPolicies map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(serialized), &serializedPolicies); err != nil {
		t.Fatalf("unmarshal serialized policies: %v", err)
	}
	groupPolicy := serializedPolicies["default"]
	if groupPolicy["rpm"] != float64(5) || groupPolicy["rpd"] != float64(20) || groupPolicy["tpm"] != float64(2000) || groupPolicy["tpd"] != float64(10000) {
		t.Fatalf("unexpected serialized request/token limits: %#v", groupPolicy)
	}
	if groupPolicy["hourly"] != float64(300) || groupPolicy["daily"] != float64(900) || groupPolicy["weekly"] != float64(1800) || groupPolicy["monthly"] != float64(3000) {
		t.Fatalf("unexpected serialized budget limits: %#v", groupPolicy)
	}
	if groupPolicy["rpm_hide_details"] != true || groupPolicy["weekly_hide_details"] != true {
		t.Fatalf("unexpected serialized hide-details flags: %#v", groupPolicy)
	}
}

func TestUserGroupUsageLimitBudgetUsesDisplayCurrencyWithoutExchangeRate(t *testing.T) {
	originalPolicies := UserGroupUsageLimits2JSONString()
	originalDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	originalQuotaPerUnit := common.QuotaPerUnit
	t.Cleanup(func() {
		if err := UpdateUserGroupUsageLimitsByJSONString(originalPolicies); err != nil {
			t.Errorf("restore usage limit policies: %v", err)
		}
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalDisplayType
		common.QuotaPerUnit = originalQuotaPerUnit
	})

	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeCNY
	common.QuotaPerUnit = common.DefaultQuotaPerUnit

	if err := UpdateUserGroupUsageLimitsByJSONString(`{"default":{"hourly":2,"daily":3}}`); err != nil {
		t.Fatalf("UpdateUserGroupUsageLimitsByJSONString returned error: %v", err)
	}

	policy, found := GetUserGroupUsageLimit("default")
	if !found {
		t.Fatalf("expected default group policy to exist")
	}
	if policy.Hourly == nil || *policy.Hourly != int64(2*common.DefaultQuotaPerUnit) {
		t.Fatalf("expected hourly to use quotaPerUnit without exchange, got %#v", policy.Hourly)
	}
	if policy.Daily == nil || *policy.Daily != int64(3*common.DefaultQuotaPerUnit) {
		t.Fatalf("expected daily to use quotaPerUnit without exchange, got %#v", policy.Daily)
	}

	var serialized map[string]map[string]int64
	if err := json.Unmarshal([]byte(UserGroupUsageLimits2JSONString()), &serialized); err != nil {
		t.Fatalf("unmarshal serialized policies: %v", err)
	}
	if serialized["default"]["hourly"] != 2 || serialized["default"]["daily"] != 3 {
		t.Fatalf("expected display values to round trip without exchange, got %#v", serialized)
	}
}
