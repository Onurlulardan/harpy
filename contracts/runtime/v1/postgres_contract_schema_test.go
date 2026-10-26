package runtimev1

import (
	"strings"
	"testing"
)

func TestPostgresReceiptsAndSchemaContractsValidate(t *testing.T) {
	digest := strings.Repeat("a", 64)
	tests := []struct {
		schema string
		value  map[string]any
	}{
		{"bootstrap-receipt.schema.json", validBootstrapReceipt(digest)},
		{"extension-receipt.schema.json", validExtensionReceipt(digest)},
		{"binary-schema-contract.schema.json", map[string]any{
			"contract_version": "database-schema-contract.v1", "component": "core",
			"runtime_generation": 1, "minimum_database_generation": 1,
			"migration_manifest_sha256": digest, "extension_manifest_sha256": digest,
			"postgres_image_digest": "sha256:c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84",
		}},
		{"schema-contract.schema.json", map[string]any{
			"component": "core", "database_generation": 1, "minimum_runtime_generation": 1,
			"migration_manifest_sha256": digest, "extension_manifest_sha256": digest,
			"status": "ready", "verified_run_id": "33333333-3333-4333-8333-333333333333",
			"revision": 1, "verified_at": "2026-07-13T00:00:00Z",
		}},
	}
	for _, test := range tests {
		t.Run(test.schema, func(t *testing.T) {
			if err := compileSchema(t, test.schema).Validate(test.value); err != nil {
				t.Fatalf("valid contract rejected: %v", err)
			}
		})
	}
}

func TestPostgresReceiptsRejectAttestationMismatch(t *testing.T) {
	digest := strings.Repeat("a", 64)
	tests := []struct {
		name   string
		schema string
		value  map[string]any
		mutate func(map[string]any)
	}{
		{"missing kind", "bootstrap-receipt.schema.json", validBootstrapReceipt(digest), func(value map[string]any) {
			delete(value, "kind")
		}},
		{"missing environment id", "bootstrap-receipt.schema.json", validBootstrapReceipt(digest), func(value map[string]any) {
			delete(payloadOf(value), "environment_id")
		}},
		{"wrong image", "bootstrap-receipt.schema.json", validBootstrapReceipt(digest), func(value map[string]any) {
			payloadOf(value)["postgres_image_digest"] = "sha256:" + strings.Repeat("b", 64)
		}},
		{"wrong locale", "bootstrap-receipt.schema.json", validBootstrapReceipt(digest), func(value map[string]any) {
			payloadOf(value)["database_collation"] = "C.UTF-8"
		}},
		{"wrong locale provider", "bootstrap-receipt.schema.json", validBootstrapReceipt(digest), func(value map[string]any) {
			payloadOf(value)["database_locale_provider"] = "icu"
		}},
		{"checksum disabled", "bootstrap-receipt.schema.json", validBootstrapReceipt(digest), func(value map[string]any) {
			payloadOf(value)["data_checksums"] = false
		}},
		{"missing release id", "bootstrap-receipt.schema.json", validBootstrapReceipt(digest), func(value map[string]any) {
			delete(payloadOf(value), "release_id")
		}},
		{"missing ownership manifest", "bootstrap-receipt.schema.json", validBootstrapReceipt(digest), func(value map[string]any) {
			delete(payloadOf(value), "ownership_manifest_sha256")
		}},
		{"wrong bootstrap role", "bootstrap-receipt.schema.json", validBootstrapReceipt(digest), func(value map[string]any) {
			payloadOf(value)["session_user"] = "harpy_migrator"
		}},
		{"zero database oid", "bootstrap-receipt.schema.json", validBootstrapReceipt(digest), func(value map[string]any) {
			payloadOf(value)["database_oid"] = 0
		}},
		{"wrong kind", "extension-receipt.schema.json", validExtensionReceipt(digest), func(value map[string]any) {
			value["kind"] = "cluster_bootstrap"
		}},
		{"wrong extension role", "extension-receipt.schema.json", validExtensionReceipt(digest), func(value map[string]any) {
			payloadOf(value)["current_user"] = "harpy_owner"
		}},
		{"wrong extension owner", "extension-receipt.schema.json", validExtensionReceipt(digest), func(value map[string]any) {
			payloadOf(value)["extensions"].([]any)[0].(map[string]any)["owner"] = "harpy_owner"
		}},
		{"missing operation digest", "extension-receipt.schema.json", validExtensionReceipt(digest), func(value map[string]any) {
			delete(payloadOf(value), "operation_digest")
		}},
		{"failed result", "extension-receipt.schema.json", validExtensionReceipt(digest), func(value map[string]any) {
			payloadOf(value)["result"] = "failed"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mutate(test.value)
			if err := compileSchema(t, test.schema).Validate(test.value); err == nil {
				t.Fatal("unsafe privileged receipt accepted")
			}
		})
	}
}

