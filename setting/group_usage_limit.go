package setting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
)

type GroupUsageLimitPolicy struct {
	RPM                *int64 `json:"rpm"`
	RPMHideDetails     bool   `json:"rpm_hide_details,omitempty"`
	RPD                *int64 `json:"rpd"`
	RPDHideDetails     bool   `json:"rpd_hide_details,omitempty"`
	TPM                *int64 `json:"tpm"`
	TPMHideDetails     bool   `json:"tpm_hide_details,omitempty"`
	TPD                *int64 `json:"tpd"`
	TPDHideDetails     bool   `json:"tpd_hide_details,omitempty"`
	Hourly             *int64 `json:"hourly"`
	HourlyHideDetails  bool   `json:"hourly_hide_details,omitempty"`
	Daily              *int64 `json:"daily"`
	DailyHideDetails   bool   `json:"daily_hide_details,omitempty"`
	Weekly             *int64 `json:"weekly"`
	WeeklyHideDetails  bool   `json:"weekly_hide_details,omitempty"`
	Monthly            *int64 `json:"monthly"`
	MonthlyHideDetails bool   `json:"monthly_hide_details,omitempty"`
}

type groupUsageLimitPolicyInput struct {
	RPM                *int64 `json:"rpm"`
	RPMHideDetails     bool   `json:"rpm_hide_details,omitempty"`
	RPD                *int64 `json:"rpd"`
	RPDHideDetails     bool   `json:"rpd_hide_details,omitempty"`
	TPM                *int64 `json:"tpm"`
	TPMHideDetails     bool   `json:"tpm_hide_details,omitempty"`
	TPD                *int64 `json:"tpd"`
	TPDHideDetails     bool   `json:"tpd_hide_details,omitempty"`
	Hourly             *int64 `json:"hourly"`
	HourlyHideDetails  bool   `json:"hourly_hide_details,omitempty"`
	Daily              *int64 `json:"daily"`
	DailyHideDetails   bool   `json:"daily_hide_details,omitempty"`
	Weekly             *int64 `json:"weekly"`
	WeeklyHideDetails  bool   `json:"weekly_hide_details,omitempty"`
	Monthly            *int64 `json:"monthly"`
	MonthlyHideDetails bool   `json:"monthly_hide_details,omitempty"`
}

var userGroupUsageLimits = map[string]GroupUsageLimitPolicy{}
var userGroupUsageLimitsMutex sync.RWMutex

func cloneNullableInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneGroupUsageLimitPolicy(policy GroupUsageLimitPolicy) GroupUsageLimitPolicy {
	return GroupUsageLimitPolicy{
		RPM:                cloneNullableInt64(policy.RPM),
		RPMHideDetails:     policy.RPMHideDetails,
		RPD:                cloneNullableInt64(policy.RPD),
		RPDHideDetails:     policy.RPDHideDetails,
		TPM:                cloneNullableInt64(policy.TPM),
		TPMHideDetails:     policy.TPMHideDetails,
		TPD:                cloneNullableInt64(policy.TPD),
		TPDHideDetails:     policy.TPDHideDetails,
		Hourly:             cloneNullableInt64(policy.Hourly),
		HourlyHideDetails:  policy.HourlyHideDetails,
		Daily:              cloneNullableInt64(policy.Daily),
		DailyHideDetails:   policy.DailyHideDetails,
		Weekly:             cloneNullableInt64(policy.Weekly),
		WeeklyHideDetails:  policy.WeeklyHideDetails,
		Monthly:            cloneNullableInt64(policy.Monthly),
		MonthlyHideDetails: policy.MonthlyHideDetails,
	}
}

func (policy GroupUsageLimitPolicy) ShouldHideMetricDetails(metric string) bool {
	switch metric {
	case "rpm":
		return policy.RPMHideDetails
	case "rpd":
		return policy.RPDHideDetails
	case "tpm":
		return policy.TPMHideDetails
	case "tpd":
		return policy.TPDHideDetails
	case "hourly":
		return policy.HourlyHideDetails
	case "daily":
		return policy.DailyHideDetails
	case "weekly":
		return policy.WeeklyHideDetails
	case "monthly":
		return policy.MonthlyHideDetails
	default:
		return false
	}
}

func normalizeBudgetDisplayValue(value int64, field string) (int64, error) {
	if value < 0 {
		return 0, fmt.Errorf("%s must be null or a non-negative integer", field)
	}
	switch operation_setting.GetQuotaDisplayType() {
	case operation_setting.QuotaDisplayTypeTokens:
		return value, nil
	case operation_setting.QuotaDisplayTypeCNY:
		rate := decimal.NewFromFloat(operation_setting.USDExchangeRate)
		if rate.LessThanOrEqual(decimal.Zero) {
			return 0, fmt.Errorf("usd exchange rate must be greater than 0")
		}
		return decimal.NewFromInt(value).
			Div(rate).
			Mul(decimal.NewFromFloat(common.QuotaPerUnit)).
			Round(0).
			IntPart(), nil
	case operation_setting.QuotaDisplayTypeCustom:
		rate := decimal.NewFromFloat(operation_setting.GetUsdToCurrencyRate(operation_setting.USDExchangeRate))
		if rate.LessThanOrEqual(decimal.Zero) {
			return 0, fmt.Errorf("custom currency exchange rate must be greater than 0")
		}
		return decimal.NewFromInt(value).
			Div(rate).
			Mul(decimal.NewFromFloat(common.QuotaPerUnit)).
			Round(0).
			IntPart(), nil
	default:
		return decimal.NewFromInt(value).
			Mul(decimal.NewFromFloat(common.QuotaPerUnit)).
			Round(0).
			IntPart(), nil
	}
}

