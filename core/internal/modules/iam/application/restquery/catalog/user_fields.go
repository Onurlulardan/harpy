package catalog

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"

func userFields() []coreFieldSpec {
	internal := profile.ClassificationInternal
	return []coreFieldSpec{
		{ID: userIDFieldID, Name: "id", Column: "id", Kind: "uuid",
			Operators: []string{"eq", "in"}, Sortable: true, CursorBytes: 36,
			Classification: internal, Default: true},
		{ID: userEmailFieldID, Name: "email", Column: "email", Kind: "text",
			Operators:      []string{"eq", "ne", "in", "not_in", "is_null"},
			Classification: profile.ClassificationConfidential, Default: true},
		{ID: userPortalIdentifierFieldID, Name: "portal_identifier",
			Column: "portal_identifier", Kind: "text",
			Operators:      []string{"eq", "ne", "in", "not_in", "is_null"},
			Classification: profile.ClassificationConfidential},
		{ID: userStatusFieldID, Name: "status", Column: "status", Kind: "enum",
			EnumVersion:    "d1000000-0000-4000-8000-000000000001",
			EnumValues:     []string{"pending", "active", "suspended", "disabled", "tombstoned"},
			Operators:      []string{"eq", "ne", "in", "not_in"},
			Classification: internal, Default: true},
		{ID: userIdentityKindFieldID, Name: "identity_kind", Column: "identity_kind", Kind: "enum",
			EnumVersion:    "d1000000-0000-4000-8000-000000000002",
			EnumValues:     []string{"human"},
			Operators:      []string{"eq", "ne", "in", "not_in"},
			Classification: internal, Default: true},
		{ID: userRoleIDsFieldID, Name: "role_ids", Column: "role_ids", Kind: "uuid_list",
			Operators: []string{"contains_any", "contains_all"}, Classification: internal, Default: true},
		{ID: userOrganizationIDsFieldID, Name: "organization_ids", Column: "organization_ids",
			Kind: "uuid_list", Operators: []string{"contains_any", "contains_all"},
			Classification: internal, Default: true},
		{ID: userProfileSchemaFieldID, Name: "profile_schema_version_id",
			Column: "profile_schema_version_id", Kind: "uuid", Operators: []string{"eq", "in"},
			Classification: internal},
		{ID: userRevisionFieldID, Name: "revision", Column: "revision", Kind: "integer",
			Operators:      []string{"eq", "gt", "gte", "lt", "lte", "between"},
			Classification: internal, Default: true},
		{ID: userCreatedAtFieldID, Name: "created_at", Column: "created_at", Kind: "timestamp",
			Operators: []string{"eq", "gt", "gte", "lt", "lte", "between"},
			Sortable:  true, CursorBytes: 27, Classification: internal, Default: true},
		{ID: userUpdatedAtFieldID, Name: "updated_at", Column: "updated_at", Kind: "timestamp",
			Operators: []string{"eq", "gt", "gte", "lt", "lte", "between"},
			Sortable:  true, CursorBytes: 27, Classification: internal, Default: true},
	}
}
