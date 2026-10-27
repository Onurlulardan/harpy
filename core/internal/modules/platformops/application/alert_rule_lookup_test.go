package application

import "testing"

func TestAlertRuleLookupValidatesAndForwardsOwnedIdentity(t *testing.T) {
	fixture := &detailLookupFixture{}
	service := &OperationsService{alerts: fixture}
	result, err := service.AlertRuleLookup(t.Context(), detailEnvironmentID, detailRuleID)
	if err != nil || fixture.operation != "alert_rule" || result.ID != detailRuleID {
		t.Fatalf("AlertRuleLookup() = %#v, %v", result, err)
	}
	if _, err = service.AlertRuleLookup(t.Context(), detailEnvironmentID, "other"); err == nil {
		t.Fatal("AlertRuleLookup accepted a non-canonical rule identity")
	}
}
