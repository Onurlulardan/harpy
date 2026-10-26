package runtimev1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func expectedCorePathCount(t *testing.T) int {
	t.Helper()
	directory := filepath.Join("..", "..", "http", "core-api-paths")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			t.Fatal(readErr)
		}
		var paths map[string]json.RawMessage
		if err = json.Unmarshal(content, &paths); err != nil || paths == nil {
			t.Fatalf("core path fragment %s is invalid: %v", entry.Name(), err)
		}
		count += len(paths)
	}
	if count == 0 {
		t.Fatal("core path fragment catalog is empty")
	}
	return count
}
