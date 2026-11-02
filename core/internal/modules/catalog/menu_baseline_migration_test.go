package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	emptyMenuBaselineVersion = "20260728103100"
	emptyMenuRepairVersion   = "20260728105100"
)

func catalogMenuMigration(t *testing.T, version string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(
		"migrations", "tenant", "fresh", version+"_*.fresh.sql",
	))
	if err != nil || len(matches) != 1 {
		t.Fatalf("catalog menu migration %s fresh count = %d err=%v",
			version, len(matches), err)
	}
	fresh, err := os.ReadFile(matches[0])
	if err != nil || len(fresh) == 0 {
		t.Fatalf("catalog menu migration %s did not load: %v", version, err)
	}
	return string(fresh)
}
