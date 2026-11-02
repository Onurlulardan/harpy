package analytics

import (
	"os"
	"path/filepath"
	"testing"
)

func readAnalyticsSnapshotMigration(t *testing.T, origin, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(
		"migrations", "platform", origin, name,
	))
	if err != nil {
		t.Fatal(err)
	}
	return content
}
