package runtimev1

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestSystemResourceLayoutOpenAPIAggregateIsExact(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	wanted := map[string][]string{
		"/api/system-resources/{resource}/layout-revisions": {"get"},
		"/api/system-resources/{resource}/layout-revisions/{revision_id}": {
			"delete", "get", "patch", "put",
		},
		"/api/system-resources/{resource}/layout-revisions/{revision_id}/publication": {"put"},
		"/api/system-resources/{resource}/presentation":                               {"get"},
	}
	for path, methods := range wanted {
		item := resolveFormsPathItem(t, directory, paths[path])
		assertFormsTenantHeaders(t, path, item)
		var found []string
		for _, method := range []string{"delete", "get", "patch", "post", "put"} {
			operation, exists := item[method].(map[string]any)
			if !exists {
				continue
			}
			found = append(found, method)
			if operation["operationId"] == "" || operation["security"] == nil ||
				operation["x-harpy-permission"] == "" ||
				operation["x-harpy-authority"] == "" {
				t.Fatalf("incomplete layout operation %s %s", method, path)
			}
			responses, valid := operation["responses"].(map[string]any)
			if !valid || responses["default"] == nil {
				t.Fatalf("layout error contract missing for %s %s", method, path)
			}
		}
		sort.Strings(methods)
		if strings.Join(found, ",") != strings.Join(methods, ",") {
			t.Fatalf("layout methods for %s = %v, want %v", path, found, methods)
		}
	}
}
