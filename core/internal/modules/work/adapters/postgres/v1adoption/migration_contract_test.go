package v1adoption

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func workCutoverPaths(t *testing.T, directory, suffix string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	paths := make([]string, 0, 14)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "202607182301") &&
			strings.HasSuffix(entry.Name(), suffix) {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(paths)
	return paths
}

func readWorkCutoverMigration(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Work cutover migration %s: %v", path, err)
	}
	return content
}
