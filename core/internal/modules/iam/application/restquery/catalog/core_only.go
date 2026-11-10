package catalog

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func BuildCore(kind profile.ResourceKind) (Published, error) {
	resource, resourceErr := ResourceFor(kind)
	core, coreErr := coreFields(kind)
	schemaID, schemaErr := coreSchemaID(kind)
	if resourceErr != nil || coreErr != nil || schemaErr != nil {
		return Published{}, ErrContract
	}
	definitions := make([]api.FieldDefinition, 0, len(core))
	classes := make(map[uuid.UUID]profile.Classification, len(core))
	required := make(map[uuid.UUID]bool, len(core))
	maximumBytes := make(map[uuid.UUID]int, len(core))
	for _, spec := range core {
		definition, err := newCoreField(spec)
		if err != nil {
			return Published{}, ErrContract
		}
		identifier, _ := definition.ID()
		if _, duplicate := classes[identifier]; duplicate {
			return Published{}, ErrContract
		}
		definitions = append(definitions, definition)
		classes[identifier] = spec.Classification
		required[identifier] = spec.CreateRequired
		maximumBytes[identifier] = spec.MaximumBytes
	}
	_, customSearch, err := customFields(kind, nil, classes)
	if err != nil {
		return Published{}, ErrContract
	}
	digest, err := finalDigest(resource, schemaID, [sha256.Size]byte{}, core)
	if err != nil {
		return Published{}, ErrContract
	}
	return assemble(
		resource, schemaID, digest, core, definitions, customSearch,
		classes, required, maximumBytes,
	)
}

func coreSchemaID(kind profile.ResourceKind) (string, error) {
	switch kind {
	case profile.ResourceUser:
		return "41000000-0000-4000-8000-000000000001", nil
	case profile.ResourceRole:
		return "42000000-0000-4000-8000-000000000001", nil
	case profile.ResourceOrganization:
		return "43000000-0000-4000-8000-000000000001", nil
	default:
		return "", ErrContract
	}
}
