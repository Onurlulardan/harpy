package postgres

import (
	"strings"
	"testing"
)

func TestIAMUserRelationReplaceMigrationIsSingleAtomicGraphWriter(t *testing.T) {
	source := string(readIAMIdentityMigration(
		t, "fresh", "20260716020051_create_iam_identity_user_relations_replace_v1",
	))
	for _, required := range []string{
		"CREATE FUNCTION iam_identity_user_relations_replace_v1(",
		"p_replace_organizations boolean", "p_replace_roles boolean",
		"p_expected_revision bigint", "FOR UPDATE",
		"UPDATE iam_users SET", "authorization_version = authorization_version +",
		"revision = revision + 1", "iam_identity_record_v1(",
		"p_user_id = p_actor_id", "role.is_protected",
		"membership.membership_kind = 'owner'", "NOT BETWEEN 0 AND 100",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("user relation writer is missing %q", required)
		}
	}
	if strings.Count(source, "iam_identity_record_v1(") != 1 ||
		strings.Contains(source, "iam_identity_membership_add_v1(") ||
		strings.Contains(source, "iam_identity_assignment_grant_v1(") {
		t.Fatal("user relation writer chains leaf commands or emits multiple receipts")
	}
}

func TestIAMUserRelationReplaceGrantUsesExactArraySignature(t *testing.T) {
	source := string(readIAMIdentityMigration(
		t, "fresh", "20260716020052_grant_iam_identity_user_relations_replace_v1",
	))
	for _, required := range []string{
		"GRANT EXECUTE ON FUNCTION iam_identity_user_relations_replace_v1(",
		"boolean,uuid[],uuid[],boolean,uuid[],uuid[]",
		") TO harpy_runtime;",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("user relation grant is missing %q", required)
		}
	}
}
