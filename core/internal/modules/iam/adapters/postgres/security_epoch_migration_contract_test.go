package postgres

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var securityEpochMigrationNames = map[string][]string{
	"platform": {
		"20260722022000_create_iam_security_epochs",
		"20260722022001_guard_iam_security_epochs",
		"20260722022002_create_iam_security_epoch_initialize_v1",
		"20260725123500_grant_iam_installation_security_epoch_runtime",
	},
	"tenant": {
		"20260722022100_create_iam_security_epoch_current_v1",
		"20260722022101_bind_iam_credentials_security_epoch",
		"20260722022102_guard_iam_credential_security_epoch",
		"20260722022103_create_iam_session_issue_v2",
		"20260722022104_create_iam_refresh_credential_v3",
		"20260722022105_create_iam_refresh_rotate_v4",
		"20260722022106_create_iam_current_session_read_v2",
		"20260722022107_grant_iam_security_epoch_runtime_v1",
		"20260722022108_create_iam_security_epoch_assert_v1",
	},
}

func TestIAMSecurityEpochMigrationsArePairedAndClosed(t *testing.T) {
	for scope, names := range securityEpochMigrationNames {
		for _, name := range names {
			t.Run(scope+"/"+name, func(t *testing.T) {
				fresh := readSecurityEpochMigration(t, scope, "fresh", name)
				if len(fresh) == 0 {
					t.Fatal("fresh migration is empty")
				}
				if bytes.Count(fresh, []byte("\n"))+1 > 200 {
					t.Fatal("security epoch migration exceeds 200 lines")
				}
				if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
					t.Fatalf("invalid security epoch migration: %v", err)
				}
				upper := strings.ToUpper(string(fresh))
				for _, forbidden := range []string{
					"IF NOT EXISTS", "CREATE OR REPLACE", "-- +GOOSE DOWN",
					"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
				} {
					if strings.Contains(upper, forbidden) {
						t.Fatalf("migration contains forbidden token %q", forbidden)
					}
				}
			})
		}
	}
}

func TestIAMSecurityEpochMigrationMarkersAreExact(t *testing.T) {
	joined := make([]byte, 0, 32_768)
	for scope, names := range securityEpochMigrationNames {
		for _, name := range names {
			joined = append(joined,
				readSecurityEpochMigration(t, scope, "fresh", name)...)
		}
	}
	source := string(joined)
	for _, marker := range []string{
		"CREATE TABLE platform.iam_security_epochs",
		"pg_catalog.octet_length(security_epoch) = 32",
		"ENABLE ALWAYS TRIGGER iam_security_epochs_guard",
		"ADD COLUMN security_epoch bytea",
		"ENABLE ALWAYS TRIGGER iam_sessions_security_epoch_guard",
		"iam_session_issue_v2(", "iam_refresh_credential_v3(",
		"iam_refresh_rotate_v4(", "iam_current_session_read_v2(",
		"iam_security_epoch_assert_v1(",
		"REVOKE EXECUTE ON FUNCTION iam_current_session_read_v1",
		"GRANT EXECUTE ON FUNCTION iam_current_session_read_v2",
		"GRANT EXECUTE ON FUNCTION platform.iam_security_epoch_initialize_v1",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("security epoch migrations lack marker %q", marker)
		}
	}
}

func readSecurityEpochMigration(
	t *testing.T, scope, origin, name string,
) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "migrations", scope, origin,
		name+"."+origin+".sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}
