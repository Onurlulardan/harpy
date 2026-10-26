package catalog

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"

func organizationFields() []coreFieldSpec {
	internal := profile.ClassificationInternal
	return []coreFieldSpec{
		{ID: organizationIDFieldID, Name: "id", Column: "id", Kind: "uuid",
			Operators: []string{"eq", "in"}, Sortable: true, CursorBytes: 36,
			Classification: internal, Default: true},
		{ID: organizationNameFieldID, Name: "name", Column: "name", Kind: "text",
			Operators:  []string{"eq", "ne", "in", "not_in"},
			Searchable: true, FoldedColumn: "name_folded",
			Classification: internal, Default: true},
		{ID: organizationSlugFieldID, Name: "slug", Column: "slug", Kind: "text",
			Operators:  []string{"eq", "ne", "in", "not_in"},
			Searchable: true, FoldedColumn: "slug_folded",
			Classification: internal, Default: true},
		{ID: organizationStatusFieldID, Name: "status", Column: "status", Kind: "enum",
			EnumVersion:    "d3000000-0000-4000-8000-000000000001",
			EnumValues:     []string{"active", "suspended", "tombstoned"},
			Operators:      []string{"eq", "ne", "in", "not_in"},
			Classification: internal, Default: true},
		{ID: organizationParentIDFieldID, Name: "parent_id", Column: "parent_id", Kind: "relation",
			Operators:      []string{"eq", "ne", "in", "not_in", "is_null"},
			Classification: internal, Default: true},
		{ID: organizationUserIDsFieldID, Name: "user_ids", Column: "user_ids", Kind: "uuid_list",
			Operators: []string{"contains_any", "contains_all"}, Classification: internal, Default: true},
		{ID: organizationRoleIDsFieldID, Name: "role_ids", Column: "role_ids", Kind: "uuid_list",
			Operators: []string{"contains_any", "contains_all"}, Classification: internal, Default: true},
		{ID: organizationOwnerIDsFieldID, Name: "owner_ids", Column: "owner_ids", Kind: "uuid_list",
			Operators: []string{"contains_any", "contains_all"}, Classification: internal,
			CreateRequired: true, MaximumBytes: 4096},
		{ID: organizationProfileFieldID, Name: "profile_schema_version_id",
			Column: "profile_schema_version_id", Kind: "uuid", Operators: []string{"eq", "in"},
			Classification: internal},
		{ID: organizationRevisionFieldID, Name: "revision", Column: "revision", Kind: "integer",
			Operators:      []string{"eq", "gt", "gte", "lt", "lte", "between"},
			Classification: internal, Default: true},
		{ID: organizationCreatedAtFieldID, Name: "created_at", Column: "created_at",
			Kind: "timestamp", Operators: []string{"eq", "gt", "gte", "lt", "lte", "between"},
			Sortable: true, CursorBytes: 27, Classification: internal, Default: true},
		{ID: organizationUpdatedAtFieldID, Name: "updated_at", Column: "updated_at",
			Kind: "timestamp", Operators: []string{"eq", "gt", "gte", "lt", "lte", "between"},
			Sortable: true, CursorBytes: 27, Classification: internal, Default: true},
	}
}
