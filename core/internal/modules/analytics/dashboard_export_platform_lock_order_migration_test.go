package analytics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readDashboardExportPlatformMigration(t *testing.T, origin, logical string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("migrations", "tenant", origin,
		logical+"."+origin+".sql"))
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func dashboardExportPlatformFunctionBody(t *testing.T, source, function string) string {
	t.Helper()
	start := strings.Index(source, "CREATE FUNCTION "+function+"(")
	if start < 0 {
		t.Fatalf("dashboard export Platform wrapper is missing: %s", function)
	}
	end := strings.Index(source[start:], "$function$;")
	if end < 0 {
		t.Fatalf("dashboard export Platform wrapper is unterminated: %s", function)
	}
	return source[start : start+end]
}
