package recordkernel

import (
	"slices"

	"github.com/google/uuid"
)

type Record struct {
	values map[uuid.UUID]Value
}

func NewRecord(values map[uuid.UUID]Value) Record {
	result := Record{values: make(map[uuid.UUID]Value, len(values))}
	for fieldID, value := range values {
		result.values[fieldID] = value.clone()
	}
	return result
}

func (record Record) Value(fieldID uuid.UUID) (Value, bool) {
	value, found := record.values[fieldID]
	return value.clone(), found
}

func (record Record) FieldIDs() []uuid.UUID {
	result := make([]uuid.UUID, 0, len(record.values))
	for fieldID := range record.values {
		result = append(result, fieldID)
	}
	slices.SortFunc(result, compareUUID)
	return result
}

func (record Record) clone() Record { return NewRecord(record.values) }

func compareUUID(left, right uuid.UUID) int {
	return slices.Compare(left[:], right[:])
}
