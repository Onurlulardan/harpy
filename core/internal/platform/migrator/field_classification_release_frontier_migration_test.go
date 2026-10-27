package migrator

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMFieldClassificationReleaseFrontierIsPairedAndFailClosed(
	t *testing.T,
) {
	const name = "20260730113300_assert_iam_field_classification_release_frontier_v1"
	fresh := readIAMNamedPlatformMigration(t, "fresh", name)
	adopted := readIAMNamedPlatformMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("IAM field-classification frontier origin pair differs")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect IAM field-classification frontier migration: %v", err)
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("IAM field-classification frontier migration exceeds 200 lines")
	}
	source := string(fresh)
	for _, marker := range []string{
		"authority repair owns database generation 75",
		"Advance the platform frontier to 76",
		"server_version_num",
		"BETWEEN 180000 AND 189999",
		"role_record.rolname = 'harpy_runtime'",
		"role_record.rolname = 'harpy_v2_runtime'",
		"NOT role_record.rolsuper",
		"NOT role_record.rolinherit",
		"NOT role_record.rolcanlogin",
		"AND role_record.rolcanlogin",
		"iam_field_classification_release_frontier_assertion",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM field-classification frontier lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"-- +goose Down", "IF NOT EXISTS", "CREATE OR REPLACE",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "pg_temp",
		"platform.platform_migration_history",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("IAM field-classification frontier contains %q", forbidden)
		}
	}
}
