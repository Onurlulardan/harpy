package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTenantMigrationsDoNotReachOwnerTables(t *testing.T) {
	paths, err := filepath.Glob("migrations/tenant/fresh/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		lower := strings.ToLower(string(body))
		for _, forbidden := range []string{
			"references iam_", "references forms_", "references workflow_", "references analytics_",
			"join iam_", "join forms_", "join workflow_", "join analytics_",
		} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("%s contains owner-external SQL %q", filepath.Base(path), forbidden)
			}
		}
	}
}
