package runtimev1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAPIContractsAreStrictJSONAndResolveLocalReferences(t *testing.T) {
	for _, name := range []string{"core-api-v1.openapi.yaml", "management-v1.openapi.yaml"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "..", "http", name)
			document := readOpenAPIDocument(t, path)
			if document["openapi"] != "3.1.1" {
				t.Fatalf("unexpected OpenAPI version: %v", document["openapi"])
			}
			assertLocalReferences(t, filepath.Dir(path), document)
		})
	}
}

func TestCoreAndManagementRoutesAreExact(t *testing.T) {
	corePath := filepath.Join("..", "..", "http", "core-api-v1.openapi.yaml")
	core := readOpenAPIDocument(t, corePath)
	corePaths := core["paths"].(map[string]any)
	for path := range coreFragmentPaths(t) {
		if _, found := corePaths[path]; !found {
			t.Fatalf("core API is missing fragment path %s", path)
		}
	}
	jwks, found := corePaths["/.well-known/jwks.json"].(map[string]any)
	if !found || jwks["get"] == nil || jwks["head"] == nil || len(jwks) != 2 {
		t.Fatal("core API does not publish exact JWKS GET/HEAD operations")
	}
	authPaths := map[string]string{
		"/api/auth/login":        "operatorLogin",
		"/api/auth/portal/login": "portalLogin",
		"/api/auth/refresh":      "refresh",
		"/api/auth/logout":       "logout",
		"/api/auth/csrf":         "csrf",
		"/api/auth/session":      "currentSession",
	}
	for route, pathItem := range authPaths {
		entry, found := corePaths[route].(map[string]any)
		want := "iam-auth-v1.paths.json#/pathItems/" + pathItem
		if !found || len(entry) != 1 || entry["$ref"] != want {
			t.Fatalf("core API auth path %s is not exact: %v", route, entry)
		}
	}
	systemPaths := map[string]string{
		"/api/users":                                           "iam-users-v1.paths.json#/pathItems/collection",
		"/api/users/query":                                     "iam-users-v1.paths.json#/pathItems/query",
		"/api/users/{id}":                                      "iam-users-v1.paths.json#/pathItems/item",
		"/api/users/{id}/state":                                "iam-users-v1.paths.json#/pathItems/state",
		"/api/roles":                                           "iam-roles-v1.paths.json#/pathItems/collection",
		"/api/roles/query":                                     "iam-roles-v1.paths.json#/pathItems/query",
		"/api/roles/{id}":                                      "iam-roles-v1.paths.json#/pathItems/item",
		"/api/roles/{id}/state":                                "iam-roles-v1.paths.json#/pathItems/state",
		"/api/organizations":                                   "iam-organizations-v1.paths.json#/pathItems/collection",
		"/api/organizations/query":                             "iam-organizations-v1.paths.json#/pathItems/query",
		"/api/organizations/{id}":                              "iam-organizations-v1.paths.json#/pathItems/item",
		"/api/organizations/{id}/state":                        "iam-organizations-v1.paths.json#/pathItems/state",
		"/api/permission-requests":                             "iam-permission-requests-v1.paths.json#/pathItems/collection",
		"/api/permission-requests/mine":                        "iam-permission-requests-v1.paths.json#/pathItems/mine",
		"/api/permission-requests/catalog/resources":           "iam-permission-requests-v1.paths.json#/pathItems/catalogResources",
		"/api/permission-requests/catalog/actions":             "iam-permission-requests-v1.paths.json#/pathItems/catalogActions",
		"/api/permission-requests/catalog/missing-resources":   "iam-permission-requests-v1.paths.json#/pathItems/catalogMissingResources",
		"/api/permission-requests/{id}":                        "iam-permission-requests-v1.paths.json#/pathItems/item",
		"/api/permission-requests/{id}/review":                 "iam-permission-requests-v1.paths.json#/pathItems/review",
		"/api/permission-requests/{id}/approve":                "iam-permission-requests-v1.paths.json#/pathItems/approve",
		"/api/permission-requests/{id}/reject":                 "iam-permission-requests-v1.paths.json#/pathItems/reject",
		"/api/permission-requests/{id}/cancel":                 "iam-permission-requests-v1.paths.json#/pathItems/cancel",
		"/api/permission-requests/{id}/expire":                 "iam-permission-requests-v1.paths.json#/pathItems/expire",
		"/api/invitations":                                     "iam-invitations-v1.paths.json#/pathItems/collection",
		"/api/invitations/{id}":                                "iam-invitations-v1.paths.json#/pathItems/item",
		"/api/invitations/verify":                              "iam-invitations-v1.paths.json#/pathItems/verify",
		"/api/invitations/accept":                              "iam-invitations-v1.paths.json#/pathItems/accept",
		"/api/permissions":                                     "iam-rbac-v1.paths.json#/pathItems/collection",
		"/api/permissions/{id}":                                "iam-rbac-v1.paths.json#/pathItems/item",
		"/api/permissions/resources":                           "iam-rbac-v1.paths.json#/pathItems/resourceCollection",
		"/api/permissions/resources/{id}":                      "iam-rbac-v1.paths.json#/pathItems/resourceItem",
		"/api/permissions/actions":                             "iam-rbac-v1.paths.json#/pathItems/actionCollection",
		"/api/permissions/actions/{id}":                        "iam-rbac-v1.paths.json#/pathItems/actionItem",
		"/api/permissions/check":                               "iam-rbac-v1.paths.json#/pathItems/check",
		"/api/permissions/batch-check":                         "iam-rbac-v1.paths.json#/pathItems/batchCheck",
		"/api/users/{id}/permissions":                          "iam-rbac-v1.paths.json#/pathItems/userEffective",
		"/api/roles/{id}/permissions":                          "iam-rbac-v1.paths.json#/pathItems/roleEffective",
		"/api/organizations/{id}/permissions":                  "iam-rbac-v1.paths.json#/pathItems/organizationEffective",
		"/api/permissions/cache/stats":                         "iam-rbac-v1.paths.json#/pathItems/cacheStats",
		"/api/permissions/cache/invalidate/{user_id}":          "iam-rbac-v1.paths.json#/pathItems/cacheInvalidateUser",
		"/api/permissions/cache/invalidate/role/{role_id}":     "iam-rbac-v1.paths.json#/pathItems/cacheInvalidateRole",
		"/api/permissions/cache/invalidate/org/{org_id}":       "iam-rbac-v1.paths.json#/pathItems/cacheInvalidateOrganization",
		"/api/permissions/cache/invalidate/tenant/{tenant_id}": "iam-rbac-v1.paths.json#/pathItems/cacheInvalidateTenant",
		"/api/permissions/cache/invalidate/all":                "iam-rbac-v1.paths.json#/pathItems/cacheInvalidateAll",
	}
	for route, reference := range systemPaths {
		entry, found := corePaths[route].(map[string]any)
		if !found || len(entry) != 1 || entry["$ref"] != reference {
			t.Fatalf("core API system-resource path %s is not exact: %v", route, entry)
		}
	}
	profilePaths := map[string]string{
		"/api/profile": "iam-profile-v1.paths.json#/pathItems/self",
		"/api/iam/profile-schemas/{resource}/active":                      "iam-profile-v1.paths.json#/pathItems/active",
		"/api/iam/profile-schemas/{resource}/{schema_id}":                 "iam-profile-v1.paths.json#/pathItems/historical",
		"/api/iam/profile-schemas/{resource}/publish":                     "iam-profile-v1.paths.json#/pathItems/publish",
		"/api/iam/profile-schemas/{resource}/backfills/{plan_id}/recover": "iam-profile-v1.paths.json#/pathItems/recovery",
		"/api/iam/profile-schemas/{resource}/fields/{field_id}/retire":    "iam-profile-v1.paths.json#/pathItems/retire",
	}
	for route, reference := range profilePaths {
		entry, found := corePaths[route].(map[string]any)
		if !found || len(entry) != 1 || entry["$ref"] != reference {
			t.Fatalf("core API profile path %s is not exact: %v", route, entry)
		}
	}
	managementPath := filepath.Join("..", "..", "http", "management-v1.openapi.yaml")
	management := readOpenAPIDocument(t, managementPath)
	paths := management["paths"].(map[string]any)
	if len(paths) != 3 {
		t.Fatalf("unexpected management path count: %d", len(paths))
	}
	for _, path := range []string{"/livez", "/readyz", "/version"} {
		operations, found := paths[path].(map[string]any)
		if !found || operations["get"] == nil || operations["head"] == nil || len(operations) != 2 {
			t.Fatalf("management path %s does not expose exact GET/HEAD", path)
		}
	}
}

func readOpenAPIDocument(t *testing.T, path string) map[string]any {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(content, &document); err != nil {
		t.Fatalf("OpenAPI contract must remain strict JSON: %v", err)
	}
	return document
}

func assertLocalReferences(t *testing.T, directory string, value any) {
	t.Helper()
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			if key == "$ref" {
				reference := child.(string)
				if !strings.HasPrefix(reference, "#") {
					path := strings.SplitN(reference, "#", 2)[0]
					base := filepath.Base(filepath.FromSlash(path))
					if strings.HasPrefix(base, "forms-") || strings.HasPrefix(base, "workflow-") {
						continue
					}
					if _, err := os.Stat(filepath.Join(directory, filepath.FromSlash(path))); err != nil {
						t.Fatalf("unresolved OpenAPI reference %s: %v", reference, err)
					}
				}
			}
			assertLocalReferences(t, directory, child)
		}
	case []any:
		for _, child := range current {
			assertLocalReferences(t, directory, child)
		}
	}
}
