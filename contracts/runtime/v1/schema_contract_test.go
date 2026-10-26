package runtimev1

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestConfigCatalogSatisfiesItsSchema(t *testing.T) {
	schema := compileSchema(t, "config.schema.json")
	if err := schema.Validate(readJSONValue(t, "config.json")); err != nil {
		t.Fatalf("config catalog does not satisfy its schema: %v", err)
	}
}

func TestProfileSchemasAcceptMinimalExactSnapshots(t *testing.T) {
	tests := []struct {
		schema   string
		snapshot map[string]any
	}{
		{"core-api-config.schema.json", profileSnapshot("api")},
		{"core-worker-config.schema.json", profileSnapshot("worker")},
		{"core-migrator-config.schema.json", profileSnapshot("migrator_deploy")},
		{"core-migrator-config.schema.json", profileSnapshot("schema_executor")},
		{"core-migrator-config.schema.json", profileSnapshot("cluster_bootstrap")},
		{"core-migrator-config.schema.json", profileSnapshot("extension_admin")},
	}
	for _, test := range tests {
		t.Run(test.snapshot["profile"].(string), func(t *testing.T) {
			if err := compileSchema(t, test.schema).Validate(test.snapshot); err != nil {
				t.Fatalf("valid profile snapshot rejected: %v", err)
			}
		})
	}
}

func TestProfileSchemasRejectInvalidSnapshots(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		snapshot map[string]any
	}{
		{"both DSN sources", "core-api-config.schema.json", snapshotWith("api", map[string]any{"HARPY_POSTGRES_DSN_FILE": "C:/secret"})},
		{"unknown key", "core-worker-config.schema.json", snapshotWith("worker", map[string]any{"HARPY_UNKNOWN": "value"})},
		{"production without OTLP", "core-api-config.schema.json", snapshotWith("api", map[string]any{
			"HARPY_ENVIRONMENT":  "production",
			"HARPY_POSTGRES_DSN": "postgresql://user@example.invalid/harpy?sslmode=verify-full",
		})},
		{"worker public listener", "core-worker-config.schema.json", snapshotWith("worker", map[string]any{"HARPY_PUBLIC_LISTEN_ADDR": "127.0.0.1:8000"})},
		{"deploy management listener", "core-migrator-config.schema.json", snapshotWith("migrator_deploy", map[string]any{"HARPY_MANAGEMENT_LISTEN_ADDR": "127.0.0.1:9000"})},
		{"headers without endpoint", "core-api-config.schema.json", snapshotWith("api", map[string]any{"HARPY_OTLP_HEADERS": "authorization=value"})},
		{"bootstrap runtime DSN", "core-migrator-config.schema.json", snapshotWith("cluster_bootstrap", map[string]any{
			"HARPY_POSTGRES_DSN": "postgresql://user@example.invalid/harpy?sslmode=disable",
		})},
		{"extension bootstrap receipt", "core-migrator-config.schema.json", snapshotWith("extension_admin", map[string]any{
			"HARPY_BOOTSTRAP_RECEIPT_PATH": "/receipt.json",
		})},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := compileSchema(t, test.schema).Validate(test.snapshot); err == nil {
				t.Fatal("invalid profile snapshot was accepted")
			}
		})
	}
}

func compileSchema(t *testing.T, path string) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	entries, err := filepath.Glob("*.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	const base = "https://harpy.io/contracts/runtime/v1/"
	for _, entry := range entries {
		if err := compiler.AddResource(base+entry, readJSONValue(t, entry)); err != nil {
			t.Fatalf("register schema %s: %v", entry, err)
		}
	}
	schema, err := compiler.Compile(base + path)
	if err != nil {
		t.Fatalf("compile schema %s: %v", path, err)
	}
	return schema
}

