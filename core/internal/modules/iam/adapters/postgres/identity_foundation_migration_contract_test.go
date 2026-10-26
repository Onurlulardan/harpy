package postgres

import (
	"strings"
	"testing"
)

func TestIAMIdentityMigrationContracts(t *testing.T) {
	tests := []struct {
		name        string
		foreignKeys int
		markers     []string
	}{
		{name: iamIdentityMigrationNames[0], foreignKeys: 1, markers: []string{
			"CREATE TABLE iam_users (",
			"account_type IN ('operator', 'portal')",
			"status IN ('pending', 'active', 'suspended', 'disabled', 'tombstoned')",
			"FOREIGN KEY (profile_resource_kind, profile_schema_version_id)",
			"REFERENCES iam_profile_schema_versions (resource_kind, id)",
			"CHECK (profile_resource_kind = 'user')",
			"pg_catalog.jsonb_typeof(profile_values) = 'object'",
			"pg_catalog.octet_length(profile_sha256) = 32",
			"CHECK (token_version > 0)", "CHECK (authorization_version >= 0)",
			"CHECK (revision > 0)", "status = 'tombstoned' AND tombstoned_at IS NOT NULL",
		}},
		{name: iamIdentityMigrationNames[1], foreignKeys: 1, markers: []string{
			"CREATE TABLE iam_user_identifiers (",
			"kind IN ('email_v1', 'portal_v1')",
			"state IN ('pending', 'verified', 'active', 'tombstoned', 'retired')",
			"pg_catalog.octet_length(normalized_sha256) = 32",
			"normalized_value text COLLATE \"C\" NOT NULL",
			"verified_at IS NOT NULL", "activated_at IS NOT NULL",
			"tombstoned_at IS NOT NULL", "cooldown_until IS NOT NULL",
			"cooldown_until = tombstoned_at + INTERVAL '30 days'",
			"retired_at IS NOT NULL", "retired_at >= cooldown_until",
			"CREATE UNIQUE INDEX iam_user_identifiers_reserved_value_uidx",
			"ON iam_user_identifiers (kind, normalized_value)",
			"WHERE state <> 'retired'",
			"CREATE UNIQUE INDEX iam_user_identifiers_one_active_uidx",
			"ON iam_user_identifiers (user_id, kind)", "WHERE state = 'active'",
		}},
		{name: iamIdentityMigrationNames[2], foreignKeys: 2, markers: []string{
			"CREATE TABLE iam_organizations (",
			"slug text COLLATE \"C\" NOT NULL",
			"CONSTRAINT iam_organizations_slug_key UNIQUE (slug)",
			"FOREIGN KEY (parent_id)", "REFERENCES iam_organizations (id)",
			"CHECK (parent_id IS DISTINCT FROM id)",
			"FOREIGN KEY (profile_resource_kind, profile_schema_version_id)",
			"CHECK (profile_resource_kind = 'organization')",
			"pg_catalog.jsonb_typeof(profile_values) = 'object'",
			"pg_catalog.octet_length(profile_sha256) = 32",
			"status IN ('active', 'suspended', 'tombstoned')",
		}},
		{name: iamIdentityMigrationNames[3], foreignKeys: 1, markers: []string{
			"CREATE TABLE iam_password_credentials (",
			"CONSTRAINT iam_password_credentials_pkey PRIMARY KEY (user_id)",
			"algorithm IN ('argon2id_v1', 'bcrypt_v1_adopted')",
			"algorithm = 'argon2id_v1'", "pg_catalog.octet_length(salt) = 16",
			"salt IS NOT NULL AND tag IS NOT NULL AND memory_kib IS NOT NULL",
			"iterations IS NOT NULL AND parallelism IS NOT NULL",
			"pg_catalog.octet_length(tag) = 32", "memory_kib = 65536",
			"iterations = 3", "parallelism = 4",
			"legacy_bcrypt_hash IS NULL AND NOT upgrade_required",
			"algorithm = 'bcrypt_v1_adopted'", "legacy_bcrypt_hash IS NOT NULL",
			"pg_catalog.char_length(legacy_bcrypt_hash) = 60",
			"legacy_bcrypt_hash ~ '^\\$2[aby]\\$",
			"state IN ('active', 'reset_required', 'disabled', 'compromised')",
		}},
		{name: iamIdentityMigrationNames[4], foreignKeys: 2, markers: []string{
			"CREATE TABLE iam_organization_memberships (",
			"UNIQUE (organization_id, user_id)",
			"UNIQUE (id, organization_id, user_id)",
			"FOREIGN KEY (organization_id) REFERENCES iam_organizations (id)",
			"FOREIGN KEY (user_id)", "REFERENCES iam_users (id)",
			"membership_kind IN ('member', 'owner')",
			"state IN ('active', 'suspended', 'tombstoned')",
		}},
		{name: iamIdentityMigrationNames[5], foreignKeys: 2, markers: []string{
			"CREATE TABLE iam_roles (",
			"UNIQUE (id, scope_kind, scope_id)",
			"FOREIGN KEY (organization_id)", "REFERENCES iam_organizations (id)",
			"FOREIGN KEY (profile_resource_kind, profile_schema_version_id)",
			"scope_kind = 'tenant' AND organization_id IS NULL",
			"scope_kind = 'organization' AND organization_id IS NOT NULL",
			"organization_id = scope_id",
			"CHECK (profile_resource_kind = 'role')",
			"name_key bytea NOT NULL", "name_key = pg_catalog.convert_to(name, 'UTF8')",
			"is_protected boolean NOT NULL", "is_default boolean NOT NULL",
			"CREATE UNIQUE INDEX iam_roles_tenant_name_uidx",
			"ON iam_roles (name_key) WHERE scope_kind = 'tenant'",
			"CREATE UNIQUE INDEX iam_roles_organization_name_uidx",
			"ON iam_roles (organization_id, name_key)",
			"CREATE UNIQUE INDEX iam_roles_one_tenant_default_uidx",
			"CREATE UNIQUE INDEX iam_roles_one_organization_default_uidx",
			"status = 'active' AND is_default",
		}},
		{name: iamIdentityMigrationNames[6], foreignKeys: 4, markers: []string{
			"CREATE TABLE iam_user_role_assignments (",
			"UNIQUE (user_id, role_id)",
			"FOREIGN KEY (role_id, role_scope_kind, role_scope_id)",
			"REFERENCES iam_roles (id, scope_kind, scope_id)",
			"FOREIGN KEY (membership_id, organization_id, user_id)",
			"REFERENCES iam_organization_memberships (id, organization_id, user_id)",
			"role_scope_kind = 'tenant' AND organization_id IS NULL AND membership_id IS NULL",
			"role_scope_kind = 'organization' AND organization_id IS NOT NULL",
			"organization_id = role_scope_id",
			"membership_id IS NOT NULL", "state IN ('active', 'suspended', 'revoked')",
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := string(readIAMIdentityMigration(t, "fresh", test.name))
			assertIAMIdentityMarkers(t, source, test.foreignKeys, test.markers)
		})
	}
}

func assertIAMIdentityMarkers(t *testing.T, source string, foreignKeys int, markers []string) {
	t.Helper()
	if strings.Count(source, "CREATE TABLE ") != 1 ||
		strings.Count(source, "FOREIGN KEY (") != foreignKeys ||
		strings.Count(source, "ON UPDATE NO ACTION ON DELETE RESTRICT") != foreignKeys {
		t.Fatal("IAM identity migration relation shape differs from the final contract")
	}
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM identity migration lacks exact marker %q", marker)
		}
	}
}
