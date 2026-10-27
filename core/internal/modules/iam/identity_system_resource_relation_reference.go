package iam

import (
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

const maximumIdentityRelationReferences = 1000

type identityRelationDefinition struct {
	field    string
	fieldID  string
	target   publicresource.Resource
	multiple bool
}

type identityRelationReference struct {
	definition identityRelationDefinition
	ids        []uuid.UUID
}

func identityRelationDefinitions(
	resource publicresource.Resource,
) []identityRelationDefinition {
	switch resource {
	case publicresource.Users:
		return []identityRelationDefinition{
			{"role_ids", publicresource.UserRoleIDsFieldID, publicresource.Roles, true},
			{"organization_ids", publicresource.UserOrganizationIDsFieldID,
				publicresource.Organizations, true},
		}
	case publicresource.Roles:
		return []identityRelationDefinition{
			{"organization_id", publicresource.RoleOrganizationIDFieldID,
				publicresource.Organizations, false},
			{"user_ids", publicresource.RoleUserIDsFieldID, publicresource.Users, true},
		}
	case publicresource.Organizations:
		return []identityRelationDefinition{
			{"parent_id", publicresource.OrganizationParentIDFieldID,
				publicresource.Organizations, false},
			{"user_ids", publicresource.OrganizationUserIDsFieldID,
				publicresource.Users, true},
			{"role_ids", publicresource.OrganizationRoleIDsFieldID,
				publicresource.Roles, true},
			{"owner_ids", publicresource.OrganizationOwnerIDsFieldID,
				publicresource.Users, true},
		}
	default:
		return nil
	}
}

func identityRelationReferences(
	resource publicresource.Resource,
	item json.RawMessage,
) ([]identityRelationReference, error) {
	var document map[string]json.RawMessage
	definitions := identityRelationDefinitions(resource)
	if len(definitions) == 0 || json.Unmarshal(item, &document) != nil ||
		document == nil {
		return nil, publicresource.ErrContract
	}
	result := make([]identityRelationReference, 0, len(definitions))
	for _, definition := range definitions {
		raw, present := document[definition.field]
		if !present {
			continue
		}
		ids, err := identityRelationIDs(raw, definition.multiple)
		if err != nil {
			return nil, publicresource.ErrUnavailable
		}
		result = append(result, identityRelationReference{
			definition: definition, ids: ids,
		})
	}
	return result, nil
}

func identityRelationIDs(raw json.RawMessage, multiple bool) ([]uuid.UUID, error) {
	if !multiple {
		if string(raw) == "null" {
			return []uuid.UUID{}, nil
		}
		var value string
		if json.Unmarshal(raw, &value) != nil {
			return nil, publicresource.ErrContract
		}
		id, err := uuid.Parse(value)
		if err != nil || id.String() != value {
			return nil, publicresource.ErrContract
		}
		return []uuid.UUID{id}, nil
	}
	var values []string
	if json.Unmarshal(raw, &values) != nil || values == nil ||
		len(values) > maximumIdentityRelationReferences {
		return nil, publicresource.ErrContract
	}
	result := make([]uuid.UUID, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for index, value := range values {
		id, err := uuid.Parse(value)
		if err != nil || id.String() != value {
			return nil, publicresource.ErrContract
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, publicresource.ErrContract
		}
		seen[id], result[index] = struct{}{}, id
	}
	return result, nil
}
