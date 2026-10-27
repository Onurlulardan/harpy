package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamAuthorizationRuntimeMigrationNames = []string{
	"20260716013424_create_iam_authorization_evidence_read_v1",
	"20260716013425_grant_iam_authorization_evidence_read_v1",
	"20260716020057_create_iam_authorization_evidence_read_v2",
	"20260716020058_grant_iam_authorization_evidence_read_v2",
	"20260716020100_create_iam_authorization_evidence_read_v3",
	"20260716020101_grant_iam_authorization_evidence_read_v3",
}

func TestIAMAuthorizationEvidenceV3IsReadOnlySnapshotSafe(t *testing.T) {
	source := string(readIAMIdentityMigration(
		t, "fresh", iamAuthorizationRuntimeMigrationNames[4],
	))
	for _, marker := range []string{
		"iam_authorization_evidence_read_v3", "tenancy_schema_identity",
		"platform.cutover_groups", "platform.domain_cutovers",
		"platform.cutover_group_members",
		"iam_authorization_evidence_read_v1",
		"REVOKE ALL ON FUNCTION iam_authorization_evidence_read_v3",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("authorization evidence V3 lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"STABLE", "iam_assert_current_tenant_authority_v1",
		"FOR SHARE", "FOR UPDATE", "INSERT ", "UPDATE ", "DELETE ",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("authorization evidence V3 is not read-only: %q", forbidden)
		}
	}
}

func TestIAMAuthorizationRuntimeMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range iamAuthorizationRuntimeMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and v1_adopted authorization runtime differs")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("authorization runtime migration is invalid: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{
				"IF NOT EXISTS", "-- +GOOSE DOWN", "DO $$", "CREATE OR REPLACE",
				"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
				"ON DELETE CASCADE", "ON DELETE SET NULL",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("authorization runtime contains forbidden token %q", forbidden)
				}
			}
		})
	}
}

func TestIAMAuthorizationEvidenceV2FencesCrossOwnerDecision(t *testing.T) {
	source := string(readIAMIdentityMigration(
		t, "fresh", iamAuthorizationRuntimeMigrationNames[2],
	))
	for _, marker := range []string{
		"iam_authorization_evidence_read_v2",
		"FROM tenancy_schema_identity",
		"FROM platform.cutover_groups",
		"FROM iam_users WHERE id = p_actor_id FOR SHARE",
		"FROM iam_authorization_clock",
		"FOR SHARE",
		"iam_authorization_evidence_read_v1",
		"REVOKE ALL ON FUNCTION iam_authorization_evidence_read_v2",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("authorization evidence V2 lacks %q", marker)
		}
	}
	if strings.Contains(source, "iam_assert_current_tenant_authority_v1") {
		t.Fatal("authorization evidence V2 delegates its authority decision")
	}
	grant := string(readIAMIdentityMigration(
		t, "fresh", iamAuthorizationRuntimeMigrationNames[3],
	))
	if strings.Count(grant, "GRANT EXECUTE") != 1 ||
		!strings.Contains(grant, "TO harpy_runtime") {
		t.Fatal("authorization evidence V2 execute grant is not exact")
	}
}

func TestIAMAuthorizationEvidenceRoutineIsRuntimeClosed(t *testing.T) {
	source := string(readIAMIdentityMigration(
		t, "fresh", iamAuthorizationRuntimeMigrationNames[0],
	))
	for _, marker := range []string{
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
		"p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()",
		"FROM iam_user_permission_bindings", "FROM iam_user_role_assignments",
		"FROM iam_organization_memberships", "FROM iam_permission_implications",
		"REVOKE ALL ON FUNCTION iam_authorization_evidence_read_v1",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("authorization evidence routine lacks %q", marker)
		}
	}
	grant := string(readIAMIdentityMigration(
		t, "fresh", iamAuthorizationRuntimeMigrationNames[1],
	))
	if strings.Count(grant, "GRANT EXECUTE") != 1 ||
		!strings.Contains(grant, "TO harpy_runtime") {
		t.Fatal("authorization evidence execute grant is not exact")
	}
}
