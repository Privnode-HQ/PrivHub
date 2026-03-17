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
	RPM     *int64 `json:"rpm"`
	RPD     *int64 `json:"rpd"`
	TPM     *int64 `json:"tpm"`
	TPD     *int64 `json:"tpd"`
	Monthly *int64 `json:"monthly"`
}

type groupUsageLimitPolicyInput struct {
	RPM     *int64 `json:"rpm"`
	RPD     *int64 `json:"rpd"`
	TPM     *int64 `json:"tpm"`
	TPD     *int64 `json:"tpd"`
	Monthly *int64 `json:"monthly"`
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
		RPM:     cloneNullableInt64(policy.RPM),
		RPD:     cloneNullableInt64(policy.RPD),
		TPM:     cloneNullableInt64(policy.TPM),
		TPD:     cloneNullableInt64(policy.TPD),
		Monthly: cloneNullableInt64(policy.Monthly),
	}
}

func normalizeMonthlyDisplayValue(value int64) (int64, error) {
	if value < 0 {
		return 0, fmt.Errorf("monthly must be null or a non-negative integer")
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

func normalizeMonthlyQuotaForDisplay(value int64) int64 {
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
		if err := validateNonNegativeNullableInt(policyInput.Monthly, fmt.Sprintf("group %s monthly", groupName)); err != nil {
			return nil, err
		}

		policy := GroupUsageLimitPolicy{
			RPM: cloneNullableInt64(policyInput.RPM),
			RPD: cloneNullableInt64(policyInput.RPD),
			TPM: cloneNullableInt64(policyInput.TPM),
			TPD: cloneNullableInt64(policyInput.TPD),
		}
		if policyInput.Monthly != nil {
			normalizedMonthly, err := normalizeMonthlyDisplayValue(*policyInput.Monthly)
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
			RPM: cloneNullableInt64(policy.RPM),
			RPD: cloneNullableInt64(policy.RPD),
			TPM: cloneNullableInt64(policy.TPM),
			TPD: cloneNullableInt64(policy.TPD),
		}
		if policy.Monthly != nil {
			displayMonthly := normalizeMonthlyQuotaForDisplay(*policy.Monthly)
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
