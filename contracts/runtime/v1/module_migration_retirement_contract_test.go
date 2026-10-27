package runtimev1

import (
	"path/filepath"
	"testing"
)

func TestStandaloneModuleMigrationOpenAPIIsRejectOnly(t *testing.T) {
	t.Parallel()
	directory := filepath.Join("..", "..", "http")
	document := readOpenAPIDocument(t,
		filepath.Join(directory, "catalog-v1-module-migration-retired.paths.json"))
	items, ok := document["pathItems"].(map[string]any)
	if !ok || len(items) != 4 {
		t.Fatalf("retired path items = %v", items)
	}
	for name, raw := range items {
		item := raw.(map[string]any)
		for method, operationRaw := range item {
			if method == "parameters" {
				continue
			}
			operation := operationRaw.(map[string]any)
			responses := operation["responses"].(map[string]any)
			if operation["deprecated"] != true || len(responses) != 1 || responses["410"] == nil {
				t.Fatalf("%s is not reject-only: %v", name, operation)
			}
		}
	}
}
