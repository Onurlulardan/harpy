package identity

type MutationResult struct {
	ID                 IdentityID
	Revision           int64
	AuthorizationClock int64
	Replayed           bool
}

func (result MutationResult) Validate() error {
	if result.ID == (IdentityID{}) || result.Revision < 1 ||
		result.AuthorizationClock < 1 {
		return ErrInvalidRecord
	}
	return nil
}
