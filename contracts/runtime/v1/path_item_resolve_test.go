package runtimev1

import (
	"path/filepath"
	"strings"
	"testing"
)

func resolveFormsPathItem(t *testing.T, directory string, value any) map[string]any {
	t.Helper()
	item, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("path item is invalid: %#v", value)
	}
	reference, referenced := item["$ref"].(string)
	if !referenced {
		return item
	}
	parts := strings.SplitN(reference, "#", 2)
	target := filepath.Join(directory, filepath.FromSlash(parts[0]))
	document := readOpenAPIDocument(t, target)
	var current any = document
	if len(parts) == 2 {
		for _, token := range strings.Split(strings.TrimPrefix(parts[1], "/"), "/") {
			name := strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
			next, valid := current.(map[string]any)[name]
			if !valid {
				t.Fatalf("OpenAPI reference %s misses %s", reference, name)
			}
			current = next
		}
	}
	return resolveFormsPathItem(t, directory, current)
}

func assertFormsTenantHeaders(t *testing.T, directory, path string, item map[string]any) {
	t.Helper()
	parameters, ok := item["parameters"].([]any)
	if !ok {
		t.Fatalf("tenant parameters missing: %s", path)
	}
	foundID, foundSlug := false, false
	for _, value := range parameters {
		parameter := resolveFormsPathItem(t, directory, value)
		foundID = foundID || parameter["name"] == "X-Tenant-ID"
		foundSlug = foundSlug || parameter["name"] == "X-Tenant-Slug"
	}
	if !foundID || !foundSlug {
		t.Fatalf("tenant headers incomplete: %s", path)
	}
}
