package postgres

import (
	"strings"
	"testing"
)

func TestLayoutMigrationContractIsPresentationOnly(t *testing.T) {
	markers := map[int][]string{
		0:  {"'contract', 'sections'", "'field_id', 'order', 'translations', 'widget'"},
		1:  {"system_resource_layout_revisions", "system_resource_layout_publications"},
		2:  {"published system resource layout is immutable", "ENABLE ALWAYS TRIGGER"},
		3:  {"publication target is invalid", "FOR KEY SHARE"},
		5:  {"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'", "pg_advisory_xact_lock"},
		10: {"fields do not match IAM catalog", "ON CONFLICT (resource_kind)"},
		12: {"GRANT EXECUTE ON FUNCTION", "TO harpy_runtime"},
		13: {"field_reference_evidence_v1", "pg_advisory_xact_lock"},
		14: {"iam_profile_active_catalog_read_v1", "fields are outside"},
		15: {"REVOKE EXECUTE", "system_resource_layout_publish_v2"},
		16: {"source_definition_sha256", "presentation_sha256"},
		17: {"publication_json_v2", "catalog_json"},
		18: {"system_resource_layout_publish_v3", "p_presentation_json"},
		19: {"system_resource_layout_resolve_v2", "REVOKE EXECUTE"},
		20: {"translations_valid_v2", "jsonb_object_keys"},
		21: {"sections_valid_v2", "interaction"},
		22: {"definition_valid_v2", "record_filter"},
		23: {"definition_supported_v2", "DROP CONSTRAINT"},
		24: {"field_ids_v2", "SELECT DISTINCT"},
		25: {"system_resource_layout_create_v2", "pg_advisory_xact_lock"},
		26: {"system_resource_layout_update_v2", "v_contract"},
		27: {"system_resource_layout_publish_v4", "compiler.v2"},
		28: {"GRANT EXECUTE", "system_resource_layout_publish_v4"},
		29: {"compiler.v2", "DROP CONSTRAINT"},
		30: {"field_reference_evidence_v2", "field_ids_v2"},
		31: {"GRANT EXECUTE", "field_reference_evidence_v2"},
		32: {"system_resource_layout_create_v3", "uuid_extract_version"},
		33: {"system_resource_layout_publish_v5", "publication replay conflicts"},
		34: {"REVOKE EXECUTE", "system_resource_layout_publish_v5"},
	}
	for index, expected := range markers {
		content := string(readLayoutMigration(
			t, "fresh", layoutMigrationNames[index],
		))
		for _, marker := range expected {
			if !strings.Contains(content, marker) {
				t.Fatalf("migration %d is missing %q", index, marker)
			}
		}
	}
}

func TestLayoutStaticSQLCallsOnlyOwnedRoutines(t *testing.T) {
	for name, statement := range map[string]string{
		"create": createSQL, "get": getSQL, "list": listSQL,
		"update": updateSQL, "delete": deleteSQL,
		"publish": publishSQL, "resolve": resolveSQL,
	} {
		version := "_v1"
		if name == "update" {
			version = "_v2"
		}
		if name == "create" {
			version = "_v3"
		}
		if name == "publish" {
			version = "_v5"
		}
		if name == "resolve" {
			version = "_v2"
		}
		if !strings.Contains(statement, "system_resource_layout_"+name+version) ||
			!strings.Contains(statement,
				"(pg_catalog.current_schemas(false))[2]::name") ||
			strings.Contains(statement, "pg_catalog.current_schema()") ||
			strings.Contains(statement, "iam_") || strings.Contains(statement, " JOIN ") {
			t.Fatalf("%s static SQL crossed its owner boundary: %s", name, statement)
		}
	}
}
