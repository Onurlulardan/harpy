package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamAutomationMigrationNames = []string{
	"20260716013400_create_iam_service_principals",
	"20260716013410_create_iam_capability_grants",
	"20260716013420_guard_iam_capability_grants_v1",
	"20260716013421_guard_iam_capability_grant_terminal_v1",
	"20260716013422_create_iam_capability_grant_read_v1",
	"20260716013423_grant_iam_capability_grant_read_v1",
	"20260716013600_create_iam_service_principal_actor_assert_v1",
	"20260716013601_create_iam_service_principal_create_v1",
	"20260716013602_create_iam_service_principal_state_v1",
	"20260716013603_create_iam_capability_grant_write_v1",
	"20260716013604_create_iam_capability_grant_revoke_v1",
	"20260716013605_grant_iam_service_principal_commands_v1",
	"20260716015600_expand_iam_rbac_change_audit_automation_v1",
}

func TestIAMAutomationMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamAutomationMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and v1_adopted migration content differs")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("automation migration source is invalid: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"IF NOT EXISTS", "-- +GOOSE DOWN", "DO $$", "CREATE OR REPLACE",
				"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
				"ON DELETE CASCADE", "ON DELETE SET NULL",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("automation migration contains forbidden token %q", forbidden)
				}
			}
			if strings.Contains(upper, "\nGRANT ") &&
				name != "20260716013423_grant_iam_capability_grant_read_v1" &&
				name != "20260716013605_grant_iam_service_principal_commands_v1" {
				t.Fatal("automation runtime grant appeared before the closed read routine")
			}
		})
	}
}

func TestIAMAutomationStorageUsesImmutablePinnedRevisions(t *testing.T) {
	principal := string(readIAMIdentityMigration(
		t, "fresh", iamAutomationMigrationNames[0],
	))
	grant := string(readIAMIdentityMigration(
		t, "fresh", iamAutomationMigrationNames[1],
	))
	guard := string(readIAMIdentityMigration(
		t, "fresh", iamAutomationMigrationNames[2],
	))
	terminal := string(readIAMIdentityMigration(
		t, "fresh", iamAutomationMigrationNames[3],
	))
	for _, marker := range []string{
		"kind = 'workflow'", "authorization_version > 0",
		"active>disabled", "disabled>active", "iam_service_principals_removal_guard",
	} {
		if !strings.Contains(principal, marker) {
			t.Fatalf("service principal migration lacks %q", marker)
		}
	}
	for _, marker := range []string{
		"iam_capability_grant_revisions", "previous_revision = grant_revision - 1",
		"scope_kind IN ('tenant', 'resource')", "DEFERRABLE INITIALLY DEFERRED",
		"pg_catalog.octet_length(manifest_sha256) = 32",
	} {
		if !strings.Contains(grant, marker) {
			t.Fatalf("capability migration lacks %q", marker)
		}
	}
	for _, marker := range []string{
		"CREATE CONSTRAINT TRIGGER iam_capability_grants_head_guard",
		"iam_capability_grant_revisions_mutation_guard",
	} {
		if !strings.Contains(guard, marker) {
			t.Fatalf("capability guard migration lacks %q", marker)
		}
	}
	if !strings.Contains(terminal, "Revoked IAM capability grant is immutable") ||
		!strings.Contains(terminal, "iam_capability_grants_terminal_guard") {
		t.Fatal("capability terminal guard migration is incomplete")
	}
}
