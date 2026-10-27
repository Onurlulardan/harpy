package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const iamFormsLifecycleAuthorityMigration = "20260716020102_seed_iam_forms_lifecycle_authority_v1"

func TestIAMFormsLifecycleAuthorityMigrationContract(t *testing.T) {
	fresh := readIAMIdentityMigration(t, "fresh", iamFormsLifecycleAuthorityMigration)
	adopted := readIAMIdentityMigration(t, "v1_adopted", iamFormsLifecycleAuthorityMigration)
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("Forms lifecycle authority differs by origin")
	}
	if bytes.Count(fresh, []byte("\n"))+1 > 200 {
		t.Fatal("migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect Forms lifecycle authority migration: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"'forms.definition'", "'forms.publication'", "'system_resource'",
		"'read'", "'create'", "'update'", "'retire'", "'validate'",
		"'plan'", "'activate'", "'deactivate'", "MERGE INTO iam_permissions",
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
