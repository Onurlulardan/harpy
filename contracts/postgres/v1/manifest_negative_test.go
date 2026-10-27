package postgresv1

import (
	"strings"
	"testing"
)

type mutationCase struct {
	name   string
	mutate func(map[string]any)
}

func TestMigrationManifestRejectsUnsafeVariants(t *testing.T) {
	tests := []mutationCase{
		{"path traversal", func(value map[string]any) {
			migrationVariant(value, "fresh")["path"] = "../escape.sql"
		}},
		{"missing variants", func(value map[string]any) {
			delete(migrationEntry(value), "variants")
		}},
		{"unknown transaction mode", func(value map[string]any) {
			migrationEntry(value)["transaction_mode"] = "best_effort"
		}},
		{"duplicate entry", func(value map[string]any) {
			entries := value["entries"].([]any)
			value["entries"] = append(entries, entries[0])
		}},
		{"uppercase hash", func(value map[string]any) {
			migrationVariant(value, "v1_adopted")["sha256"] = strings.Repeat("A", 64)
		}},
	}
	assertMutationsInvalid(t, "migration-manifest.schema.json", validMigrationManifest, tests)
}

func TestExtensionManifestRejectsUnsafeInventory(t *testing.T) {
	tests := []mutationCase{
		{"wrong image", func(value map[string]any) {
			value["postgres_image_digest"] = "sha256:" + digest("b")
		}},
		{"unknown extension", func(value map[string]any) {
			extensionEntry(value)["name"] = "plpython3u"
		}},
		{"unsafe control path", func(value map[string]any) {
			value["control_path_policy"].(map[string]any)["allow_absolute_paths"] = true
		}},
		{"duplicate capability", func(value map[string]any) {
			extensionEntry(value)["capabilities"] = []any{"vector_similarity", "vector_similarity"}
		}},
		{"wrong vector version", func(value map[string]any) {
			value["extensions"].([]any)[1].(map[string]any)["version"] = "0.8.2"
		}},
	}
	assertMutationsInvalid(t, "extension-manifest.schema.json", validExtensionManifest, tests)
}

func TestRolesManifestRejectsCatalogDrift(t *testing.T) {
	tests := []mutationCase{
		{"wrong order", func(value map[string]any) {
			roles := value["roles"].([]any)
			roles[0], roles[1] = roles[1], roles[0]
		}},
		{"extension admin not superuser", func(value map[string]any) {
			value["roles"].([]any)[6].(map[string]any)["superuser"] = false
		}},
		{"cluster admin can login", func(value map[string]any) {
			value["roles"].([]any)[5].(map[string]any)["login"] = true
		}},
		{"role has unknown attribute", func(value map[string]any) {
			value["roles"].([]any)[0].(map[string]any)["temporary"] = true
		}},
	}
	assertMutationsInvalid(t, "roles-manifest.schema.json", func() map[string]any {
		return validRolesManifest("fresh")
	}, tests)

	adopted := validRolesManifest("v1_adopted")
	adopted["roles"] = adopted["roles"].([]any)[:10]
	if err := compileArtifactSchema(t, "roles-manifest.schema.json").Validate(adopted); err == nil {
		t.Fatal("incomplete adopted role set accepted")
	}
}

func TestPrivilegesManifestRejectsAuthorityDrift(t *testing.T) {
	tests := []mutationCase{
		{"membership order", func(value map[string]any) {
			memberships := value["memberships"].([]any)
			memberships[0], memberships[1] = memberships[1], memberships[0]
		}},
		{"admin option", func(value map[string]any) {
			value["memberships"].([]any)[0].(map[string]any)["admin_option"] = true
		}},
		{"public temp", func(value map[string]any) {
			value["public_database_privileges"] = []any{"TEMP"}
		}},
		{"wrong executor", func(value map[string]any) {
			value["executor_decisions"].(map[string]any)["ownership_executor"] =
				"external_root_provisioner"
		}},
		{"unknown privilege", func(value map[string]any) {
			value["grants"].([]any)[0].(map[string]any)["privileges"] = []any{"SUPERUSER"}
		}},
		{"missing login connect", func(value map[string]any) {
			grants := value["grants"].([]any)
			value["grants"] = grants[:len(grants)-1]
		}},
	}
	assertMutationsInvalid(t, "privileges-manifest.schema.json", func() map[string]any {
		return validPrivilegesManifest("fresh")
	}, tests)
}

func migrationEntry(value map[string]any) map[string]any {
	return value["entries"].([]any)[0].(map[string]any)
}

func migrationVariant(value map[string]any, origin string) map[string]any {
	return migrationEntry(value)["variants"].(map[string]any)[origin].(map[string]any)
}

func extensionEntry(value map[string]any) map[string]any {
	return value["extensions"].([]any)[0].(map[string]any)
}

func assertMutationsInvalid(
	t *testing.T,
	schema string,
	fixture func() map[string]any,
	tests []mutationCase,
) {
	t.Helper()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := fixture()
			test.mutate(value)
			if err := compileArtifactSchema(t, schema).Validate(value); err == nil {
				t.Fatal("unsafe document accepted")
			}
		})
	}
}
