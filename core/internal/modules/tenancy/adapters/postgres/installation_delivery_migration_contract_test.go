package postgres

import (
	"os"
	"strings"
	"testing"
)

func migrationWithPrefix(t *testing.T, origin, prefix string) string {
	t.Helper()
	entries, err := os.ReadDir(platformMigrationDirectory(origin))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) {
			return entry.Name()
		}
	}
	t.Fatalf("missing migration %s/%s", origin, prefix)
	return ""
}
