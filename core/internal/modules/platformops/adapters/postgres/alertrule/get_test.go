package alertrule

import "testing"

func TestGetForwardsAlertRuleLookup(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewAlertRuleOperations(backend).GetAlertRule(
		t.Context(), "environment", "rule")
	if err != nil || backend.operation != "get" || result.ID != "rule" {
		t.Fatalf("GetAlertRule() = %#v, %v", result, err)
	}
}
