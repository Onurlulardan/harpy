package profile

import "github.com/google/uuid"

type MutationResult struct {
	ID                 uuid.UUID
	Revision           int64
	AuthorizationClock int64
	Replayed           bool
}

func (result MutationResult) Validate() error {
	if result.ID == uuid.Nil || result.ID.Variant() != uuid.RFC4122 ||
		result.Revision < 1 || result.AuthorizationClock < 1 {
		return ErrEvidenceMissing
	}
	return nil
}
