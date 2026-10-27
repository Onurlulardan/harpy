package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const iamFormsRecordAuthorityMigration = "20260716020056_seed_iam_forms_record_authority_v1"

func TestIAMFormsRecordAuthorityMigrationContract(t *testing.T) {
	fresh := readIAMIdentityMigration(t, "fresh", iamFormsRecordAuthorityMigration)
	adopted := readIAMIdentityMigration(t, "v1_adopted", iamFormsRecordAuthorityMigration)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("Forms record authority migration differs by origin")
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("Forms record authority migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect Forms record authority migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"DROP CONSTRAINT iam_permission_implications_kind_check",
		"'v1_form_parent_v1', 'forms_record_operation_v1'",
		"MERGE INTO iam_resources AS target",
		"MERGE INTO iam_actions AS target",
		"MERGE INTO iam_permissions AS target",
		"MERGE INTO iam_permission_implications AS target",
		"'forms.records', 'forms'", "'system_resource'",
		"'create'::text", "'read'", "'update'", "'delete'",
		"'import'", "'export'", "retirement_contract_key",
		"'forms.records.operation.v1'",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Forms record authority migration lacks %q", marker)
		}
	}
	upper := strings.ToUpper(source)
	for _, forbidden := range []string{
		"CREATE OR REPLACE", "DO $$", "EXECUTE ", "IF NOT EXISTS",
		"INSERT INTO IAM_USER_PERMISSION_BINDINGS",
		"INSERT INTO IAM_ROLE_PERMISSION_BINDINGS",
		"INSERT INTO IAM_ORGANIZATION_PERMISSION_BINDINGS",
		"INSERT INTO IAM_USER_ROLE_ASSIGNMENTS",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("Forms record authority migration contains %q", forbidden)
		}
	}
	assertIAMRBACMigrationSource(t, fresh)
}
