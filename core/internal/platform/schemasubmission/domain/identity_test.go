package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestJobIdentityRequiresUUIDv7(t *testing.T) {
	valid := uuid.MustParse("018f4f9e-7c21-7a3b-8c4d-123456789abc")
	if id, err := NewJobID(valid); err != nil || id.UUID() != valid {
		t.Fatalf("valid UUIDv7 rejected: %v", err)
	}
	for _, value := range []uuid.UUID{
		uuid.Nil,
		uuid.MustParse("018f4f9e-7c21-4a3b-8c4d-123456789abc"),
		uuid.MustParse("018f4f9e-7c21-7a3b-0c4d-123456789abc"),
	} {
		if _, err := NewJobID(value); err == nil {
			t.Fatalf("invalid identity accepted: %s", value)
		}
	}
}
