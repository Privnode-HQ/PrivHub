package setting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"

	"github.com/QuantumNous/new-api/common"
)

const (
	UsageLimitTargetAll    = "all"
	UsageLimitTargetGroups = "groups"
	UsageLimitTargetUsers  = "users"
)

var validUsageLimitMetrics = map[string]struct{}{
	"rpm":     {},
	"rpd":     {},
	"tpm":     {},
	"tpd":     {},
	"hourly":  {},
	"daily":   {},
	"weekly":  {},
	"monthly": {},
}

type UserUsageLimitMultiplierRule struct {
	Scope      string   `json:"scope"`
	GroupNames []string `json:"group_names,omitempty"`
	UserIDs    []int    `json:"user_ids,omitempty"`
	Metrics    []string `json:"metrics"`
	Multiplier float64  `json:"multiplier"`
}

var userUsageLimitMultiplierRules []UserUsageLimitMultiplierRule
var userUsageLimitMultiplierRulesMutex sync.RWMutex

func normalizeStringList(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeIntList(values []int) []int {
	normalized := make([]int, 0, len(values))
	seen := make(map[int]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func cloneUserUsageLimitMultiplierRule(rule UserUsageLimitMultiplierRule) UserUsageLimitMultiplierRule {
	return UserUsageLimitMultiplierRule{
		Scope:      rule.Scope,
		GroupNames: slices.Clone(rule.GroupNames),
		UserIDs:    slices.Clone(rule.UserIDs),
		Metrics:    slices.Clone(rule.Metrics),
		Multiplier: rule.Multiplier,
	}
}

func cloneUserUsageLimitMultiplierRules(rules []UserUsageLimitMultiplierRule) []UserUsageLimitMultiplierRule {
	cloned := make([]UserUsageLimitMultiplierRule, 0, len(rules))
	for _, rule := range rules {
		cloned = append(cloned, cloneUserUsageLimitMultiplierRule(rule))
	}
	return cloned
}

func parseUserUsageLimitMultiplierRules(jsonStr string) ([]UserUsageLimitMultiplierRule, error) {
	trimmed := strings.TrimSpace(jsonStr)
	if trimmed == "" {
		return []UserUsageLimitMultiplierRule{}, nil
	}

	decoder := json.NewDecoder(bytes.NewReader([]byte(trimmed)))
	decoder.DisallowUnknownFields()

	var rawRules []UserUsageLimitMultiplierRule
	if err := decoder.Decode(&rawRules); err != nil {
		return nil, err
	}

	normalizedRules := make([]UserUsageLimitMultiplierRule, 0, len(rawRules))
	for idx, rawRule := range rawRules {
		rule := cloneUserUsageLimitMultiplierRule(rawRule)
		rule.Scope = strings.TrimSpace(strings.ToLower(rule.Scope))

		switch rule.Scope {
		case UsageLimitTargetAll:
			rule.GroupNames = nil
			rule.UserIDs = nil
		case UsageLimitTargetGroups:
			rule.GroupNames = normalizeStringList(rule.GroupNames)
			rule.UserIDs = nil
			if len(rule.GroupNames) == 0 {
				return nil, fmt.Errorf("rule %d must specify at least one group", idx+1)
			}
		case UsageLimitTargetUsers:
			rule.GroupNames = nil
			rule.UserIDs = normalizeIntList(rule.UserIDs)
			if len(rule.UserIDs) == 0 {
				return nil, fmt.Errorf("rule %d must specify at least one user id", idx+1)
			}
		default:
			return nil, fmt.Errorf("rule %d scope must be one of all, groups, users", idx+1)
		}

		rule.Metrics = normalizeStringList(rule.Metrics)
		if len(rule.Metrics) == 0 {
			return nil, fmt.Errorf("rule %d must specify at least one metric", idx+1)
		}
		for _, metric := range rule.Metrics {
			if _, ok := validUsageLimitMetrics[metric]; !ok {
				return nil, fmt.Errorf("rule %d metric %s is invalid", idx+1, metric)
			}
		}

		if math.IsNaN(rule.Multiplier) || math.IsInf(rule.Multiplier, 0) || rule.Multiplier <= 0 {
			return nil, fmt.Errorf("rule %d multiplier must be greater than 0", idx+1)
		}

		normalizedRules = append(normalizedRules, rule)
	}

	return normalizedRules, nil
}

func UserUsageLimitMultiplierRules2JSONString() string {
	userUsageLimitMultiplierRulesMutex.RLock()
	defer userUsageLimitMultiplierRulesMutex.RUnlock()

	jsonBytes, err := json.Marshal(userUsageLimitMultiplierRules)
	if err != nil {
		common.SysLog("error marshalling user usage limit multiplier rules: " + err.Error())
		return "[]"
	}
	return string(jsonBytes)
}

func UpdateUserUsageLimitMultiplierRulesByJSONString(jsonStr string) error {
	rules, err := parseUserUsageLimitMultiplierRules(jsonStr)
	if err != nil {
		return err
	}

	userUsageLimitMultiplierRulesMutex.Lock()
	defer userUsageLimitMultiplierRulesMutex.Unlock()

	userUsageLimitMultiplierRules = rules
	return nil
}

func CheckUserUsageLimitMultiplierRules(jsonStr string) error {
	_, err := parseUserUsageLimitMultiplierRules(jsonStr)
	return err
}

func GetUserUsageLimitMultiplierRulesCopy() []UserUsageLimitMultiplierRule {
	userUsageLimitMultiplierRulesMutex.RLock()
	defer userUsageLimitMultiplierRulesMutex.RUnlock()

	return cloneUserUsageLimitMultiplierRules(userUsageLimitMultiplierRules)
}

func ruleContainsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func ruleContainsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func ruleMatchesUser(rule UserUsageLimitMultiplierRule, userID int, group string) bool {
	switch rule.Scope {
	case UsageLimitTargetAll:
		return true
	case UsageLimitTargetGroups:
		return ruleContainsString(rule.GroupNames, group)
	case UsageLimitTargetUsers:
		return ruleContainsInt(rule.UserIDs, userID)
	default:
		return false
	}
}

func ResolveUserUsageLimitMetricMultipliers(userID int, group string) map[string]float64 {
	rules := GetUserUsageLimitMultiplierRulesCopy()
	if len(rules) == 0 {
		return map[string]float64{}
	}

	resolved := make(map[string]float64)
	for _, scope := range []string{UsageLimitTargetAll, UsageLimitTargetGroups, UsageLimitTargetUsers} {
		for _, rule := range rules {
			if rule.Scope != scope || !ruleMatchesUser(rule, userID, group) {
				continue
			}
			for _, metric := range rule.Metrics {
				resolved[metric] = rule.Multiplier
			}
		}
	}
	return resolved
}
