package runtimev1

import (
	"strings"
	"testing"
)

func TestBootstrapAuthoritySchemasAcceptExactContracts(t *testing.T) {
	digest := strings.Repeat("a", 64)
	retained := validAuthorityManifest("v1_adopted", digest)
	authorityDatabase(retained)["extension_owner"] = "oracle"
	authorityDatabase(retained)["extension_owner_mode"] = "adopted_retained_inventory"
	tests := []struct {
		name   string
		schema string
		value  map[string]any
	}{
		{"fresh authority", "bootstrap-authority-manifest.schema.json", validAuthorityManifest("fresh", digest)},
		{"adopted authority", "bootstrap-authority-manifest.schema.json", validAuthorityManifest("v1_adopted", digest)},
		{"adopted retained owner", "bootstrap-authority-manifest.schema.json", retained},
		{"authority signature", "bootstrap-authority-signature.schema.json", map[string]any{
			"contract_version": "bootstrap-authority-signature.v1",
			"manifest_sha256":  digest, "signature": validReceiptSignature(),
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := compileSchema(t, test.schema).Validate(test.value); err != nil {
				t.Fatalf("valid authority contract rejected: %v", err)
			}
		})
	}
}

func TestBootstrapAuthorityRejectsPrivilegeOrOriginDrift(t *testing.T) {
	digest := strings.Repeat("a", 64)
	tests := []struct {
		name   string
		origin string
		mutate func(map[string]any)
	}{
		{"missing environment identity", "fresh", func(value map[string]any) { delete(value, "environment_id") }},
		{"creator self grant", "fresh", func(value map[string]any) { value["createrole_self_grant"] = "set" }},
		{"owner login", "fresh", func(value map[string]any) { authorityRole(value, 0)["login"] = true }},
		{"wrong role count", "fresh", func(value map[string]any) {
			value["roles"] = value["roles"].([]any)[:3]
		}},
		{"cutover role name", "fresh", func(value map[string]any) {
			authorityRole(value, 1)["name"] = "harpy_cutover_decoder"
		}},
		{"wrong membership", "fresh", func(value map[string]any) {
			authorityMembership(value, 0)["grantor"] = "harpy_cluster_admin"
		}},
		{"implicit admin membership", "fresh", func(value map[string]any) { authorityMembership(value, 0)["admin"] = true }},
		{"fresh legacy extension owner", "fresh", func(value map[string]any) {
			authorityDatabase(value)["extension_owner"] = "harpy_legacy_extension_owner"
		}},
		{"adopted missing source owner", "v1_adopted", func(value map[string]any) {
			authorityDatabase(value)["source_owner"] = nil
		}},
		{"database owner drift", "v1_adopted", func(value map[string]any) {
			authorityDatabase(value)["owner"] = "postgres"
		}},
		{"checksum disabled", "fresh", func(value map[string]any) {
			authorityDatabase(value)["data_checksums"] = false
		}},
		{"fresh template missing", "fresh", func(value map[string]any) {
			authorityDatabase(value)["template"] = nil
		}},
		{"adopted template create", "v1_adopted", func(value map[string]any) {
			authorityDatabase(value)["template"] = "template0"
		}},
		{"retained app owner", "v1_adopted", func(value map[string]any) {
			authorityDatabase(value)["extension_owner"] = "harpy_owner"
			authorityDatabase(value)["extension_owner_mode"] = "adopted_retained_inventory"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := validAuthorityManifest(test.origin, digest)
			test.mutate(value)
			if err := compileSchema(t, "bootstrap-authority-manifest.schema.json").Validate(value); err == nil {
				t.Fatal("unsafe bootstrap authority accepted")
			}
		})
	}
}

