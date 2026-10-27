package runtimev1

import (
	"fmt"
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

func TestIAMSystemTransferRetiredV1AliasesAreExact(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t, filepath.Join(directory, "core-api-v1.openapi.yaml"))
	paths := core["paths"].(map[string]any)
	routes := map[string]string{
		"export":                              "get",
		"import/analyze":                      "post",
		"import/execute":                      "post",
		"import/jobs":                         "get",
		"import/jobs/{job_id}":                "get",
		"import/jobs/{job_id}/rows":           "get",
		"import/jobs/{job_id}/resume":         "post",
		"import/jobs/{job_id}/cancel":         "post",
		"import/jobs/{job_id}/report":         "get",
		"import/sessions/{session_id}/cancel": "post",
	}
	for _, resource := range []string{"users", "roles", "organizations"} {
		for suffix, method := range routes {
			path := fmt.Sprintf("/api/%s/%s", resource, suffix)
			item := resolveFormsPathItem(t, directory, paths[path])
			operation, valid := item[method].(map[string]any)
			if !valid || operation["deprecated"] != true ||
				operation["x-harpy-successor"] == "" {
				t.Fatalf("retired IAM transfer alias is incomplete: %s %s", method, path)
			}
			security, valid := operation["security"].([]any)
			responses, responseValid := operation["responses"].(map[string]any)
			if !valid || len(security) != 0 || !responseValid || responses["426"] == nil {
				t.Fatalf("retired IAM transfer alias is not fail-closed: %s %s", method, path)
			}
		}
	}
}
