package analytics

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func analyticsCutoverFiles(t *testing.T, scope, pattern string, want int) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("migrations", scope, "fresh", pattern))
	if err != nil || len(paths) != want {
		t.Fatalf("Analytics %s cutover migrations=%d want=%d err=%v",
			scope, len(paths), want, err)
	}
	return paths
}

func analyticsCutoverPairs(t *testing.T, scope string, fresh []string) string {
	t.Helper()
	var combined strings.Builder
	for _, freshPath := range fresh {
		body, err := os.ReadFile(freshPath)
		if err != nil || len(body) == 0 {
			t.Fatalf("Analytics cutover migration did not load %s: %v", freshPath, err)
		}
		if bytes.Count(body, []byte{'\n'}) > 200 {
			t.Fatalf("Analytics cutover migration exceeds 200 lines: %s", freshPath)
		}
		combined.Write(body)
	}
	return strings.ToLower(combined.String())
}
