package runtimev1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func coreFragmentPaths(t *testing.T) map[string]struct{} {
	t.Helper()
	directory := filepath.Join("..", "..", "http", "core-api-paths")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			t.Fatal(readErr)
		}
		var fragment map[string]json.RawMessage
		if err = json.Unmarshal(content, &fragment); err != nil || fragment == nil {
			t.Fatalf("core path fragment %s is invalid: %v", entry.Name(), err)
		}
		for path := range fragment {
			if _, exists := paths[path]; exists {
				t.Fatalf("core path %s is owned twice", path)
			}
			paths[path] = struct{}{}
		}
	}
	if len(paths) == 0 {
		t.Fatal("core path fragment catalog is empty")
	}
	return paths
}
