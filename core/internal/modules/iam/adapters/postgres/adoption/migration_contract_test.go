package adoption

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func adoptionMigrationPaths(t *testing.T, directory, suffix string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read IAM V1 adoption migrations: %v", err)
	}
	result := make([]string, 0, 15)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "202607160139") && strings.HasSuffix(entry.Name(), suffix) {
			result = append(result, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(result)
	return result
}

func readMigration(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration %s: %v", path, err)
	}
	return content
}
