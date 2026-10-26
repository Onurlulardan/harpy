package cutoverv1

import (
	"strings"
	"testing"
)

func TestCredentialLeaseSchemasAcceptExactAuthority(t *testing.T) {
	digest := strings.Repeat("a", 64)
	lease := validCredentialLease(digest)
	if err := compileCutoverSchema(t, "credential-lease.schema.json").Validate(lease); err != nil {
		t.Fatalf("valid credential lease rejected: %v", err)
	}
	signature := map[string]any{
		"contract_version": "cutover-credential-lease-signature.v1",
		"manifest_sha256":  digest,
		"signature": map[string]any{
			"algorithm": "Ed25519", "key_id": "cutover-2026",
			"value": strings.Repeat("A", 86),
		},
	}
	if err := compileCutoverSchema(t, "credential-lease-signature.schema.json").Validate(signature); err != nil {
		t.Fatalf("valid lease signature rejected: %v", err)
	}
}

func TestCredentialLeaseRejectsPrivilegeAndPlanDrift(t *testing.T) {
	digest := strings.Repeat("a", 64)
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"decoder replication disabled", func(value map[string]any) {
			roleLease(value, 0)["replication"] = false
		}},
		{"adopter replication enabled", func(value map[string]any) {
			roleLease(value, 1)["replication"] = true
		}},
		{"unbounded source sessions", func(value map[string]any) {
			value["source_connection_limit"] = 33
		}},
		{"missing relation authority", func(value map[string]any) {
			value["relations"] = []any{}
		}},
		{"unknown authority", func(value map[string]any) { value["raw_sql"] = "SELECT 1" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := validCredentialLease(digest)
			test.mutate(value)
			if err := compileCutoverSchema(t, "credential-lease.schema.json").Validate(value); err == nil {
				t.Fatal("unsafe credential lease accepted")
			}
		})
	}
}

func validCredentialLease(digest string) map[string]any {
	return map[string]any{
		"contract_version": "cutover-credential-lease.v1",
		"lease_id":         "11111111-1111-4111-8111-111111111111",
		"release_id":       "22222222-2222-4222-8222-222222222222",
		"environment_id":   "33333333-3333-4333-8333-333333333333",
		"operation_id":     "44444444-4444-4444-8444-444444444444",
		"run_id":           "55555555-5555-4555-8555-555555555555",
		"issued_at":        "2026-07-13T00:00:00Z", "not_before": "2026-07-13T00:00:00Z",
		"expires_at": "2026-07-13T00:15:00Z", "max_duration_seconds": 900,
		"scope_kind": "tenant", "scope_id": "66666666-6666-4666-8666-666666666666",
		"group_id": "iam_system_resources", "domains": []any{"iam", "system_resource_layout"},
		"plan_sha256": digest, "transform_manifest_sha256": digest,
		"source": cutoverDatabaseIdentity(), "target": cutoverDatabaseIdentity(),
		"publication_name": "cutover_iam", "slot_name": "cutover_iam_slot",
		"source_connection_limit": 2, "target_connection_limit": 1,
		"role_leases": []any{
			map[string]any{"principal": "harpy_cutover_decoder", "login": true, "replication": true, "connection_limit": 2, "password_state": "leased_secret"},
			map[string]any{"principal": "harpy_adopter", "login": true, "replication": false, "connection_limit": 1, "password_state": "leased_secret"},
		},
		"relations": []any{cutoverRelationMapping()},
	}
}

func cutoverDatabaseIdentity() map[string]any {
	return map[string]any{
		"database_name": "harpy", "database_origin": "v1_adopted",
		"system_identifier": "123456789", "database_oid": 16384,
	}
}

func cutoverRelationMapping() map[string]any {
	identity := func(oid int, name string) map[string]any {
		return map[string]any{"schema_oid": 2200, "relation_oid": oid, "schema_name": "platform", "relation_name": name}
	}
	return map[string]any{
		"ordinal": 1, "source": identity(17000, "iam_source"), "target": identity(18000, "iam_target"),
		"replica_identity": "primary_key", "procedure_contract_id": "platform.cutover_probe.v1",
		"apply_procedure_oid":       19000,
		"apply_procedure_signature": "tenant_a.apply_iam_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)",
		"expected_checkpoint_keys":  []any{"main"},
		"columns": []any{map[string]any{
			"ordinal": 1, "source_attnum": 1, "target_attnum": 1,
			"transform_id": "identity.uuid", "primary_key": true,
		}},
	}
}

func roleLease(value map[string]any, index int) map[string]any {
	return value["role_leases"].([]any)[index].(map[string]any)
}
