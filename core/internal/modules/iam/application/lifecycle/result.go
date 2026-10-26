package lifecycle

import "github.com/google/uuid"

type MutationResult struct {
	ID                   uuid.UUID
	SubjectUserID        uuid.UUID
	Revision             int64
	TokenVersion         int64
	AuthorizationVersion int64
	Replayed             bool
}

func (result MutationResult) Validate() error {
	if result.ID == uuid.Nil || result.SubjectUserID == uuid.Nil ||
		result.Revision < 1 || result.TokenVersion < 1 ||
		result.AuthorizationVersion < 1 {
		return ErrUnavailable
	}
	return nil
}
