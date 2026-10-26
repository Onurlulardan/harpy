package alertrule

import "testing"

func TestDeleteForwardsAlertRuleIdentity(t *testing.T) {
	backend := &backendFixture{}
	err := NewAlertRuleOperations(backend).DeleteAlertRule(t.Context(), "environment", "rule")
	if err != nil || backend.operation != "delete" {
		t.Fatalf("DeleteAlertRule() error = %v", err)
	}
}
