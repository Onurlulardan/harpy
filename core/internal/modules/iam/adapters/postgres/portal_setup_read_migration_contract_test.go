package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamPortalSetupReadMigrationNames = []string{
	"20260730145000_create_iam_portal_setup_read_v1",
	"20260730145001_grant_iam_portal_setup_read_v1",
}

func TestIAMPortalSetupReadMigrationsArePairedAndBounded(t *testing.T) {
	var joined strings.Builder
	for _, name := range iamPortalSetupReadMigrationNames {
		fresh := readIAMIdentityMigration(t, "fresh", name)
		adopted := readIAMIdentityMigration(t, "v1_adopted", name)
		if !bytes.Equal(fresh, adopted) {
			t.Fatalf("IAM portal setup read origin pair differs: %s", name)
		}
		if bytes.Count(fresh, []byte("\n"))+1 > 200 {
			t.Fatalf("IAM portal setup read migration exceeds 200 lines: %s", name)
		}
		if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
			t.Fatalf("inspect IAM portal setup read migration %s: %v", name, err)
		}
		joined.Write(fresh)
	}
	source := joined.String()
	for _, marker := range []string{
		"CREATE FUNCTION iam_portal_setup_read_v1(",
		"iam_identity_query_assert_v1(",
		"FROM platform.tenants AS tenant",
		"REVOKE ALL ON FUNCTION iam_portal_setup_read_v1(",
		"GRANT EXECUTE ON FUNCTION iam_portal_setup_read_v1(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM portal setup read migrations lack %q", marker)
		}
	}
	for _, forbidden := range []string{
		"CREATE OR REPLACE", "-- +goose Down", "source_evidence_sha256",
		"decision_actor_id", "provenance",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("IAM portal setup read migrations contain %q", forbidden)
		}
	}
}