func validAuthorityManifest(origin, digest string) map[string]any {
	adopted := origin == "v1_adopted"
	database := map[string]any{
		"name": "harpy", "create_database": !adopted, "source_owner": nil,
		"owner": "harpy_owner", "extension_owner": "harpy_extension_admin",
		"extension_owner_mode": "fresh_leased_admin", "extension_owner_application_login": false,
		"ownership_executor": "cluster_bootstrap", "ownership_attestation": "bootstrap_live_catalog",
		"encoding": "UTF8", "locale_provider": "libc", "collation": "en_US.utf8",
		"ctype": "en_US.utf8", "data_checksums": true, "template": "template0", "public_connect": false,
		"public_create": false, "public_temp": false,
	}
	if adopted {
		database["source_owner"] = "oracle"
		database["extension_owner"] = "harpy_legacy_extension_owner"
		database["extension_owner_mode"] = "adopted_reassigned"
		database["ownership_executor"] = "external_root_provisioner"
		database["template"] = nil
	}
	return map[string]any{
		"contract_version": "bootstrap-authority-manifest.v1",
		"manifest_id":      "77777777-7777-4777-8777-777777777777",
		"environment_id":   "88888888-8888-4888-8888-888888888888",
		"release_id":       "44444444-4444-4444-8444-444444444444",
		"issued_at":        "2026-07-13T00:00:00Z", "expires_at": "2026-07-13T00:15:00Z",
		"database_origin": origin, "database": database,
		"connections": map[string]any{
			"endpoint_kind": "direct_postgres", "maintenance_database": "postgres",
			"target_database": "harpy", "same_server_required": true,
		},
		"createrole_self_grant": "", "roles": authorityRoles(),
		"memberships": authorityMemberships(), "credential_leases": authorityLeases(),
		"roles_manifest_sha256": digest, "privileges_manifest_sha256": digest,
		"ownership_manifest_sha256": digest, "bootstrap_operation_digest": digest,
		"credential_cleanup_operation_digest": digest,
	}
}

func authorityRoles() []any {
	return []any{
		authorityRoleFixture("harpy_owner", false, false, -1),
		authorityRoleFixture("harpy_goose", true, false, 4),
		authorityRoleFixture("harpy_api", true, true, 20),
		authorityRoleFixture("harpy_worker", true, true, 20),
	}
}

func authorityRoleFixture(name string, login, inherit bool, limit int) map[string]any {
	return map[string]any{
		"name": name, "enabled": true, "credential_mode": "permanent", "login": login,
		"inherit": inherit, "superuser": false, "createdb": false,
		"createrole": false, "replication": false, "bypassrls": false,
		"connection_limit": limit,
	}
}

func authorityMemberships() []any {
	return []any{
		authorityMembershipFixture("harpy_goose", "harpy_owner", true, false, true),
	}
}

func authorityMembershipFixture(member, role string, enabled, inherit, set bool) map[string]any {
	return map[string]any{
		"member": member, "role": role, "enabled": enabled, "admin": false,
		"inherit": inherit, "set": set, "grantor": "harpy_owner",
	}
}

func authorityLeases() []any {
	return []any{
		authorityLeaseFixture("55555555-5555-4555-8555-555555555555", "harpy_cluster_admin"),
		authorityLeaseFixture("66666666-6666-4666-8666-666666666666", "harpy_extension_admin"),
	}
}

func authorityLeaseFixture(id, principal string) map[string]any {
	return map[string]any{
		"lease_id": id, "principal": principal, "not_before": "2026-07-13T00:00:00Z",
		"expires_at": "2026-07-13T00:15:00Z", "max_duration_seconds": 900,
		"max_connections": 1, "cleanup_policy": "disable_login_clear_password_terminate_sessions",
	}
}

func authorityDatabase(value map[string]any) map[string]any {
	return value["database"].(map[string]any)
}
func authorityRole(value map[string]any, index int) map[string]any {
	return value["roles"].([]any)[index].(map[string]any)
}
func authorityMembership(value map[string]any, index int) map[string]any {
	return value["memberships"].([]any)[index].(map[string]any)
}
