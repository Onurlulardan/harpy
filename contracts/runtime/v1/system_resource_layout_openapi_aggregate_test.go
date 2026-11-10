package runtimev1

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSystemResourceLayoutOpenAPIAggregateIsAbsent(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	for path := range paths {
		if strings.Contains(path, "/layout-revisions") || strings.HasSuffix(path, "/presentation") {
			t.Fatalf("layout path still published: %s", path)
		}
	}
}
