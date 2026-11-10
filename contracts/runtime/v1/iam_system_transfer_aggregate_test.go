package runtimev1

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestIAMSystemTransferAggregateIsAbsent(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	for path := range paths {
		if strings.Contains(path, "/record-exports") ||
			strings.Contains(path, "/record-imports") ||
			strings.Contains(path, "/record-import-sources") ||
			strings.Contains(path, "/profile-schemas/") {
			t.Fatalf("form engine path still published: %s", path)
		}
	}
}
