package platformauthority

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestInstallationHeadPinUsesAdditiveCapabilitySuccessor(t *testing.T) {
	const base = "20260729100032_create_installation_head_capability_v1"
	fresh, err := os.ReadFile(
		"migrations/platform/fresh/" + base + ".fresh.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	adopted, err := os.ReadFile(
		"migrations/platform/v1_adopted/" + base + ".v1_adopted.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("installation head capability migration origins differ")
	}
	source := string(fresh)
	for _, marker := range []string{
		"CREATE FUNCTION platform.platformauthority_installation_head_pin_v3(",
		"platform.tenancy_activation_step_context_assert_v1(",
		"'platformauthority.pin_protected_authority_head'",
		"'environment_initialize'",
		"'saas_customer_tenant_provision'",
		"platform.tenancy_activation_step_consume_v1(",
		"CHECK (scope_kind IN ('environment', 'tenant'))",
		"TO harpy_runtime",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("installation head capability lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"current_setting('application_name'",
		"DROP FUNCTION platform.platformauthority_installation_head_pin_v2",
		"CREATE OR REPLACE FUNCTION " +
			"platform.platformauthority_installation_head_pin_v2",
		"REVOKE EXECUTE ON FUNCTION " +
			"platform.platformauthority_installation_head_pin_v2",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("installation head expand contains %q", forbidden)
		}
	}
	route, err := os.ReadFile(
		"adapters/postgres/sql/platform/installation_pin_head.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(
		string(route), "platformauthority_installation_head_pin_v3(",
	) || strings.Contains(
		string(route), "platformauthority_installation_head_pin_v2(",
	) {
		t.Fatal("installation head adapter is not routed exclusively to v3")
	}
}
