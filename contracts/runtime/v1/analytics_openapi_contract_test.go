package runtimev1

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestAnalyticsOpenAPIRoutesMatchRuntimeSurface(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	corePath := filepath.Join(directory, "core-api-v1.openapi.yaml")
	core := readOpenAPIDocument(t, corePath)
	paths := core["paths"].(map[string]any)
	wanted := map[string][]string{
		"/api/analytics/queries":                                      {"get", "post"},
		"/api/analytics/queries/{query_id}":                           {"delete", "get", "patch"},
		"/api/analytics/queries/{query_id}/publications":              {"post"},
		"/api/analytics/queries/{query_id}/revalidations":             {"post"},
		"/api/analytics/queries/{query_id}/executions":                {"post"},
		"/api/analytics/dashboards":                                   {"get", "post"},
		"/api/analytics/dashboard-defaults/{surface}":                 {"get"},
		"/api/analytics/dashboards/{dashboard_id}":                    {"delete", "get", "patch"},
		"/api/analytics/dashboards/{dashboard_id}/publications":       {"post"},
		"/api/analytics/dashboards/{dashboard_id}/acl-grants":         {"post"},
		"/api/analytics/dashboards/{dashboard_id}/acl-revocations":    {"post"},
		"/api/analytics/dashboards/{dashboard_id}/defaults/{surface}": {"put"},
		"/api/analytics/publications":                                 {"get", "post"},
		"/api/analytics/publications/{release_id}":                    {"get"},
		"/api/analytics/publications/{release_id}/submissions":        {"post"},
		"/api/analytics/publications/{release_id}/approvals":          {"post"},
		"/api/analytics/publications/{release_id}/activations":        {"post"},
		"/api/analytics/publications/{release_id}/revocations":        {"post"},
		"/api/published/queries/{public_id}/{slug}":                   {"get", "post"},
		"/api/published/dashboards/{public_id}/{slug}":                {"get"},
		"/api/openapi/execute/{slug}":                                 {"get", "post"},
		"/api/analytics/execute/{slug}":                               {"get", "post"},
	}
	operations := 0
	for path, methods := range wanted {
		item := resolveAnalyticsPathItem(t, corePath, paths[path])
		found := analyticsMethods(item)
		sort.Strings(methods)
		if strings.Join(found, ",") != strings.Join(methods, ",") {
			t.Fatalf("Analytics methods for %s = %v, want %v", path, found, methods)
		}
		operations += len(found)
		assertAnalyticsPathContract(t, corePath, path, item, found)
	}
	if operations != 32 {
		t.Fatalf("Analytics operation count = %d, want 32", operations)
	}
}

func analyticsMethods(item map[string]any) []string {
	result := []string{}
	for _, method := range []string{"delete", "get", "patch", "post", "put"} {
		if _, exists := item[method].(map[string]any); exists {
			result = append(result, method)
		}
	}
	return result
}

func assertAnalyticsPathContract(
	t *testing.T, source, path string, item map[string]any, methods []string,
) {
	t.Helper()
	public := strings.HasPrefix(path, "/api/published/")
	legacy := path == "/api/openapi/execute/{slug}" || path == "/api/analytics/execute/{slug}"
	if public && item["x-harpy-tenant-headers"] != "forbidden" {
		t.Fatalf("public Analytics path accepts tenant headers: %s", path)
	}
	if legacy {
		values, ok := item["x-harpy-exactly-one-tenant-header"].([]any)
		if !ok || len(values) != 2 {
			t.Fatalf("legacy Analytics tenant tuple is not exact: %s", path)
		}
	}
	if !public && !legacy {
		assertAnalyticsTenantHeaders(t, source, path, item)
	}
	for _, method := range methods {
		operation := item[method].(map[string]any)
		if operation["operationId"] == "" || operation["x-harpy-authority"] == "" {
			t.Fatalf("Analytics operation is incomplete: %s %s", method, path)
		}
		responses, ok := operation["responses"].(map[string]any)
		if !ok || responses["default"] == nil {
			t.Fatalf("Analytics fail-closed response is missing: %s %s", method, path)
		}
		if !public && !legacy {
			security, secured := operation["security"].([]any)
			if !secured || len(security) != 1 || operation["x-harpy-permission"] == nil {
				t.Fatalf("Analytics auth contract is incomplete: %s %s", method, path)
			}
		}
	}
}

func assertAnalyticsTenantHeaders(t *testing.T, source, path string, item map[string]any) {
	t.Helper()
	parameters, ok := item["parameters"].([]any)
	if !ok {
		t.Fatalf("Analytics tenant parameters missing: %s", path)
	}
	foundID, foundSlug := false, false
	for _, value := range parameters {
		parameter := resolveAnalyticsPathItem(t, source, value)
		foundID = foundID || parameter["name"] == "X-Tenant-ID"
		foundSlug = foundSlug || parameter["name"] == "X-Tenant-Slug"
	}
	if !foundID || !foundSlug {
		t.Fatalf("Analytics tenant headers incomplete: %s", path)
	}
}

func TestAnalyticsOpenAPIDocumentsResolveLocalFiles(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	files := []string{
		"analytics-common-v2.schemas.json", "analytics-query-v2.schemas.json",
		"analytics-dashboard-v2.schemas.json", "analytics-publication-v2.schemas.json",
		"analytics-query-v2.paths.json", "analytics-dashboard-v2.paths.json",
		"analytics-publication-v2.paths.json", "analytics-v1-compat.paths.json",
	}
	for _, name := range files {
		path := filepath.Join(directory, name)
		document := readOpenAPIDocument(t, path)
		assertLocalReferences(t, directory, document)
	}
}

func resolveAnalyticsPathItem(t *testing.T, source string, value any) map[string]any {
	t.Helper()
	item, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("Analytics reference value is invalid: %#v", value)
	}
	reference, referenced := item["$ref"].(string)
	if !referenced {
		return item
	}
	parts := strings.SplitN(reference, "#", 2)
	target := source
	if parts[0] != "" {
		target = filepath.Join(filepath.Dir(source), filepath.FromSlash(parts[0]))
	}
	document := readOpenAPIDocument(t, target)
	var current any = document
	if len(parts) == 2 {
		for _, token := range strings.Split(strings.TrimPrefix(parts[1], "/"), "/") {
			name := strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
			current = current.(map[string]any)[name]
		}
	}
	return resolveAnalyticsPathItem(t, target, current)
}
