package runtimev1

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestIAMRBACOpenAPIContractIsExactAndFailClosed(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	paths := readOpenAPIDocument(t, filepath.Join(directory, "iam-rbac-v1.paths.json"))
	assertLocalReferences(t, directory, paths)
	items := paths["pathItems"].(map[string]any)
	want := map[string][]string{
		"collection": {"get", "post"}, "item": {"get", "put", "delete"},
		"resourceCollection": {"get", "post"},
		"resourceItem":       {"get", "put", "delete"},
		"actionCollection":   {"get", "post"},
		"actionItem":         {"get", "put", "delete"},
		"check":              {"post"}, "batchCheck": {"post"},
		"userEffective": {"get"}, "roleEffective": {"get"},
		"organizationEffective": {"get"}, "cacheStats": {"get"},
		"cacheInvalidateUser": {"post"}, "cacheInvalidateRole": {"post"},
		"cacheInvalidateOrganization": {"post"},
		"cacheInvalidateTenant":       {"post"}, "cacheInvalidateAll": {"post"},
	}
	if len(items) != len(want) {
		t.Fatalf("RBAC path item count=%d want=%d", len(items), len(want))
	}
	operations := 0
	for name, methods := range want {
		operations += len(methods)
		item := items[name].(map[string]any)
		if len(item) != len(methods) {
			t.Fatalf("RBAC path item %s=%v", name, item)
		}
		for _, method := range methods {
			operation := item[method].(map[string]any)
			assertRBACBearer(t, name, method, operation)
		}
	}
	if operations != 26 {
		t.Fatalf("RBAC operation count=%d want=26", operations)
	}
	assertRBACNoBodyAuthority(t, directory)
	assertRBACCacheRetirement(t, items)
}

func assertRBACBearer(
	t *testing.T,
	name string,
	method string,
	operation map[string]any,
) {
	t.Helper()
	security := operation["security"].([]any)
	if len(security) != 1 || security[0].(map[string]any)["bearerAuth"] == nil {
		t.Fatalf("RBAC %s %s security=%v", name, method, security)
	}
}

func assertRBACNoBodyAuthority(t *testing.T, directory string) {
	t.Helper()
	schemas := readOpenAPIDocument(t, filepath.Join(directory, "iam-rbac-v1.schemas.json"))
	assertLocalReferences(t, directory, schemas)
	definitions := schemas["schemas"].(map[string]any)
	for _, name := range []string{"Check", "BatchCheck"} {
		wire, err := json.Marshal(definitions[name])
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(wire))
		for _, forbidden := range []string{"user_id", "tenant_id", "super_admin", "all"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("RBAC decision request %s contains body authority %s", name, forbidden)
			}
		}
	}
}

func assertRBACCacheRetirement(t *testing.T, items map[string]any) {
	t.Helper()
	for _, name := range []string{
		"cacheStats", "cacheInvalidateUser", "cacheInvalidateRole",
		"cacheInvalidateOrganization", "cacheInvalidateTenant", "cacheInvalidateAll",
	} {
		item := items[name].(map[string]any)
		for _, raw := range item {
			operation := raw.(map[string]any)
			responses := operation["responses"].(map[string]any)
			if operation["deprecated"] != true || responses["410"] == nil || responses["200"] != nil {
				t.Fatalf("RBAC cache route %s is not explicit 410: %v", name, operation)
			}
		}
	}
}
