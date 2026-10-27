package installationactivation

import (
	"strings"
	"testing"
)

func TestAssignmentUsesCapabilitySuccessor(t *testing.T) {
	if !strings.Contains(assignSQL,
		"platform.billing_installation_assign_v3(") {
		t.Fatal("Billing installation assignment does not call v3")
	}
	for _, legacy := range []string{
		"platform.billing_installation_assign_v1(",
		"platform.billing_installation_assign_v2(",
	} {
		if strings.Contains(assignSQL, legacy) {
			t.Fatalf("Billing installation assignment calls %s", legacy)
		}
	}
}

func TestUsageOpeningUsesCapabilitySuccessor(t *testing.T) {
	if !strings.Contains(openSQL,
		"platform.billing_installation_open_usage_v2(") {
		t.Fatal("Billing installation usage opening does not call v2")
	}
	if strings.Contains(openSQL,
		"platform.billing_installation_open_usage_v1(") {
		t.Fatal("Billing installation usage opening calls legacy v1")
	}
}
