package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamCatalogLifecycleV3Migrations = []string{
	"20260730145300_create_iam_catalog_blocker_snapshot_v2",
	"20260730145301_route_iam_catalog_blockers_v2",
	"20260730145302_create_iam_catalog_grant_lifecycle_v2",
	"20260730145303_create_iam_catalog_transition_v3",
	"20260730145304_guard_iam_catalog_version_purge_v2",
	"20260730145305_create_iam_catalog_purge_data_v2",
	"20260730145306_create_iam_catalog_component_purge_v3",
	"20260730145307_route_iam_catalog_component_apply_v3",
}

func TestIAMCatalogLifecycleV3MigrationsArePairedAndFailClosed(t *testing.T) {
	var joined strings.Builder
	for _, name := range iamCatalogLifecycleV3Migrations {
		fresh := readIAMIdentityMigration(t, "fresh", name)
		adopted := readIAMIdentityMigration(t, "v1_adopted", name)
		if !bytes.Equal(fresh, adopted) {
			t.Fatalf("IAM Catalog lifecycle origin pair differs: %s", name)
		}
		if bytes.Count(fresh, []byte("\n"))+1 > 200 {
			t.Fatalf("IAM Catalog lifecycle migration exceeds 200 lines: %s", name)
		}
		if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
			t.Fatalf("inspect IAM Catalog lifecycle migration %s: %v", name, err)
		}
		joined.Write(fresh)
	}
	source := joined.String()
	for _, marker := range []string{
		"iam_catalog_blocker_snapshot_v2(",
		"IAM_CATALOG_EXTERNAL_PERMISSION_REFERENCE",
		"iam_catalog_grants_set_v2(",
		"iam_catalog_authorization_record_v2(",
		"iam_catalog_purge_contexts",
		"iam_catalog_purge_data_v2(",
		"iam_catalog_component_purge_v3(",
		"GRANT EXECUTE ON FUNCTION iam_catalog_component_apply_v1",
		"REVOKE EXECUTE ON FUNCTION iam_catalog_component_apply_v2",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM Catalog lifecycle migrations lack %q", marker)
		}
	}
	for _, forbidden := range []string{
		"CREATE OR REPLACE FUNCTION", "-- +goose Down",
		"DROP TABLE iam_catalog_", "GRANT EXECUTE ON FUNCTION iam_catalog_component_apply_v2",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("IAM Catalog lifecycle migrations contain %q", forbidden)
		}
	}
}
