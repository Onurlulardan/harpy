package application

import "testing"

func TestAlertEventLookupValidatesAndForwardsOwnedIdentity(t *testing.T) {
	fixture := &detailLookupFixture{}
	service := &OperationsService{alerts: fixture}
	result, err := service.AlertEventLookup(t.Context(), detailEnvironmentID, detailEventID)
	if err != nil || fixture.operation != "alert_event" || result.ID != detailEventID {
		t.Fatalf("AlertEventLookup() = %#v, %v", result, err)
	}
	if _, err = service.AlertEventLookup(t.Context(), detailEnvironmentID, "other"); err == nil {
		t.Fatal("AlertEventLookup accepted a non-canonical event identity")
	}
}
