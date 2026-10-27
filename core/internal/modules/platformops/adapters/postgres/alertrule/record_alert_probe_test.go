package alertrule

import "testing"

func TestRecordForwardsAlertProbe(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewAlertRuleOperations(backend).RecordAlertTest(
		t.Context(), "environment", "rule", "actor", "message")
	if err != nil || backend.operation != "record" || result.ID != "event" {
		t.Fatalf("RecordAlertTest() = %#v, %v", result, err)
	}
}