func TestPostgresContractsRejectUnknownOrUnsafeValues(t *testing.T) {
	invalid := map[string]any{
		"contract_version": "database-schema-contract.v1", "component": "core",
		"runtime_generation": 0, "minimum_database_generation": 1, "unknown": true,
	}
	if err := compileSchema(t, "binary-schema-contract.schema.json").Validate(invalid); err == nil {
		t.Fatal("unsafe binary schema contract accepted")
	}
}

func validBootstrapReceipt(digest string) map[string]any {
	return map[string]any{
		"contract_version": "postgres-bootstrap-receipt.v1", "kind": "cluster_bootstrap",
		"payload": map[string]any{
			"run_id":                  "11111111-1111-4111-8111-111111111111",
			"release_id":              "44444444-4444-4444-8444-444444444444",
			"environment_id":          "88888888-8888-4888-8888-888888888888",
			"issued_at":               "2026-07-13T00:00:00Z",
			"release_manifest_sha256": digest, "postgres_image_digest": PostgresImageDigest,
			"bootstrap_authority_manifest_sha256": digest,
			"credential_lease_id":                 "55555555-5555-4555-8555-555555555555",
			"ownership_manifest_sha256":           digest,
			"database_name":                       "harpy",
			"database_system_identifier":          "123456789", "database_oid": 16384,
			"server_version_num": 180004, "session_user": "harpy_cluster_admin",
			"database_origin": "fresh", "database_owner": "harpy_owner",
			"database_encoding": "UTF8", "database_locale_provider": "libc",
			"database_collation": "en_US.utf8", "data_checksums": true,
			"database_ctype": "en_US.utf8",
			"current_user":   "harpy_cluster_admin", "operation_digest": digest,
			"result": "succeeded", "roles_manifest_sha256": digest,
			"privileges_manifest_sha256": digest,
		},
		"signature": validReceiptSignature(),
	}
}

func validExtensionReceipt(digest string) map[string]any {
	return map[string]any{
		"contract_version": "postgres-extension-receipt.v1", "kind": "extension_admin",
		"payload": map[string]any{
			"run_id":                  "22222222-2222-4222-8222-222222222222",
			"release_id":              "44444444-4444-4444-8444-444444444444",
			"issued_at":               "2026-07-13T00:00:00Z",
			"release_manifest_sha256": digest, "postgres_image_digest": PostgresImageDigest,
			"bootstrap_authority_manifest_sha256": digest,
			"credential_lease_id":                 "66666666-6666-4666-8666-666666666666",
			"ownership_manifest_sha256":           digest,
			"database_name":                       "harpy",
			"database_system_identifier":          "123456789", "database_oid": 16384,
			"server_version_num": 180004, "session_user": "harpy_extension_admin",
			"database_origin": "fresh", "database_owner": "harpy_owner",
			"database_encoding": "UTF8", "database_locale_provider": "libc",
			"database_collation": "en_US.utf8", "data_checksums": true,
			"database_ctype": "en_US.utf8",
			"current_user":   "harpy_extension_admin", "operation_digest": digest,
			"result": "succeeded", "extension_manifest_sha256": digest,
			"extensions": []any{map[string]any{
				"name": "vector", "version": "0.8.3", "schema": "platform_ext",
				"owner":               "harpy_extension_admin",
				"control_file_sha256": digest,
			}},
		},
		"signature": validReceiptSignature(),
	}
}

func validReceiptSignature() map[string]any {
	return map[string]any{
		"algorithm": "Ed25519", "key_id": "release-1", "value": strings.Repeat("A", 86),
	}
}

func payloadOf(receipt map[string]any) map[string]any {
	return receipt["payload"].(map[string]any)
}
