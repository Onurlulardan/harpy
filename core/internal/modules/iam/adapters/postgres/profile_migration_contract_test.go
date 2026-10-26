package postgres

import (
	"strings"
	"testing"
)

func TestIAMProfileSchemaVersionMigrationContract(t *testing.T) {
	source := string(readProfileMigration(t, "fresh", profileMigrationNames[0]))
	assertProfileMigrationShape(t, source, 10, 1, []string{
		"CREATE TABLE iam_profile_schema_versions (",
		"id uuid NOT NULL", "resource_kind text NOT NULL",
		"version bigint NOT NULL", "state text NOT NULL",
		"definition_sha256 bytea NOT NULL", "activated_at timestamptz",
		"retired_at timestamptz", "created_at timestamptz NOT NULL",
		"updated_at timestamptz NOT NULL", "revision bigint NOT NULL",
		"CONSTRAINT iam_profile_schema_versions_pkey PRIMARY KEY (id)",
		"CONSTRAINT iam_profile_schema_versions_resource_version_key",
		"UNIQUE (resource_kind, version)",
		"CONSTRAINT iam_profile_schema_versions_resource_id_key",
		"UNIQUE (resource_kind, id)",
		"resource_kind IN ('user', 'organization', 'role')",
		"CHECK (version > 0)", "state IN ('draft', 'active', 'retired')",
		"pg_catalog.octet_length(definition_sha256) = 32",
		"CHECK (revision > 0)",
		"updated_at >= created_at", "activated_at >= created_at",
		"retired_at >= updated_at",
		"state = 'draft' AND activated_at IS NULL AND retired_at IS NULL",
		"state = 'active' AND activated_at IS NOT NULL AND retired_at IS NULL",
		"state = 'retired'", "retired_at >= activated_at",
		"CREATE UNIQUE INDEX iam_profile_schema_versions_one_active_resource_idx",
		"ON iam_profile_schema_versions (resource_kind)", "WHERE state = 'active'",
	})
}

func TestIAMProfileFieldMigrationContract(t *testing.T) {
	source := string(readProfileMigration(t, "fresh", profileMigrationNames[1]))
	assertProfileMigrationShape(t, source, 14, 0, []string{
		"CREATE TABLE iam_profile_fields (", "id uuid NOT NULL",
		"resource_kind text NOT NULL", "field_key text COLLATE \"C\" NOT NULL",
		"value_kind text NOT NULL", "cardinality text NOT NULL",
		"required boolean NOT NULL", "classification text NOT NULL",
		"normalization_profile text NOT NULL",
		"state text NOT NULL", "created_at timestamptz NOT NULL",
		"updated_at timestamptz NOT NULL", "retired_at timestamptz",
		"revision bigint NOT NULL",
		"CONSTRAINT iam_profile_fields_pkey PRIMARY KEY (id)",
		"CONSTRAINT iam_profile_fields_resource_field_key",
		"UNIQUE (resource_kind, field_key)",
		"CONSTRAINT iam_profile_fields_resource_id_key UNIQUE (resource_kind, id)",
		"resource_kind IN ('user', 'organization', 'role')",
		"pg_catalog.char_length(field_key) <= 63",
		"field_key ~ '^[a-z][a-z0-9]*(_[a-z0-9]+)*$'",
		"'string', 'integer', 'decimal', 'boolean', 'date'",
		"'timestamp', 'uuid', 'enum', 'json'",
		"cardinality IN ('single', 'list')",
		"classification IN ('public', 'internal', 'confidential', 'restricted')",
		"'none_v1', 'email_v1', 'portal_v1'",
		"normalization_profile = 'none_v1' OR value_kind = 'string'",
		"state IN ('active', 'retired')", "CHECK (revision > 0)",
		"updated_at >= created_at", "retired_at >= updated_at",
		"state = 'active' AND retired_at IS NULL",
		"state = 'retired' AND retired_at IS NOT NULL",
	})
}

func TestIAMProfileFieldRevisionMigrationContract(t *testing.T) {
	source := string(readProfileMigration(t, "fresh", profileMigrationNames[2]))
	assertProfileMigrationShape(t, source, 6, 0, []string{
		"CREATE TABLE iam_profile_field_revisions (",
		"schema_version_id uuid NOT NULL", "resource_kind text NOT NULL",
		"field_id uuid NOT NULL", "validation_schema jsonb NOT NULL",
		"validation_sha256 bytea NOT NULL", "created_at timestamptz NOT NULL",
		"PRIMARY KEY (schema_version_id, field_id)",
		"FOREIGN KEY (resource_kind, schema_version_id)",
		"REFERENCES iam_profile_schema_versions (resource_kind, id)",
		"FOREIGN KEY (resource_kind, field_id)",
		"REFERENCES iam_profile_fields (resource_kind, id)",
		"ON UPDATE NO ACTION ON DELETE RESTRICT",
		"resource_kind IN ('user', 'organization', 'role')",
		"pg_catalog.jsonb_typeof(validation_schema) = 'object'",
		"pg_catalog.octet_length(validation_sha256) = 32",
	})
	for _, immutable := range []string{
		"value_kind text", "cardinality text", "required boolean",
		"classification text", "normalization_profile text",
	} {
		if strings.Contains(source, immutable) {
			t.Fatalf("profile revision repeats immutable field semantic %q", immutable)
		}
	}
}

func assertProfileMigrationShape(
	t *testing.T,
	source string,
	constraintCount int,
	indexCount int,
	markers []string,
) {
	t.Helper()
	if strings.Count(source, "CREATE TABLE ") != 1 ||
		strings.Count(source, "CONSTRAINT iam_") != constraintCount ||
		strings.Count(source, "CREATE UNIQUE INDEX ") != indexCount {
		t.Fatal("profile migration object shape differs from the final contract")
	}
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("profile migration lacks exact marker %q", marker)
		}
	}
}
