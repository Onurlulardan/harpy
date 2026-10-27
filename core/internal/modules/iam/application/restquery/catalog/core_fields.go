package catalog

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func coreFields(resource profile.ResourceKind) ([]coreFieldSpec, error) {
	switch resource {
	case profile.ResourceUser:
		return userFields(), nil
	case profile.ResourceRole:
		return roleFields(), nil
	case profile.ResourceOrganization:
		return organizationFields(), nil
	default:
		return nil, ErrContract
	}
}

func stableDefaultsAndSearch(
	specs []coreFieldSpec,
) (uuid.UUID, []uuid.UUID, []uuid.UUID, error) {
	if len(specs) == 0 {
		return uuid.Nil, nil, nil, ErrContract
	}
	stable, err := uuid.Parse(specs[0].ID)
	if err != nil || specs[0].Name != "id" || !specs[0].Sortable {
		return uuid.Nil, nil, nil, ErrContract
	}
	defaults := make([]uuid.UUID, 0, len(specs))
	search := make([]uuid.UUID, 0, len(specs))
	for _, spec := range specs {
		identifier, parseErr := uuid.Parse(spec.ID)
		if parseErr != nil || identifier.String() != spec.ID {
			return uuid.Nil, nil, nil, ErrContract
		}
		if spec.Default {
			defaults = append(defaults, identifier)
		}
		if spec.Searchable {
			search = append(search, identifier)
		}
	}
	return stable, defaults, search, nil
}
