package iam

import (
	"strings"
	"testing"
)

func TestPermissionRequestCatalogMigrationIsSelfServiceAndFailClosed(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	wanted := map[string][]string{
		"20260801011000_create_iam_permission_request_resource_catalog_v1": {
			"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
			"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
			"p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()",
			"actor.status OPERATOR(pg_catalog.=) 'active'",
			"permission.match_kind OPERATOR(pg_catalog.=) 'exact'",
			"NOT permission.is_protected",
			"NOT resource.is_protected",
			"NOT action.is_protected",
			"NOT iam_user_has_permission_v1(p_actor_user_id, permission.id)",
		},
		"20260801011001_create_iam_permission_request_action_catalog_v1": {
			"actor.status OPERATOR(pg_catalog.=) 'active'",
			"permission.match_kind OPERATOR(pg_catalog.=) 'exact'",
			"NOT permission.is_protected",
			"NOT resource.is_protected",
			"NOT action.is_protected",
		},
	}
	found := map[string]bool{}
	for _, document := range documents {
		for logical, markers := range wanted {
			if !strings.Contains(document.Path(), logical+".fresh.sql") {
				continue
			}
			content := string(document.Content())
			for _, marker := range markers {
				if !strings.Contains(content, marker) {
					t.Fatalf("%s misses security marker %q", logical, marker)
				}
			}
			if strings.Contains(content, "iam_rbac_read_assert_v1") {
				t.Fatalf("%s delegates to the admin catalog authority", logical)
			}
			found[logical] = true
		}
	}
	for logical := range wanted {
		if !found[logical] {
			t.Fatalf("catalog migration not embedded: %s", logical)
		}
	}
}
