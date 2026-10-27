package serviceprincipal

import "github.com/google/uuid"

type MutationResult struct {
	ID                 uuid.UUID
	Revision           int64
	AuthorizationClock int64
	Replayed           bool
}

func (result MutationResult) Validate() error {
	if result.ID == uuid.Nil || result.Revision < 1 || result.AuthorizationClock < 1 {
		return ErrInvalidGrant
	}
	return nil
}
