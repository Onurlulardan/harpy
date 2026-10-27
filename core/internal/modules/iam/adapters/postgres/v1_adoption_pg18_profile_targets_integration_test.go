//go:build integration

package postgres

func adoptedProfileTargets(tenant iamFoundationTenant) []iamV1AdoptionTarget {
	schemaID, fieldID := adoptedID(tenant, 16), adoptedID(tenant, 17)
	marker := adoptedMarker(tenant)
	return []iamV1AdoptionTarget{
		adoptedTarget("profiles", "schema", "iam_profile_schema_versions", schemaID,
			map[string]any{
				"id": schemaID, "resource_kind": "user", "version": 2,
				"state": "draft", "definition_sha256": adoptedHash("b5"),
				"activated_at": nil, "retired_at": nil, "created_at": adoptedAt,
				"updated_at": adoptedAt, "revision": 1,
			}),
		adoptedTarget("profiles", "field", "iam_profile_fields", fieldID,
			map[string]any{
				"id": fieldID, "resource_kind": "user",
				"field_key": "legacy_alias_" + marker, "value_kind": "string",
				"cardinality": "single", "required": false,
				"classification": "internal", "normalization_profile": "none_v1",
				"state": "active", "created_at": adoptedAt, "updated_at": adoptedAt,
				"retired_at": nil, "revision": 1,
			}),
		adoptedTarget("profiles", "field-revision", "iam_profile_field_revisions",
			schemaID+":"+fieldID, map[string]any{
				"schema_version_id": schemaID, "resource_kind": "user",
				"field_id": fieldID, "validation_schema": map[string]any{"type": "string"},
				"validation_sha256": adoptedHash("b6"), "created_at": adoptedAt,
			}),
	}
}
