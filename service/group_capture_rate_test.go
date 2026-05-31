package service

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

func TestValidateTrainingDataGroupConsent(t *testing.T) {
	originalCaptureRate := ratio_setting.GroupCaptureRate2JSONString()
	originalAutoGroups := setting.AutoGroups2JsonString()
	originalUserUsableGroups := setting.UserUsableGroups2JSONString()
	t.Cleanup(func() {
		if err := ratio_setting.UpdateGroupCaptureRateByJSONString(originalCaptureRate); err != nil {
			t.Fatalf("restore group capture rate: %v", err)
		}
		if err := setting.UpdateAutoGroupsByJsonString(originalAutoGroups); err != nil {
			t.Fatalf("restore auto groups: %v", err)
		}
		if err := setting.UpdateUserUsableGroupsByJSONString(originalUserUsableGroups); err != nil {
			t.Fatalf("restore user usable groups: %v", err)
		}
	})

	if err := ratio_setting.UpdateGroupCaptureRateByJSONString(`{"captured":0.25,"auto_high":0.5,"safe":0}`); err != nil {
		t.Fatalf("set group capture rates: %v", err)
	}
	if err := setting.UpdateUserUsableGroupsByJSONString(`{"captured":"Captured","auto_high":"Auto High","safe":"Safe"}`); err != nil {
		t.Fatalf("set user usable groups: %v", err)
	}
	if err := setting.UpdateAutoGroupsByJsonString(`["safe","auto_high"]`); err != nil {
		t.Fatalf("set auto groups: %v", err)
	}

	if err := ValidateTrainingDataGroupConsent("safe", []string{"safe"}, dto.UserSetting{}); err != nil {
		t.Fatalf("safe group rejected without consent: %v", err)
	}
	if err := ValidateTrainingDataGroupConsent("safe", []string{"captured"}, dto.UserSetting{}); err == nil {
		t.Fatal("captured group accepted without consent")
	} else if !strings.Contains(err.Error(), "25.00%") {
		t.Fatalf("captured group error = %q, want rate in message", err.Error())
	}
	if err := ValidateTrainingDataGroupConsent("captured", nil, dto.UserSetting{}); err == nil {
		t.Fatal("captured default group accepted without consent")
	}
	if err := ValidateTrainingDataGroupConsent("safe", []string{"captured"}, dto.UserSetting{AllowTrainingDataGroups: true}); err != nil {
		t.Fatalf("captured group rejected with consent: %v", err)
	}

	if got := GetGroupCaptureRateForSelection("safe", "auto"); got != 0.5 {
		t.Fatalf("GetGroupCaptureRateForSelection(auto) = %v, want 0.5", got)
	}
	if err := ValidateTrainingDataGroupConsent("safe", []string{"auto"}, dto.UserSetting{}); err == nil {
		t.Fatal("auto group with captured fallback accepted without consent")
	}
}
