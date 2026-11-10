package runtimev1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoDeprecatedHTTPOperationRemains(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		if !strings.HasSuffix(entry.Name(), ".json") && !strings.HasSuffix(entry.Name(), ".yaml") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var document any
		if err = json.Unmarshal(content, &document); err != nil {
			t.Fatalf("%s is not strict JSON: %v", entry.Name(), err)
		}
		assertNoDeprecatedOperation(t, entry.Name(), document)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertNoDeprecatedOperation(t *testing.T, name string, value any) {
	t.Helper()
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			if httpMethod(key) {
				operation, ok := child.(map[string]any)
				if ok && (operation["operationId"] != nil || operation["responses"] != nil) &&
					operation["deprecated"] == true {
					t.Errorf("%s still publishes deprecated %s %v", name, key, operation["operationId"])
				}
			}
			assertNoDeprecatedOperation(t, name, child)
		}
	case []any:
		for _, child := range current {
			assertNoDeprecatedOperation(t, name, child)
		}
	}
}

func httpMethod(value string) bool {
	switch value {
	case "get", "post", "put", "patch", "delete", "head", "options", "trace":
		return true
	default:
		return false
	}
}
