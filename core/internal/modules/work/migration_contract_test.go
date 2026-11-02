package work

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationsDoNotReferenceExternalOwnerTables(t *testing.T) {
	paths, err := filepath.Glob("migrations/tenant/fresh/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(body))
		for _, forbidden := range []string{
			"references iam_", "references forms_", "references workflow_",
			"join iam_", "join forms_", "join workflow_",
		} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("%s contains owner-external SQL %q", filepath.Base(path), forbidden)
			}
		}
	}
}
