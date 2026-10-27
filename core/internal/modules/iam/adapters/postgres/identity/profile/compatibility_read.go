package profilepostgres

import (
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func profileSourceRead(
	contract profileWriteContract,
	current profileWriteResource,
) (profileapp.Document, []byte, error) {
	selected := activeProfileFieldIDs(contract.source)
	projection, err := profileapp.ProjectCompatibilityRead(
		contract.source, current.document, selected,
	)
	if err != nil {
		return profileapp.Document{}, nil, err
	}
	return current.document, projection, nil
}

func activeProfileFieldIDs(schema profileapp.Schema) []uuid.UUID {
	fields := schema.Fields()
	result := make([]uuid.UUID, 0, len(fields))
	for _, field := range fields {
		if field.State() == profileapp.FieldActive {
			result = append(result, field.ID())
		}
	}
	return result
}
