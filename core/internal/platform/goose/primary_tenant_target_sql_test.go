package goose

import (
	"os"
	"strings"
	"testing"
)

func TestPrimaryTenantTargetReadsTheSucceededActivation(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00802_tenancy_get_primary_tenant_target_persist.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(body), "-- +goose Down")
	if !found {
		t.Fatal("down section is missing")
	}
	for _, fragment := range []string{
		"tenancy_deployment_policy",
		"tenancy_activation_facts",
		"tenancy_onboarding_operations",
		"onpremise_primary",
		"saas_control",
		"operation.state = 'succeeded'",
		"SECURITY DEFINER",
		"(SELECT count(*) FROM candidate) = 1",
	} {
		if !strings.Contains(up, fragment) {
			t.Fatalf("missing %s", fragment)
		}
	}
	if strings.Contains(up, "WHERE false") {
		t.Fatal("persisted primary target still returns the empty stub")
	}
}

func TestPrimaryTenantTargetStubStaysImmutable(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00097_tenancy_get_primary_tenant_target_v1.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(body), "NULL::uuid") != 2 ||
		!strings.Contains(string(body), "WHERE false") {
		t.Fatal("applied primary tenant target migration changed")
	}
}
