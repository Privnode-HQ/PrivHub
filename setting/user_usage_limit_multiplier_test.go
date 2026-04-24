package setting

import "testing"

func TestUserUsageLimitMultiplierRulesRoundTripAndResolve(t *testing.T) {
	originalRules := UserUsageLimitMultiplierRules2JSONString()
	t.Cleanup(func() {
		if err := UpdateUserUsageLimitMultiplierRulesByJSONString(originalRules); err != nil {
			t.Errorf("restore multiplier rules: %v", err)
		}
	})

	input := `[
		{"scope":"all","metrics":["daily","weekly"],"multiplier":1.05},
		{"scope":"groups","group_names":["vip","vip"],"metrics":["daily"],"multiplier":1.2},
		{"scope":"groups","group_names":["vip"],"metrics":["daily"],"multiplier":1.1},
		{"scope":"users","user_ids":[7,7],"metrics":["weekly"],"multiplier":1.4}
	]`

	if err := UpdateUserUsageLimitMultiplierRulesByJSONString(input); err != nil {
		t.Fatalf("UpdateUserUsageLimitMultiplierRulesByJSONString returned error: %v", err)
	}

	serialized := UserUsageLimitMultiplierRules2JSONString()
	if err := CheckUserUsageLimitMultiplierRules(serialized); err != nil {
		t.Fatalf("CheckUserUsageLimitMultiplierRules returned error: %v", err)
	}

	rules := GetUserUsageLimitMultiplierRulesCopy()
	if len(rules) != 4 {
		t.Fatalf("expected 4 rules, got %d", len(rules))
	}
	if len(rules[1].GroupNames) != 1 || rules[1].GroupNames[0] != "vip" {
		t.Fatalf("expected duplicate group names to be normalized, got %#v", rules[1].GroupNames)
	}
	if len(rules[3].UserIDs) != 1 || rules[3].UserIDs[0] != 7 {
		t.Fatalf("expected duplicate user ids to be normalized, got %#v", rules[3].UserIDs)
	}

	multipliers := ResolveUserUsageLimitMetricMultipliers(7, "vip")
	if multipliers["daily"] != 1.1 {
		t.Fatalf("expected group daily multiplier to be overridden by later group rule, got %v", multipliers["daily"])
	}
	if multipliers["weekly"] != 1.4 {
		t.Fatalf("expected user weekly multiplier to override all-scope rule, got %v", multipliers["weekly"])
	}

	defaultMultipliers := ResolveUserUsageLimitMetricMultipliers(9, "default")
	if defaultMultipliers["daily"] != 1.05 {
		t.Fatalf("expected all-scope daily multiplier to apply, got %v", defaultMultipliers["daily"])
	}
	if defaultMultipliers["weekly"] != 1.05 {
		t.Fatalf("expected all-scope weekly multiplier to apply, got %v", defaultMultipliers["weekly"])
	}
}

func TestCheckUserUsageLimitMultiplierRulesRejectsInvalidRules(t *testing.T) {
	if err := CheckUserUsageLimitMultiplierRules(`[{"scope":"groups","metrics":["daily"],"multiplier":1.1}]`); err == nil {
		t.Fatalf("expected missing group_names to be rejected")
	}
	if err := CheckUserUsageLimitMultiplierRules(`[{"scope":"users","user_ids":[1],"metrics":["unknown"],"multiplier":1.1}]`); err == nil {
		t.Fatalf("expected unknown metric to be rejected")
	}
	if err := CheckUserUsageLimitMultiplierRules(`[{"scope":"all","metrics":["daily"],"multiplier":0}]`); err == nil {
		t.Fatalf("expected non-positive multiplier to be rejected")
	}
}
