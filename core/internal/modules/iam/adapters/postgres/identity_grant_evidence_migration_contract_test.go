package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMIdentityGrantEvidenceMigrationIsPairedAndClosed(t *testing.T) {
	const name = "20260730113400_accept_iam_identity_grant_evidence_v1"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("IAM identity grant-evidence migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect IAM identity grant-evidence migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"DROP FUNCTION iam_identity_actor_assert_v1(",
		"DROP FUNCTION iam_identity_query_assert_v1(",
		"DROP FUNCTION iam_profile_actor_assert_v1(",
		"DROP FUNCTION iam_v1_actor_assertion_authority_v1(",
		"CREATE FUNCTION iam_identity_actor_assert_v1(",
		"CREATE FUNCTION iam_identity_query_assert_v1(",
		"CREATE FUNCTION iam_profile_actor_assert_v1(",
		"CREATE FUNCTION iam_v1_actor_assertion_authority_v1(",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"SESSION_USER IS DISTINCT FROM 'harpy_v1_runtime'",
		"REVOKE ALL ON FUNCTION iam_identity_actor_assert_v1(",
		"REVOKE ALL ON FUNCTION iam_identity_query_assert_v1(",
		"REVOKE ALL ON FUNCTION iam_profile_actor_assert_v1(",
		"REVOKE ALL ON FUNCTION iam_v1_actor_assertion_authority_v1(",
		") TO harpy_owner;",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM identity grant-evidence migration lacks %q", marker)
		}
	}
	if strings.Count(source, "iam_user_has_rbac_evidence_v2(") != 4 {
		t.Fatal("IAM identity grant-evidence migration does not bind four assertions")
	}
	if strings.Contains(source, "iam_user_has_permission_v1(") {
		t.Fatal("IAM identity grant-evidence migration retained exact-target helper")
	}
	if strings.Contains(source, "CREATE OR REPLACE") {
		t.Fatal("IAM identity grant-evidence migration replaces a routine in place")
	}
	if strings.Count(source, "SECURITY DEFINER") != 4 ||
		strings.Count(source, "FROM PUBLIC;") != 4 {
		t.Fatal("IAM identity grant-evidence functions are not closed")
	}
}
