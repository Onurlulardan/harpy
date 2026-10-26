package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamProfileCommandMigrationNames = []string{
	"20260716013800_create_iam_profile_actor_assert_v1",
	"20260716013801_create_iam_profile_retirement_evidence",
	"20260716013802_create_iam_profile_schema_apply_v1",
	"20260716013803_create_iam_profile_schema_publish_v1",
	"20260716013804_create_iam_profile_document_backfill_v1",
	"20260716013805_create_iam_profile_schema_retire_v1",
	"20260716013806_grant_iam_profile_commands_v1",
}

func TestIAMProfileCommandMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamProfileCommandMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and v1_adopted migration content differs")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("profile command migration is invalid: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"IF NOT EXISTS", "-- +GOOSE DOWN", "DO $$", "CREATE OR REPLACE",
				"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
				"ON DELETE CASCADE", "ON DELETE SET NULL",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("profile migration contains forbidden token %q", forbidden)
				}
			}
		})
	}
}

func TestIAMProfileCommandsKeepAuthorityCASAndImmutableEvidence(t *testing.T) {
	publish := string(readIAMIdentityMigration(t, "fresh",
		iamProfileCommandMigrationNames[3]))
	backfill := string(readIAMIdentityMigration(t, "fresh",
		iamProfileCommandMigrationNames[4]))
	retire := string(readIAMIdentityMigration(t, "fresh",
		iamProfileCommandMigrationNames[5]))
	for _, marker := range []string{
		"iam_profile_actor_assert_v1", "p_expected_active_schema_id",
		"iam_authorization_operation_receipts", "iam_profile_schema_apply_v1",
	} {
		if !strings.Contains(publish, marker) {
			t.Fatalf("profile publish lacks %q", marker)
		}
	}
	for _, marker := range []string{
		"pg_catalog.sha256(p_profile_document)", "p_expected_source_schema_id",
		"revision = p_expected_revision", "iam_profile_field_revisions",
	} {
		if !strings.Contains(backfill, marker) {
			t.Fatalf("profile backfill lacks %q", marker)
		}
	}
	for _, marker := range []string{
		"iam_profile_retirement_evidence", "p_checked_owner_count",
		"p_expected_field_revision", "iam_profile_schema_apply_v1",
	} {
		if !strings.Contains(retire, marker) {
			t.Fatalf("profile retirement lacks %q", marker)
		}
	}
}
