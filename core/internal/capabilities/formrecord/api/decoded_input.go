package recordkernel

import (
	"github.com/google/uuid"
)

type DecodedInput struct {
	action      Action
	values      map[uuid.UUID]Value
	fieldIDs    []uuid.UUID
	canonical   []byte
	requirement Requirement
}

func (input DecodedInput) Action() Action { return input.action }

func (input DecodedInput) FieldIDs() []uuid.UUID {
	return append([]uuid.UUID(nil), input.fieldIDs...)
}

func (input DecodedInput) Value(fieldID uuid.UUID) (Value, bool) {
	value, found := input.values[fieldID]
	return value.clone(), found
}

func (input DecodedInput) Values() map[uuid.UUID]Value {
	result := make(map[uuid.UUID]Value, len(input.values))
	for fieldID, value := range input.values {
		result[fieldID] = value.clone()
	}
	return result
}

func (input DecodedInput) CanonicalJSON() []byte {
	return append([]byte(nil), input.canonical...)
}

func (input DecodedInput) Requirement() Requirement {
	result, _ := NewRequirement(
		input.requirement.Action(), input.requirement.PublicationID(),
		input.requirement.FormID(),
		input.requirement.FormVersionID(), input.requirement.Fields(),
	)
	return result
}
