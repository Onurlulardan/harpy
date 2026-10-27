package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var installationCredentialMigrations = []string{
	"20260722091000_create_iam_installation_password_credentials",
	"20260722091001_create_iam_installation_credential_initialize_v1",
	"20260722091002_create_iam_installation_credential_supersede_v1",
	"20260722091003_create_iam_initial_password_continuations",
	"20260722091004_create_iam_initial_password_credential_v1",
	"20260722091005_create_iam_initial_password_start_v1",
	"20260722091006_create_iam_initial_password_continuation_v1",
	"20260722091007_create_iam_initial_password_reject_v1",
	"20260722091008_create_iam_initial_password_complete_v1",
}

var installationCredentialRoutineMigrations = []string{
	installationCredentialMigrations[1],
	installationCredentialMigrations[2],
	installationCredentialMigrations[4],
	installationCredentialMigrations[5],
	installationCredentialMigrations[6],
	installationCredentialMigrations[7],
	installationCredentialMigrations[8],
}

func TestInstallationCredentialMigrationsStayPairedAndHardened(t *testing.T) {
	for _, name := range installationCredentialMigrations {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("installation credential migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("migration hardening failed: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"CREATE TABLE IF NOT EXISTS", "CREATE OR REPLACE", "-- +GOOSE DOWN",
				"PLAINTEXT", "RAW_SECRET", "RAISE NOTICE", "ON DELETE CASCADE",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("migration contains forbidden token %q", forbidden)
				}
			}
		})
	}
}

func TestInstallationCredentialRoutinesKeepPlaintextOutOfStorage(t *testing.T) {
	for _, name := range installationCredentialRoutineMigrations {
		source := string(readIAMIdentityMigration(t, "fresh", name))
		for _, forbidden := range []string{"p_password", "p_plaintext", "secret text"} {
			if strings.Contains(strings.ToLower(source), forbidden) {
				t.Fatalf("%s exposes %q", name, forbidden)
			}
		}
		for _, required := range []string{"SECURITY DEFINER", "SESSION_USER",
			"CURRENT_USER", "GRANT EXECUTE ON FUNCTION"} {
			if !strings.Contains(source, required) {
				t.Fatalf("%s misses %q", name, required)
			}
		}
	}
}
