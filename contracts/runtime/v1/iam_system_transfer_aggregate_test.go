package runtimev1

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestIAMSystemTransferAggregateIsExactAndSecured(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	wanted := map[string][]string{
		"/api/iam/system-resources/{resource}/record-exports":                     {"post"},
		"/api/iam/system-resources/{resource}/record-exports/{export_id}":         {"delete", "get"},
		"/api/iam/system-resources/{resource}/record-exports/{export_id}/content": {"get"},
		"/api/iam/system-resources/{resource}/record-import-sources":              {"post"},
		"/api/iam/system-resources/{resource}/record-imports":                     {"post"},
		"/api/iam/system-resources/{resource}/record-imports/{import_id}":         {"delete", "get"},
		"/api/iam/system-resources/{resource}/record-imports/{import_id}/report":  {"get"},
	}
	for path, methods := range wanted {
		item := resolveFormsPathItem(t, directory, paths[path])
		assertFormsTenantHeaders(t, directory, path, item)
		var found []string
		for _, method := range []string{"delete", "get", "post"} {
			operation, exists := item[method].(map[string]any)
			if !exists {
				continue
			}
			found = append(found, method)
			if operation["operationId"] == "" || operation["security"] == nil ||
				operation["x-harpy-permission-template"] == "" ||
				operation["x-harpy-authority"] == "" {
				t.Fatalf("incomplete IAM transfer operation %s %s", method, path)
			}
			responses, valid := operation["responses"].(map[string]any)
			if !valid || responses["default"] == nil {
				t.Fatalf("IAM transfer errors missing for %s %s", method, path)
			}
		}
		sort.Strings(methods)
		if strings.Join(found, ",") != strings.Join(methods, ",") {
			t.Fatalf("IAM transfer methods for %s = %v, want %v", path, found, methods)
		}
	}
}
