package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const (
	iamWorkAccessAuthorityMigration = "20260802235101_publish_iam_work_access_authority_v1"
	iamWorkAccessGrantMigration     = "20260802235102_grant_iam_work_access_authority_v1"
)

func TestIAMWorkAccessAuthorityMigrationContract(t *testing.T) {
	fresh := readIAMIdentityMigration(t, "fresh", iamWorkAccessAuthorityMigration)
	adopted := readIAMIdentityMigration(t, "v1_adopted", iamWorkAccessAuthorityMigration)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("Work access authority differs by origin")
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect Work access authority migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"'work.boards'", "'work.columns'", "'work.memberships'", "'work.tasks'",
		"'work.comments'", "'work.labels'", "'work.checklists'",
		"'work.calendar'", "'work.recurrence'", "'work.inbox'",
		"'read'", "'create'", "'update'", "'archive'", "'restore'",
		"'manage'", "'move'", "'assign'", "'transition'", "'redact'",
		"MERGE INTO iam_permissions", "iam_work_access_authority_assertion",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("migration lacks %q", marker)
		}
	}
	upper := strings.ToUpper(source)
	for _, forbidden := range []string{"CREATE OR REPLACE", "DO $$", "EXECUTE ", "IF NOT EXISTS"} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("migration contains %q", forbidden)
		}
	}
	assertIAMRBACMigrationSource(t, fresh)
}

func TestIAMWorkAccessGrantMigrationContract(t *testing.T) {
	fresh := readIAMIdentityMigration(t, "fresh", iamWorkAccessGrantMigration)
	adopted := readIAMIdentityMigration(t, "v1_adopted", iamWorkAccessGrantMigration)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("Work access grant differs by origin")
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("grant migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect Work access grant migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"'work.%'", "'Tenant Yöneticisi'", "'Yönetici'",
		"INSERT INTO iam_role_permission_bindings",
		"SELECT iam_rbac_bump_all_users_v1(clock_value)",
		"iam_work_access_grant_assertion",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("grant migration lacks %q", marker)
		}
	}
	for _, forbidden := range []string{"Tenant Admin", "CREATE OR REPLACE", "DO $$"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("grant migration contains %q", forbidden)
		}
	}
	assertIAMRBACMigrationSource(t, fresh)
}

func TestIAMWorkAccessReleaseFrontierMigrationContract(t *testing.T) {
	const name = "20260802235100_assert_iam_work_access_release_frontier_v1"
	fresh := readIAMPlatformAuthMigration(t, "fresh", name)
	adopted := readIAMPlatformAuthMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("Work access frontier differs by origin")
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("frontier migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect Work access frontier migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"Work access authority owns tenant database generation 96",
		"Advance the platform frontier",
		"BETWEEN 180000 AND 189999",
		"role_record.rolname = 'harpy_runtime'",
		"role_record.rolname = 'harpy_v2_runtime'",
		"iam_work_access_release_frontier_assertion",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("frontier migration lacks %q", marker)
		}
	}
	assertIAMPlatformAuthMigrationSource(t, fresh)
}
