package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamProfileRevisionMigrationNames = []string{
	"20260716013200_expand_iam_profile_revision_snapshots",
	"20260716013201_create_iam_profile_validation_canonical_v1",
	"20260716013202_create_iam_profile_semantic_sha256_v1",
	"20260716013203_backfill_iam_profile_revision_snapshots",
	"20260716013204_seal_iam_profile_revision_snapshots",
	"20260716013205_create_iam_profile_definition_sha256_v1",
	"20260716013206_backfill_iam_profile_definition_hashes",
	"20260716013210_guard_iam_profile_destructive_mutation_v1",
	"20260716013211_guard_iam_profile_schema_lifecycle_v1",
	"20260716013212_guard_iam_profile_field_lifecycle_v1",
	"20260716013213_guard_iam_profile_snapshot_insert_v1",
	"20260716013215_create_iam_profile_historical_read_v1",
	"20260716013216_create_iam_profile_active_catalog_read_v1",
	"20260716013217_create_iam_profile_document_read_v1",
	"20260716013220_grant_iam_profile_read_routines_v1",
}

func TestIAMProfileRevisionMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamProfileRevisionMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			assertIAMProfileRevisionMigrationSource(t, fresh)
		})
	}
}

func assertIAMProfileRevisionMigrationSource(t *testing.T, content []byte) {
	t.Helper()
	if bytes.Count(content, []byte("\n"))+1 > 200 {
		t.Fatal("IAM profile revision migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(content); err != nil {
		t.Fatalf("IAM profile revision migration source is invalid: %v", err)
	}
	upper := strings.ToUpper(string(content))
	for _, forbidden := range []string{
		"IF NOT EXISTS", "-- +GOOSE DOWN", "CREATE OR REPLACE",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
		"PUBLIC.", "ON DELETE CASCADE", "HARPY_ADOPTER",
		"GRANT SELECT", "GRANT INSERT", "GRANT UPDATE", "GRANT DELETE",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("IAM profile revision migration contains %q", forbidden)
		}
	}
	functions := strings.Count(upper, "CREATE FUNCTION ")
	if functions > 0 && (strings.Count(upper, "SECURITY DEFINER") != functions ||
		strings.Count(upper, "REVOKE ALL ON FUNCTION") != functions) {
		t.Fatal("IAM profile routine is not closed SECURITY DEFINER")
	}
	triggers := strings.Count(upper, "CREATE TRIGGER ")
	if strings.Count(upper, "ENABLE ALWAYS TRIGGER") != triggers {
		t.Fatal("IAM profile triggers are not all enabled ALWAYS")
	}
}
