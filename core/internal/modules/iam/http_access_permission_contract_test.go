package iam

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTTPPermissionAuthorityPairsAreAcceptedByIAMBridges(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(
		"..", "..", "..", "..", "contracts", "http", "*.paths.json",
	))
	if err != nil || len(files) == 0 {
		t.Fatalf("HTTP path contracts are unavailable: %v", err)
	}
	coverage := map[string]int{}
	for _, path := range files {
		document := readHTTPPermissionContract(t, path)
		walkHTTPPermissionContract(document, func(operation map[string]any) {
			authority, authorityFound := operation["x-harpy-authority"].(string)
			if !authorityFound {
				return
			}
			for _, permission := range httpContractPermissions(
				operation["x-harpy-permission"],
			) {
				bridge, handled, valid := resolveHTTPPermissionAuthority(permission, authority)
				if !handled {
					continue
				}
				coverage[bridge]++
				if !valid {
					t.Errorf("%s rejects contract pair %s/%s from %s",
						bridge, permission, authority, filepath.Base(path))
				}
			}
		})
	}
	for _, bridge := range []string{
		"analytics", "catalog", "documents",
	} {
		if coverage[bridge] == 0 {
			t.Errorf("%s HTTP bridge has no contract coverage", bridge)
		}
	}
}

func readHTTPPermissionContract(t *testing.T, path string) any {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var document any
	if err = json.Unmarshal(content, &document); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return document
}

func walkHTTPPermissionContract(value any, visit func(map[string]any)) {
	switch typed := value.(type) {
	case map[string]any:
		visit(typed)
		for _, child := range typed {
			walkHTTPPermissionContract(child, visit)
		}
	case []any:
		for _, child := range typed {
			walkHTTPPermissionContract(child, visit)
		}
	}
}

func httpContractPermissions(value any) []string {
	if permission, valid := value.(string); valid {
		return []string{permission}
	}
	values, valid := value.([]any)
	if !valid {
		return nil
	}
	permissions := make([]string, 0, len(values))
	for _, value := range values {
		if permission, valid := value.(string); valid {
			permissions = append(permissions, permission)
		}
	}
	return permissions
}

func resolveHTTPPermissionAuthority(permission, authority string) (
	bridge string, handled, valid bool,
) {
	switch {
	case strings.HasPrefix(permission, "analytics.query."),
		strings.HasPrefix(permission, "analytics.dashboard."),
		strings.HasPrefix(permission, "analytics.publication."):
		_, valid = resolveAnalyticsAccess(permission, authority)
		return "analytics", true, valid
	case strings.HasPrefix(permission, "catalog."):
		_, valid = resolveCatalogAccess(permission, authority)
		return "catalog", true, valid
	case strings.HasPrefix(permission, "documents."):
		_, valid = resolveDocumentsAccess(permission, authority)
		return "documents", true, valid
	default:
		return "", false, false
	}
}