func normalizeBudgetQuotaForDisplay(value int64) int64 {
	if value <= 0 {
		return 0
	}
	switch operation_setting.GetQuotaDisplayType() {
	case operation_setting.QuotaDisplayTypeTokens:
		return value
	case operation_setting.QuotaDisplayTypeCNY:
		return decimal.NewFromInt(value).
			Div(decimal.NewFromFloat(common.QuotaPerUnit)).
			Mul(decimal.NewFromFloat(operation_setting.USDExchangeRate)).
			Round(0).
			IntPart()
	case operation_setting.QuotaDisplayTypeCustom:
		rate := operation_setting.GetUsdToCurrencyRate(operation_setting.USDExchangeRate)
		return decimal.NewFromInt(value).
			Div(decimal.NewFromFloat(common.QuotaPerUnit)).
			Mul(decimal.NewFromFloat(rate)).
			Round(0).
			IntPart()
	default:
		return decimal.NewFromInt(value).
			Div(decimal.NewFromFloat(common.QuotaPerUnit)).
			Round(0).
			IntPart()
	}
}

func validateNonNegativeNullableInt(value *int64, field string) error {
	if value == nil {
		return nil
	}
	if *value < 0 {
		return fmt.Errorf("%s must be null or a non-negative integer", field)
	}
	if *value > math.MaxInt32 {
		return fmt.Errorf("%s exceeds the maximum supported value %d", field, math.MaxInt32)
	}
	return nil
}

func parseGroupUsageLimitPolicies(jsonStr string) (map[string]GroupUsageLimitPolicy, error) {
	trimmed := strings.TrimSpace(jsonStr)
	if trimmed == "" {
		return nil, fmt.Errorf("user group usage limits cannot be empty")
	}

	rawPolicies := make(map[string]json.RawMessage)
	if err := json.Unmarshal([]byte(trimmed), &rawPolicies); err != nil {
		return nil, err
	}

	policies := make(map[string]GroupUsageLimitPolicy, len(rawPolicies))
	for groupName, rawPolicy := range rawPolicies {
		groupName = strings.TrimSpace(groupName)
		if groupName == "" {
			return nil, fmt.Errorf("group name cannot be empty")
		}

		decoder := json.NewDecoder(bytes.NewReader(rawPolicy))
		decoder.DisallowUnknownFields()

		var policyInput groupUsageLimitPolicyInput
		if err := decoder.Decode(&policyInput); err != nil {
			return nil, fmt.Errorf("group %s policy is invalid: %w", groupName, err)
		}

		if err := validateNonNegativeNullableInt(policyInput.RPM, fmt.Sprintf("group %s rpm", groupName)); err != nil {
			return nil, err
		}
		if err := validateNonNegativeNullableInt(policyInput.RPD, fmt.Sprintf("group %s rpd", groupName)); err != nil {
			return nil, err
		}
		if err := validateNonNegativeNullableInt(policyInput.TPM, fmt.Sprintf("group %s tpm", groupName)); err != nil {
			return nil, err
		}
		if err := validateNonNegativeNullableInt(policyInput.TPD, fmt.Sprintf("group %s tpd", groupName)); err != nil {
			return nil, err
		}
		if err := validateNonNegativeNullableInt(policyInput.Hourly, fmt.Sprintf("group %s hourly", groupName)); err != nil {
			return nil, err
		}
		if err := validateNonNegativeNullableInt(policyInput.Daily, fmt.Sprintf("group %s daily", groupName)); err != nil {
			return nil, err
		}
		if err := validateNonNegativeNullableInt(policyInput.Weekly, fmt.Sprintf("group %s weekly", groupName)); err != nil {
			return nil, err
		}
		if err := validateNonNegativeNullableInt(policyInput.Monthly, fmt.Sprintf("group %s monthly", groupName)); err != nil {
			return nil, err
		}

		policy := GroupUsageLimitPolicy{
			RPM:                cloneNullableInt64(policyInput.RPM),
			RPMHideDetails:     policyInput.RPMHideDetails,
			RPD:                cloneNullableInt64(policyInput.RPD),
			RPDHideDetails:     policyInput.RPDHideDetails,
			TPM:                cloneNullableInt64(policyInput.TPM),
			TPMHideDetails:     policyInput.TPMHideDetails,
			TPD:                cloneNullableInt64(policyInput.TPD),
			TPDHideDetails:     policyInput.TPDHideDetails,
			HourlyHideDetails:  policyInput.HourlyHideDetails,
			DailyHideDetails:   policyInput.DailyHideDetails,
			WeeklyHideDetails:  policyInput.WeeklyHideDetails,
			MonthlyHideDetails: policyInput.MonthlyHideDetails,
		}
		if policyInput.Hourly != nil {
			normalizedHourly, err := normalizeBudgetDisplayValue(*policyInput.Hourly, "hourly")
			if err != nil {
				return nil, fmt.Errorf("group %s hourly is invalid: %w", groupName, err)
			}
			policy.Hourly = &normalizedHourly
		}
		if policyInput.Daily != nil {
			normalizedDaily, err := normalizeBudgetDisplayValue(*policyInput.Daily, "daily")
			if err != nil {
				return nil, fmt.Errorf("group %s daily is invalid: %w", groupName, err)
			}
			policy.Daily = &normalizedDaily
		}
		if policyInput.Weekly != nil {
			normalizedWeekly, err := normalizeBudgetDisplayValue(*policyInput.Weekly, "weekly")
			if err != nil {
				return nil, fmt.Errorf("group %s weekly is invalid: %w", groupName, err)
			}
			policy.Weekly = &normalizedWeekly
		}
		if policyInput.Monthly != nil {
			normalizedMonthly, err := normalizeBudgetDisplayValue(*policyInput.Monthly, "monthly")
			if err != nil {
				return nil, fmt.Errorf("group %s monthly is invalid: %w", groupName, err)
			}
			policy.Monthly = &normalizedMonthly
		}

		policies[groupName] = policy
	}
	return policies, nil
}

