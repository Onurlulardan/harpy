package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestIAMWorkflowPublicationAuthorityMigrationIsPairedAndFailClosed(
	t *testing.T,
) {
	const name = "20260728104000_provision_iam_workflow_publication_authority_v1"
	fresh := readIAMIdentityMigration(t, "fresh", name)
	adopted := readIAMIdentityMigration(t, "v1_adopted", name)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("IAM workflow publication authority migration origin pair differs")
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("IAM workflow publication authority migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect IAM workflow publication authority migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"MERGE INTO iam_resources AS target",
		"'iam.capability_grant'::text",
		"'iam'::text",
		"'capability_grant'::text",
		"MERGE INTO iam_permissions AS target",
		"action.action_key = 'manage'",
		"target.match_kind = 'exact'",
		"target.is_protected IS DISTINCT FROM true",
		") = 1 THEN 1 ELSE 0 END",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM workflow publication authority migration lacks %q", marker)
		}
	}
	upper := strings.ToUpper(source)
	for _, forbidden := range []string{
		"IF NOT EXISTS",
		"CREATE OR REPLACE",
		"DO $$",
		"IAM_SERVICE_PRINCIPALS",
		"IAM_SERVICE_PRINCIPAL_CREDENTIALS",
		"IAM_CAPABILITY_GRANTS",
		"IAM_CAPABILITY_GRANT_REVISIONS",
	} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf(
				"IAM workflow publication authority migration contains %q",
				forbidden,
			)
		}
	}
}
