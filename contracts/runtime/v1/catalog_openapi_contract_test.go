package runtimev1

import (
	"path/filepath"
	"strings"
	"testing"
)

type catalogOperation struct {
	method, permission, authority string
	idempotent                    bool
}

func TestCatalogOwnedOpenAPIFragmentIsExact(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	fragment := readOpenAPIDocument(t, filepath.Join(directory, "core-api-paths", "catalog.json"))
	paths := fragment
	want := catalogOpenAPIOperations()
	if len(paths) != 18 || len(want) != 19 {
		t.Fatalf("Catalog paths=%d operations=%d", len(paths), len(want))
	}
	for key, expected := range want {
		parts := strings.SplitN(key, " ", 2)
		item := resolveFormsPathItem(t, directory, paths[parts[1]])
		assertFormsTenantHeaders(t, parts[1], item)
		operation, ok := item[strings.ToLower(expected.method)].(map[string]any)
		if !ok || operation["x-harpy-permission"] != expected.permission ||
			operation["x-harpy-authority"] != expected.authority {
			t.Fatalf("Catalog operation %s is incomplete: %v", key, operation)
		}
		assertCatalogBearer(t, key, operation)
		if expected.idempotent {
			assertCatalogIdempotency(t, key, item, operation)
		}
		responses, ok := operation["responses"].(map[string]any)
		if !ok || responses["default"] == nil {
			t.Fatalf("Catalog operation %s has no default error", key)
		}
	}
}

func catalogOpenAPIOperations() map[string]catalogOperation {
	return map[string]catalogOperation{
		"GET /api/catalog/packages":                                         {"GET", "catalog.read", "catalog.package.list", false},
		"GET /api/catalog/packages/{package_id}":                            {"GET", "catalog.read", "catalog.package.get", false},
		"GET /api/catalog/packages/{package_id}/versions":                   {"GET", "catalog.read", "catalog.package_version.list", false},
		"GET /api/catalog/packages/{package_id}/versions/{version}":         {"GET", "catalog.read", "catalog.package_version.get", false},
		"GET /api/catalog/installations":                                    {"GET", "catalog.read", "catalog.installation.list", false},
		"POST /api/catalog/installations":                                   {"POST", "catalog.install", "catalog.installation.install", true},
		"GET /api/catalog/installations/{installation_id}":                  {"GET", "catalog.read", "catalog.installation.get", false},
		"POST /api/catalog/installations/{installation_id}/upgrades":        {"POST", "catalog.upgrade", "catalog.installation.upgrade", true},
		"POST /api/catalog/installations/{installation_id}/disables":        {"POST", "catalog.disable", "catalog.installation.disable", true},
		"POST /api/catalog/installations/{installation_id}/enables":         {"POST", "catalog.enable", "catalog.installation.enable", true},
		"POST /api/catalog/installations/{installation_id}/restorations":    {"POST", "catalog.restore", "catalog.installation.restore", true},
		"POST /api/catalog/installations/{installation_id}/retirements":     {"POST", "catalog.retire", "catalog.installation.retire", true},
		"POST /api/catalog/installations/{installation_id}/purge-approvals": {"POST", "catalog.purge", "catalog.installation.purge.approve", true},
		"POST /api/catalog/installations/{installation_id}/purges":          {"POST", "catalog.purge", "catalog.installation.purge", true},
		"GET /api/catalog/operations/{operation_id}":                        {"GET", "catalog.read", "catalog.operation.get", false},
		"POST /api/catalog/operations/{operation_id}/cancellations":         {"POST", "catalog.install", "catalog.operation.cancel", true},
		"POST /api/catalog/operations/{operation_id}/retries":               {"POST", "catalog.install", "catalog.operation.retry", true},
		"GET /api/catalog/menu":                                             {"GET", "catalog.read", "catalog.menu.get", false},
		"PATCH /api/catalog/menu/items/{item_id}/override":                  {"PATCH", "catalog.menu.manage", "catalog.menu.update", true},
	}
}

func assertCatalogBearer(t *testing.T, key string, operation map[string]any) {
	t.Helper()
	security, ok := operation["security"].([]any)
	if !ok || len(security) != 1 {
		t.Fatalf("Catalog operation %s security=%v", key, operation["security"])
	}
	scheme, ok := security[0].(map[string]any)
	if !ok || scheme["bearerAuth"] == nil {
		t.Fatalf("Catalog operation %s security=%v", key, security)
	}
}

func assertCatalogIdempotency(
	t *testing.T, key string, item, operation map[string]any,
) {
	t.Helper()
	for _, holder := range []map[string]any{item, operation} {
		parameters, _ := holder["parameters"].([]any)
		for _, raw := range parameters {
			parameter, _ := raw.(map[string]any)
			reference, _ := parameter["$ref"].(string)
			if strings.HasSuffix(reference, "/IdempotencyKeyParameter") {
				return
			}
		}
	}
	t.Fatalf("Catalog operation %s has no idempotency key", key)
}
