package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMFieldClassificationAuthorityMigrationIsPairedAndCanonical(
	t *testing.T,
) {
	const name = "20260730113200_provision_iam_field_classification_authority_v1"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	if len(fresh) == 0 {
		t.Fatal("fresh migration is empty")
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("IAM field-classification authority migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect IAM field-classification authority migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"MERGE INTO iam_actions AS target",
		"'read.confidential'::text",
		"'read.restricted'",
		"MERGE INTO iam_permissions AS target",
		"'iam.users'::text",
		"'iam.roles'",
		"'iam.organizations'",
		"target.match_kind = 'exact'",
		"target.is_protected IS DISTINCT FROM true",
		") = 6",
		"iam_field_classification_authority_assertion",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf(
				"IAM field-classification authority migration lacks %q",
				marker,
			)
		}
	}
	for _, forbidden := range []string{
		"IF NOT EXISTS",
		"CREATE OR REPLACE",
		"DO $$",
		"-- +GOOSE DOWN",
	} {
		if strings.Contains(strings.ToUpper(source), forbidden) {
			t.Fatalf(
				"IAM field-classification authority migration contains %q",
				forbidden,
			)
		}
	}
}
