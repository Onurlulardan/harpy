package runtimev1

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestFormsV1DefinitionAggregateIsExactAndSecured(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	wanted := map[string][]string{
		"/api/form-engine/configs":                                      {"get", "post"},
		"/api/form-engine/configs/{id}":                                 {"delete", "get", "put"},
		"/api/form-engine/configs/by-slug/{slug}":                       {"get"},
		"/api/form-engine/configs/validate":                             {"post"},
		"/api/form-engine/configs/check-slug":                           {"get"},
		"/api/form-engine/configs/{id}/versions":                        {"get"},
		"/api/form-engine/configs/{id}/versions/{version_id}/publish":   {"post"},
		"/api/form-engine/configs/{id}/versions/{version_id}/activate":  {"post"},
		"/api/form-engine/configs/{id}/versions/{version_id}/unpublish": {"post"},
	}
	for path, methods := range wanted {
		item := resolveFormsPathItem(t, directory, paths[path])
		assertFormsTenantHeaders(t, path, item)
		var found []string
		for _, method := range []string{"delete", "get", "post", "put"} {
			operation, exists := item[method].(map[string]any)
			if !exists {
				continue
			}
			found = append(found, method)
			if operation["deprecated"] != true || operation["operationId"] == "" ||
				operation["security"] == nil || operation["x-harpy-permission"] == "" ||
				operation["x-harpy-authority"] == "" {
				t.Fatalf("incomplete Forms V1 definition operation %s %s", method, path)
			}
			responses, valid := operation["responses"].(map[string]any)
			if !valid || responses["default"] == nil {
				t.Fatalf("Forms V1 definition errors missing for %s %s", method, path)
			}
		}
		sort.Strings(methods)
		if strings.Join(found, ",") != strings.Join(methods, ",") {
			t.Fatalf("Forms V1 definition methods for %s = %v, want %v", path, found, methods)
		}
	}
}
