package async

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestInstallationActivationCapabilityMigrationIsPairedAndAdditive(t *testing.T) {
	const name = "20260729100029_create_installation_activation_capability_entrypoints_v1"
	fresh := readInstallationActivationMigration(t, name)
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("activation capability migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("invalid activation capability migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"platform_async_initial_admin_setup_enqueue_v2",
		"platform_async_installation_activation_event_v2",
		"platform_async_installation_activation_receipts",
		"platform_async_installation_activation_evidence_assert_v1",
		"tenancy_activation_step_context_assert_v1",
		"'activation.complete'",
		"'environment_initialize'",
		"'saas_customer_tenant_provision'",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("activation capability migration lacks marker %q", marker)
		}
	}
	if strings.Contains(source,
		"GRANT EXECUTE ON FUNCTION platform."+
			"platform_async_installation_activation_evidence_assert_v1") {
		t.Fatal("internal activation evidence assertion is runtime executable")
	}
	for _, legacyACL := range []string{
		"REVOKE EXECUTE ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v1",
		"REVOKE EXECUTE ON FUNCTION platform.platform_async_installation_activation_event_v1",
	} {
		if strings.Contains(source, legacyACL) {
			t.Fatalf("Gen70 migration contracts legacy routine: %s", legacyACL)
		}
	}
}

func readInstallationActivationMigration(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(
		"migrations", "fresh", name+".fresh.sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	return body
}
