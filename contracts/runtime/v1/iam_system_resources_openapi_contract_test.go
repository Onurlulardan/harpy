package runtimev1

import (
	"path/filepath"
	"testing"
)

func TestIAMSystemResourceOpenAPIContractIsExactAndSecretFree(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	for _, resource := range []string{"users", "roles", "organizations"} {
		document := readOpenAPIDocument(t, filepath.Join(
			directory, "iam-"+resource+"-v1.paths.json",
		))
		assertLocalReferences(t, directory, document)
		items := document["pathItems"].(map[string]any)
		assertSystemResourceMethods(t, resource, items)
		assertCanonicalConflictRefs(t, resource, items)
	}
	for _, name := range []string{
		"iam-system-resources-v1.requests.json",
		"iam-system-resources-v1.records.json",
		"iam-system-resources-v1.schemas.json",
		"rest-query-v1.schemas.json",
	} {
		document := readOpenAPIDocument(t, filepath.Join(directory, name))
		assertLocalReferences(t, directory, document)
	}
	records := readOpenAPIDocument(t, filepath.Join(
		directory, "iam-system-resources-v1.records.json",
	))
	for _, kind := range []string{"UserRecord", "RoleRecord", "OrganizationRecord"} {
		schema := records["schemas"].(map[string]any)[kind].(map[string]any)
		properties := schema["properties"].(map[string]any)
		for _, secret := range []string{"password", "password_hash", "email_normalized"} {
			if properties[secret] != nil {
				t.Fatalf("%s exposes secret/core field %s", kind, secret)
			}
		}
	}
}

func assertCanonicalConflictRefs(
	t *testing.T, resource string, items map[string]any,
) {
	t.Helper()
	want := "iam-system-resources-v1.schemas.json#/responses/Conflict"
	for _, target := range []struct{ item, method string }{
		{"collection", "post"}, {"item", "put"},
	} {
		operation := items[target.item].(map[string]any)[target.method].(map[string]any)
		response := operation["responses"].(map[string]any)["409"].(map[string]any)
		if response["$ref"] != want {
			t.Fatalf("%s %s/%s recovery conflict=%v", resource, target.item, target.method, response)
		}
	}
}

func assertSystemResourceMethods(
	t *testing.T,
	resource string,
	items map[string]any,
) {
	t.Helper()
	want := map[string][]string{
		"collection": {"get", "post"},
		"query":      {"post"},
		"item":       {"parameters", "get", "put", "delete"},
		"state":      {"patch"},
	}
	if len(items) != len(want) {
		t.Fatalf("%s path item count=%d", resource, len(items))
	}
	for name, methods := range want {
		item, found := items[name].(map[string]any)
		if !found || len(item) != len(methods) {
			t.Fatalf("%s/%s is not exact: %v", resource, name, item)
		}
		for _, method := range methods {
			if item[method] == nil {
				t.Fatalf("%s/%s is missing %s", resource, name, method)
			}
		}
	}
}
