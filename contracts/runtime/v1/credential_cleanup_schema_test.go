package runtimev1

import (
	"strings"
	"testing"
)

func TestCredentialCleanupReceiptAcceptsRevokedPrincipals(t *testing.T) {
	digest := strings.Repeat("a", 64)
	if err := compileSchema(t, "credential-cleanup-receipt.schema.json").Validate(
		validCredentialCleanupReceipt(digest),
	); err != nil {
		t.Fatalf("valid cleanup receipt rejected: %v", err)
	}
}

func TestCredentialCleanupReceiptRejectsResidualAuthority(t *testing.T) {
	digest := strings.Repeat("a", 64)
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"login remains enabled", func(value map[string]any) { cleanedPrincipal(value, 0)["can_login"] = true }},
		{"password remains", func(value map[string]any) { cleanedPrincipal(value, 1)["password_state"] = "present" }},
		{"string null is not null", func(value map[string]any) { cleanedPrincipal(value, 1)["password_state"] = "null" }},
		{"active session remains", func(value map[string]any) { cleanedPrincipal(value, 0)["active_sessions"] = 1 }},
		{"membership remains", func(value map[string]any) { cleanedPrincipal(value, 0)["membership_count"] = 1 }},
		{"wrong owner", func(value map[string]any) { cleanupPayload(value)["database_owner"] = "postgres" }},
		{"wrong locale provider", func(value map[string]any) { cleanupPayload(value)["database_locale_provider"] = "icu" }},
		{"checksum disabled", func(value map[string]any) { cleanupPayload(value)["data_checksums"] = false }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := validCredentialCleanupReceipt(digest)
			test.mutate(value)
			if err := compileSchema(t, "credential-cleanup-receipt.schema.json").Validate(value); err == nil {
				t.Fatal("unsafe credential cleanup accepted")
			}
		})
	}
}

func validCredentialCleanupReceipt(digest string) map[string]any {
	return map[string]any{
		"contract_version": "credential-cleanup-receipt.v1", "kind": "credential_cleanup",
		"payload": map[string]any{
			"run_id":     "11111111-1111-4111-8111-111111111111",
			"release_id": "44444444-4444-4444-8444-444444444444",
			"issued_at":  "2026-07-13T00:20:00Z", "release_manifest_sha256": digest,
			"bootstrap_authority_manifest_sha256": digest, "cleanup_operation_digest": digest,
			"database_name": "harpy", "database_origin": "fresh",
			"database_system_identifier": "123456789", "database_oid": 16384,
			"database_owner": "harpy_owner", "database_encoding": "UTF8",
			"database_locale_provider": "libc", "database_collation": "en_US.utf8",
			"database_ctype": "en_US.utf8", "data_checksums": true, "principals": []any{
				cleanedPrincipalFixture("harpy_cluster_admin", "55555555-5555-4555-8555-555555555555"),
				cleanedPrincipalFixture("harpy_extension_admin", "66666666-6666-4666-8666-666666666666"),
			},
			"result": "succeeded",
		},
		"signature": validReceiptSignature(),
	}
}

func cleanedPrincipalFixture(name, leaseID string) map[string]any {
	return map[string]any{
		"name": name, "lease_id": leaseID, "cleaned_at": "2026-07-13T00:19:00Z",
		"can_login": false, "password_state": nil, "active_sessions": 0,
		"membership_count": 0,
	}
}

func cleanupPayload(value map[string]any) map[string]any {
	return value["payload"].(map[string]any)
}

func cleanedPrincipal(value map[string]any, index int) map[string]any {
	return cleanupPayload(value)["principals"].([]any)[index].(map[string]any)
}
