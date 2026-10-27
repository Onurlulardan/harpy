package tenancy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTenantSettingsMigrationPairsSecureRuntimeRoutines(t *testing.T) {
	const base = "20260801020000_create_tenancy_settings_runtime_v1"
	for _, origin := range []string{"fresh"} {
		path := filepath.Join("migrations", "platform", origin,
			base+"."+origin+".sql")
		document, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		source := strings.Join(strings.Fields(string(document)), " ")
		for _, marker := range []string{
			"ADD COLUMN settings_revision bigint NOT NULL DEFAULT 1",
			"platform.tenancy_settings_operation_receipts",
			"CREATE FUNCTION platform.tenancy_settings_read_v1",
			"CREATE FUNCTION platform.tenancy_settings_update_v1",
			"SECURITY DEFINER SET search_path TO pg_catalog",
			"FOR UPDATE",
			"idempotency_conflict",
			"CREATE FUNCTION platform.tenancy_v1_settings_fence_v1",
			"INSTEAD OF UPDATE ON public.tenants",
			"GRANT EXECUTE ON FUNCTION platform.tenancy_settings_read_v1(uuid)",
			"GRANT EXECUTE ON FUNCTION platform.tenancy_settings_update_v1(uuid,text,bigint,text,bytea)",
		} {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s is missing %q", path, marker)
			}
		}
		if strings.Contains(source, "GRANT SELECT ON TABLE platform.tenants") ||
			strings.Contains(source, "GRANT INSERT ON TABLE platform.tenancy_settings_operation_receipts") {
			t.Fatalf("%s grants direct runtime table access", path)
		}
	}
}
