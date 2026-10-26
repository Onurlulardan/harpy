package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamCatalogVersionUpgradeMigrations = []string{
	"20260730142011_create_iam_catalog_version_evidence_v2",
	"20260730142012_create_iam_catalog_version_runtime_v2",
	"20260730142013_backfill_iam_catalog_version_evidence_v2",
	"20260730142014_guard_iam_catalog_version_evidence_v2",
	"20260730142015_create_iam_catalog_version_materializer_v2",
	"20260730142016_create_iam_catalog_version_read_v2",
	"20260730142017_create_iam_catalog_version_writer_v2",
	"20260730142018_create_iam_catalog_operational_helpers_v2",
	"20260730142019_create_iam_catalog_operational_readback_v2",
	"20260730142020_create_iam_catalog_operational_activate_v2",
	"20260730142021_create_iam_catalog_component_apply_v2",
	"20260730142022_grant_iam_catalog_version_runtime_v2",
}

func TestIAMCatalogVersionUpgradeMigrationsArePairedAndForwardOnly(t *testing.T) {
	var joined strings.Builder
	for _, name := range iamCatalogVersionUpgradeMigrations {
		fresh := readIAMIdentityMigration(t, "fresh", name)
		adopted := readIAMIdentityMigration(t, "v1_adopted", name)
		if !bytes.Equal(fresh, adopted) {
			t.Fatalf("IAM catalog version origin pair differs: %s", name)
		}
		if bytes.Count(fresh, []byte("\n"))+1 > 200 {
			t.Fatalf("IAM catalog version migration exceeds 200 lines: %s", name)
		}
		if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
			t.Fatalf("inspect IAM catalog version migration %s: %v", name, err)
		}
		joined.Write(fresh)
	}
	source := joined.String()
	for _, marker := range []string{
		"CREATE TABLE iam_catalog_version_evidence",
		"CREATE TABLE iam_catalog_component_heads",
		"CREATE TABLE iam_catalog_component_candidates",
		"CREATE FUNCTION iam_catalog_project_v2(",
		"CREATE FUNCTION iam_catalog_component_apply_v2(",
		"iam_catalog_operational_activate_v2(",
		"ENABLE ALWAYS TRIGGER iam_catalog_version_evidence_immutable",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM catalog version migrations lack %q", marker)
		}
	}
	for _, forbidden := range []string{
		"CREATE OR REPLACE", "-- +goose Down",
		"DROP TABLE iam_catalog_", "DROP FUNCTION iam_catalog_",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("IAM catalog version migrations contain %q", forbidden)
		}
	}
}
