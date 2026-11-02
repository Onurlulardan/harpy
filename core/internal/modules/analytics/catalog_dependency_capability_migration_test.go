package analytics

import (
	"os"
	"path/filepath"
	"testing"
)

func analyticsTenantMigration(t *testing.T, origin, logical string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(
		"migrations", "tenant", origin, logical+"."+origin+".sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
