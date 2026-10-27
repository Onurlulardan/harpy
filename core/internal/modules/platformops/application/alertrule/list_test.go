package alertrule

import "testing"

func TestListForwardsAlertRuleScope(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewAlertRuleOperations(backend).ListAlertRules(t.Context(), "environment")
	if err != nil || backend.operation != "list" || len(result) != 1 {
		t.Fatalf("ListAlertRules() = %#v, %v", result, err)
	}
}
