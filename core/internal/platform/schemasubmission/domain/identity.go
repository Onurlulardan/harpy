package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type JobID struct{ value uuid.UUID }

func NewJobID(value uuid.UUID) (JobID, error) {
	if !ValidUUIDv7(value) {
		return JobID{}, fmt.Errorf("schema job ID must be an RFC 4122 UUIDv7")
	}
	return JobID{value: value}, nil
}

func (id JobID) UUID() uuid.UUID { return id.value }

func (id JobID) Valid() bool { return ValidUUIDv7(id.value) }

func ValidUUIDv7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}
