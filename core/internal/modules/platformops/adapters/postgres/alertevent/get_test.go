package alertevent

import "testing"

func TestGetForwardsAlertEventLookup(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewAlertEventOperations(backend).GetAlertEvent(
		t.Context(), "environment", "event")
	if err != nil || backend.operation != "get" || result.ID != "event" {
		t.Fatalf("GetAlertEvent() = %#v, %v", result, err)
	}
}
