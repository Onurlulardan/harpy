package async

import (
	"strings"
	"testing"
)

func migrationVersion(t *testing.T, path string) string {
	t.Helper()
	base := path[strings.LastIndex(path, "/")+1:]
	if len(base) < 14 {
		t.Fatalf("migration path has no version: %s", path)
	}
	return base[:14]
}
