package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMOperatorAdminAuthorityMigrationIsPairedAndCanonical(t *testing.T) {
	const name = "20260727150000_reconcile_iam_operator_admin_authority_v1"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("IAM operator-admin authority migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect IAM operator-admin authority migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"'iam.users'", "'iam.roles'", "'iam.organizations'",
		"'iam.permissions'", "'iam.permission_requests'",
		"'create'", "'read'", "'update'", "'delete'", "'manage'",
		"target.match_kind = 'exact'",
		"target.is_protected IS DISTINCT FROM true",
		") = 18 THEN 1 ELSE 0 END",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM operator-admin authority migration lacks %q", marker)
		}
	}
}

func TestIAMInstanceOwnerPlatformBoundaryIsPaired(t *testing.T) {
	const name = "20260727145959_assert_iam_instance_owner_platform_authority_v1"
	fresh := readIAMPlatformAuthMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	source := string(fresh)
	for _, marker := range []string{
		"permission.permission_key = 'environment_all'",
		"permission.match_kind = 'environment_all'",
		"permission.owner_only",
		"iam_instance_owner_platform_authority_assertion",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM Instance Owner platform boundary lacks %q", marker)
		}
	}
}