func readJSONValue(t *testing.T, path string) any {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	value, err := jsonschema.UnmarshalJSON(file)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func profileSnapshot(profile string) map[string]any {
	variables := map[string]any{
		"HARPY_ENVIRONMENT":            "development",
		"HARPY_POSTGRES_ENDPOINT_KIND": "direct_postgres",
	}
	switch profile {
	case "cluster_bootstrap":
		variables["HARPY_BOOTSTRAP_MAINTENANCE_DSN"] = "postgresql://harpy_cluster_admin@example.invalid/postgres?sslmode=disable"
		variables["HARPY_BOOTSTRAP_TARGET_DSN"] = "postgresql://harpy_cluster_admin@example.invalid/harpy?sslmode=disable"
		variables["HARPY_BOOTSTRAP_RECEIPT_PATH"] = "/receipts/bootstrap.json"
		addReleaseSnapshot(variables)
		addBootstrapArtifactsSnapshot(variables)
	case "extension_admin":
		variables["HARPY_EXTENSION_ADMIN_POSTGRES_DSN"] = "postgresql://harpy_extension_admin@example.invalid/harpy?sslmode=disable"
		variables["HARPY_EXTENSION_RECEIPT_PATH"] = "/receipts/extension.json"
		addReleaseSnapshot(variables)
		addExtensionArtifactsSnapshot(variables)
	case "migrator_deploy":
		variables["HARPY_POSTGRES_DSN"] = "postgresql://user@example.invalid/harpy?sslmode=disable"
		addReleaseVerificationSnapshot(variables)
	default:
		variables["HARPY_POSTGRES_DSN"] = "postgresql://user@example.invalid/harpy?sslmode=disable"
	}
	if profile == "api" {
		variables["HARPY_REST_QUERY_CURSOR_KEYRING"] = validCursorKeyringContractValue()
		variables["HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING"] =
			validCatalogPackageKeyringContractValue()
		addFormsExportProfileSnapshot(variables, false)
		addAnalyticsDashboardExportProfileSnapshot(variables, false)
		addFormsImportSourceProfileSnapshot(variables, true)
		addIAMSystemTransferProfileSnapshot(variables)
		addIAMAPIProfileSnapshot(variables)
	}
	if profile == "worker" {
		variables["GOMEMLIMIT"] = "512MiB"
		variables["HARPY_RABBITMQ_URL"] = "amqp://worker:secret@rabbit.invalid/%2f"
		variables["HARPY_REST_QUERY_CURSOR_KEYRING"] = validCursorKeyringContractValue()
		variables["HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING"] =
			validCatalogPackageKeyringContractValue()
		addFormsExportProfileSnapshot(variables, true)
		addAnalyticsDashboardExportProfileSnapshot(variables, true)
		addFormsImportSourceProfileSnapshot(variables, false)
		addIAMSystemTransferProfileSnapshot(variables)
		addIAMWorkerProfileSnapshot(variables)
	}
	addPhase14RuntimeSnapshot(profile, variables)
	return map[string]any{"profile": profile, "variables": variables}
}

func addReleaseInputsSnapshot(variables map[string]any) {
	variables["HARPY_RELEASE_MANIFEST_PATH"] = "/release/manifest.json"
	variables["HARPY_RELEASE_SIGNATURE_PATH"] = "/release/manifest.sig"
	variables["HARPY_RELEASE_TRUST_ROOT_PATH"] = "/release/trust.pem"
}

func addReleaseSnapshot(variables map[string]any) {
	addReleaseInputsSnapshot(variables)
	addAuthorityArtifactsSnapshot(variables)
	variables["HARPY_CREDENTIAL_LEASE_ID"] = "55555555-5555-4555-8555-555555555555"
	variables["HARPY_ATTESTATION_SIGNING_KEY"] = "test-signing-key"
}

func addAuthorityArtifactsSnapshot(variables map[string]any) {
	variables["HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH"] = "/authority/manifest.json"
	variables["HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH"] = "/authority/manifest.sig"
}

func addBootstrapArtifactsSnapshot(variables map[string]any) {
	variables["HARPY_ROLES_MANIFEST_PATH"] = "/postgres/roles.json"
	variables["HARPY_PRIVILEGES_MANIFEST_PATH"] = "/postgres/privileges.json"
	variables["HARPY_OWNERSHIP_MANIFEST_PATH"] = "/postgres/ownership.json"
	variables["HARPY_BOOTSTRAP_OPERATION_PATH"] = "/postgres/bootstrap-operation.json"
}

func addExtensionArtifactsSnapshot(variables map[string]any) {
	variables["HARPY_EXTENSION_MANIFEST_PATH"] = "/postgres/extensions.json"
	variables["HARPY_OWNERSHIP_MANIFEST_PATH"] = "/postgres/ownership.json"
	variables["HARPY_EXTENSION_OPERATION_PATH"] = "/postgres/extension-operation.json"
}

func addMigrationArtifactsSnapshot(variables map[string]any) {
	addBootstrapArtifactsSnapshot(variables)
	addExtensionArtifactsSnapshot(variables)
	variables["HARPY_MIGRATION_MANIFEST_PATH"] = "/postgres/migrations.json"
	variables["HARPY_CREDENTIAL_CLEANUP_OPERATION_PATH"] = "/postgres/cleanup-operation.json"
}
