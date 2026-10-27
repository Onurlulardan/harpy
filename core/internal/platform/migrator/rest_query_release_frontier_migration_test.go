package migrator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const iamRestQueryReleaseFrontierMigration = "20260730113100_assert_iam_rest_query_search_release_frontier_v2"

func TestIAMRestQuerySearchReleaseFrontierRepairIsPairedAndFailClosed(
	t *testing.T,
) {
	fresh := readIAMNamedPlatformMigration(
		t, "fresh", iamRestQueryReleaseFrontierMigration,
	)
	if len(fresh) == 0 {
		t.Fatal("fresh IAM search frontier migration did not load")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("invalid IAM search frontier migration: %v", err)
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("IAM search frontier migration exceeds 200 lines")
	}

	source := string(fresh)
	for _, marker := range []string{
		"immutable tenant search release already owns database generation 73",
		"advances the platform release frontier to 74",
		"SELECT 1 / CASE WHEN EXISTS (",
		"routine.proname = 'normalize'",
		"routine.proname = 'casefold'",
		"collation_record.collname = 'pg_unicode_fast'",
		"role_record.rolname = 'harpy_runtime'",
		"NOT role_record.rolsuper",
		"NOT role_record.rolcanlogin",
		"THEN 1 ELSE 0 END",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM search frontier assertion lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"-- +goose Down", "IF NOT EXISTS", "CREATE OR REPLACE",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "pg_temp", "SELECT 1;",
		"platform.platform_migration_history",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("IAM search frontier assertion contains %q", forbidden)
		}
	}
}

func readIAMNamedPlatformMigration(t *testing.T, origin, name string) []byte {
	t.Helper()
	if origin != "fresh" {
		t.Fatalf("fresh migration %s did not load", name)
	}
	path := filepath.Join("migrations", "fresh", name+".fresh.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
