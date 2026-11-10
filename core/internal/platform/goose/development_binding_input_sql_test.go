package goose

import (
	"os"
	"strings"
	"testing"
)

func TestDevelopmentBindingInputReadsTheActivationReceipt(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00800_platformauthority_development_binding_input_persist.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(body), "-- +goose Down")
	if !found {
		t.Fatal("down section is missing")
	}
	for _, fragment := range []string{
		"tenancy_onboarding_operations",
		"tenancy_onboarding_receipts",
		"activation_completed",
		"recovery_trust_bundle_sha256",
		"SECURITY DEFINER",
		"development authority binding input is missing",
	} {
		if !strings.Contains(up, fragment) {
			t.Fatalf("missing %s", fragment)
		}
	}
	if strings.Contains(up, "''::text") {
		t.Fatal("persisted reader still returns the empty stub")
	}
}

func TestDevelopmentBindingInputStubStaysImmutable(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00702_platformauthority_development_binding_input_v1.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(body), "''::text") != 4 {
		t.Fatal("applied development binding input migration changed")
	}
}