func UserGroupUsageLimits2JSONString() string {
	userGroupUsageLimitsMutex.RLock()
	defer userGroupUsageLimitsMutex.RUnlock()

	serialized := make(map[string]groupUsageLimitPolicyInput, len(userGroupUsageLimits))
	for groupName, policy := range userGroupUsageLimits {
		displayPolicy := groupUsageLimitPolicyInput{
			RPM:                cloneNullableInt64(policy.RPM),
			RPMHideDetails:     policy.RPMHideDetails,
			RPD:                cloneNullableInt64(policy.RPD),
			RPDHideDetails:     policy.RPDHideDetails,
			TPM:                cloneNullableInt64(policy.TPM),
			TPMHideDetails:     policy.TPMHideDetails,
			TPD:                cloneNullableInt64(policy.TPD),
			TPDHideDetails:     policy.TPDHideDetails,
			HourlyHideDetails:  policy.HourlyHideDetails,
			DailyHideDetails:   policy.DailyHideDetails,
			WeeklyHideDetails:  policy.WeeklyHideDetails,
			MonthlyHideDetails: policy.MonthlyHideDetails,
		}
		if policy.Hourly != nil {
			displayHourly := normalizeBudgetQuotaForDisplay(*policy.Hourly)
			displayPolicy.Hourly = &displayHourly
		}
		if policy.Daily != nil {
			displayDaily := normalizeBudgetQuotaForDisplay(*policy.Daily)
			displayPolicy.Daily = &displayDaily
		}
		if policy.Weekly != nil {
			displayWeekly := normalizeBudgetQuotaForDisplay(*policy.Weekly)
			displayPolicy.Weekly = &displayWeekly
		}
		if policy.Monthly != nil {
			displayMonthly := normalizeBudgetQuotaForDisplay(*policy.Monthly)
			displayPolicy.Monthly = &displayMonthly
		}
		serialized[groupName] = displayPolicy
	}

	jsonBytes, err := json.Marshal(serialized)
	if err != nil {
		common.SysLog("error marshalling user group usage limits: " + err.Error())
		return "{}"
	}
	return string(jsonBytes)
}

func UpdateUserGroupUsageLimitsByJSONString(jsonStr string) error {
	policies, err := parseGroupUsageLimitPolicies(jsonStr)
	if err != nil {
		return err
	}

	userGroupUsageLimitsMutex.Lock()
	defer userGroupUsageLimitsMutex.Unlock()

	userGroupUsageLimits = policies
	return nil
}

func CheckUserGroupUsageLimits(jsonStr string) error {
	_, err := parseGroupUsageLimitPolicies(jsonStr)
	return err
}

func GetUserGroupUsageLimit(group string) (GroupUsageLimitPolicy, bool) {
	userGroupUsageLimitsMutex.RLock()
	defer userGroupUsageLimitsMutex.RUnlock()

	policy, found := userGroupUsageLimits[group]
	if !found {
		return GroupUsageLimitPolicy{}, false
	}
	return cloneGroupUsageLimitPolicy(policy), true
}

func GetUserGroupUsageLimitsCopy() map[string]GroupUsageLimitPolicy {
	userGroupUsageLimitsMutex.RLock()
	defer userGroupUsageLimitsMutex.RUnlock()

	copyPolicies := make(map[string]GroupUsageLimitPolicy, len(userGroupUsageLimits))
	for groupName, policy := range userGroupUsageLimits {
		copyPolicies[groupName] = cloneGroupUsageLimitPolicy(policy)
	}
	return copyPolicies
}
