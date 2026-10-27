package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var systemTransferCleanupMigrations = []string{
	"20260718030100_add_iam_import_source_cleanup_v1",
	"20260718030101_create_iam_import_source_cleanup_claim_v1",
	"20260718030102_create_iam_import_source_cleanup_complete_v1",
	"20260718030103_create_iam_import_source_cleanup_retry_v1",
	"20260718030104_grant_iam_import_source_cleanup_v1",
}

func TestSystemTransferCleanupMigrationsStayPairedAndBounded(t *testing.T) {
	for _, name := range systemTransferCleanupMigrations {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			if len(fresh) == 0 {
				t.Fatal("fresh migration is empty")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("IAM import cleanup migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("IAM import cleanup migration is invalid: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"-- +GOOSE DOWN", "CREATE OR REPLACE", "IF NOT EXISTS",
				"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("cleanup migration contains forbidden token %q", forbidden)
				}
			}
		})
	}
}

func TestSystemTransferCleanupFencesClaimDeleteRace(t *testing.T) {
	migration := string(readIAMIdentityMigration(
		t, "fresh", systemTransferCleanupMigrations[0],
	))
	for _, required := range []string{
		"OLD.state='ready' AND OLD.cleanup_state='waiting'",
		"NEW.state='claimed'",
		"NEW.state='deleted'",
		"NEW.absence_probe_sha256 IS NOT NULL",
	} {
		if !strings.Contains(migration, required) {
			t.Fatalf("cleanup race fence is missing %q", required)
		}
	}
}
