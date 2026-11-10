package goose

import (
	"os"
	"strings"
	"testing"
)

func TestRecoveryTargetReadsTheSucceededActivation(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00801_tenancy_get_recovery_target_persist.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(body), "-- +goose Down")
	if !found {
		t.Fatal("down section is missing")
	}
	for _, fragment := range []string{
		"tenancy_activation_facts",
		"tenancy_onboarding_operations",
		"operation.state = 'succeeded'",
		"tenant-set.v1|",
		"SECURITY DEFINER",
		"tenancy recovery target is missing",
		"p_purpose IS DISTINCT FROM 'backup'",
		"p_purpose IS DISTINCT FROM 'restore'",
	} {
		if !strings.Contains(up, fragment) {
			t.Fatalf("missing %s", fragment)
		}
	}
	if strings.Contains(up, "00000000-0000-0000-0000-000000000000") {
		t.Fatal("persisted recovery target still returns the empty stub")
	}
}

func TestRecoveryTargetStubStaysImmutable(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00777_tenancy_get_recovery_target_v1.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(body), "00000000-0000-0000-0000-000000000000") != 2 {
		t.Fatal("applied recovery target migration changed")
	}
}
