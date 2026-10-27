package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamLifecycleGuardMigrationNames = []string{
	"20260716012920_guard_iam_revision_step_v1",
	"20260716012921_guard_iam_foundation_identity_v1",
	"20260716012922_guard_iam_auth_identity_v1",
	"20260716012923_guard_iam_token_identity_v1",
	"20260716012924_guard_iam_user_security_versions_v1",
	"20260716012925_guard_iam_terminal_states_v1",
	"20260716012926_guard_iam_monotonic_counters_v1",
}

func TestIAMLifecycleGuardMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamLifecycleGuardMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and V1-adopted lifecycle guard differs")
			}
			assertIAMLifecycleGuardSource(t, fresh)
		})
	}
}

func assertIAMLifecycleGuardSource(t *testing.T, content []byte) {
	t.Helper()
	if bytes.Count(content, []byte("\n"))+1 > 200 {
		t.Fatal("IAM lifecycle guard exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(content); err != nil {
		t.Fatalf("IAM lifecycle guard source is invalid: %v", err)
	}
	upper := strings.ToUpper(string(content))
	for _, forbidden := range []string{
		"IF NOT EXISTS", "-- +GOOSE DOWN", "CREATE OR REPLACE",
		"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
		"PUBLIC.", "\nGRANT ", "DISABLE TRIGGER", "ON DELETE CASCADE",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("IAM lifecycle guard contains forbidden token %q", forbidden)
		}
	}
	triggers := strings.Count(upper, "CREATE TRIGGER ")
	if triggers == 0 || strings.Count(upper, "ENABLE ALWAYS TRIGGER") != triggers {
		t.Fatal("IAM lifecycle triggers are not all enabled ALWAYS")
	}
	functions := strings.Count(upper, "CREATE FUNCTION ")
	if functions > 0 {
		for _, marker := range []string{
			"SECURITY DEFINER", "SET SEARCH_PATH = PG_CATALOG",
			"REVOKE ALL ON FUNCTION", "FROM PUBLIC",
			"-- +GOOSE STATEMENTBEGIN", "-- +GOOSE STATEMENTEND",
		} {
			if !strings.Contains(upper, marker) {
				t.Fatalf("IAM lifecycle function lacks %q", marker)
			}
		}
		if strings.Count(upper, "REVOKE ALL ON FUNCTION") != functions {
			t.Fatal("IAM lifecycle function revoke count differs")
		}
	}
}
