package migrator

import (
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMLayoutParityReleaseFrontierIsPairedAndFailClosed(t *testing.T) {
	const name = "20260730120200_assert_iam_layout_parity_release_frontier_v1"
	fresh := readIAMNamedPlatformMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh IAM layout parity frontier migration did not load")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect IAM layout parity frontier migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"profile/layout parity repairs own database generation 79",
		"Advance the platform frontier to 80",
		"BETWEEN 180000 AND 189999",
		"role_record.rolname = 'harpy_runtime'",
		"role_record.rolname = 'harpy_v2_runtime'",
		"NOT role_record.rolsuper",
		"NOT role_record.rolinherit",
		"NOT role_record.rolcanlogin",
		"AND role_record.rolcanlogin",
		"iam_layout_parity_release_frontier_assertion",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM layout parity frontier lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"-- +goose Down", "IF NOT EXISTS", "CREATE OR REPLACE",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "pg_temp",
		"platform.platform_migration_history",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("IAM layout parity frontier contains %q", forbidden)
		}
	}
}
