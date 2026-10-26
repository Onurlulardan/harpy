package migrator

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMIdentityGrantEvidenceReleaseFrontierIsPairedAndFailClosed(
	t *testing.T,
) {
	const name = "20260730113500_assert_iam_identity_grant_evidence_release_frontier_v1"
	fresh := readIAMNamedPlatformMigration(t, "fresh", name)
	adopted := readIAMNamedPlatformMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("IAM identity grant-evidence frontier origin pair differs")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect IAM identity grant-evidence frontier migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"grant-evidence repair owns database generation 77",
		"Advance the platform frontier to 78",
		"BETWEEN 180000 AND 189999",
		"role_record.rolname = 'harpy_runtime'",
		"role_record.rolname = 'harpy_v2_runtime'",
		"NOT role_record.rolsuper",
		"NOT role_record.rolinherit",
		"NOT role_record.rolcanlogin",
		"AND role_record.rolcanlogin",
		"iam_identity_grant_evidence_release_frontier_assertion",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM identity grant-evidence frontier lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"-- +goose Down", "IF NOT EXISTS", "CREATE OR REPLACE",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "pg_temp",
		"platform.platform_migration_history",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("IAM identity grant-evidence frontier contains %q", forbidden)
		}
	}
}
