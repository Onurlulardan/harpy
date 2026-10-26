package identity

type QueryEvidence struct {
	ActorID                      IdentityID
	PermissionID                 IdentityID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
}

func (evidence QueryEvidence) Validate() error {
	if evidence.ActorID == (IdentityID{}) || evidence.PermissionID == (IdentityID{}) ||
		evidence.ExpectedTokenVersion < 1 ||
		evidence.ExpectedAuthorizationVersion < 1 {
		return ErrInvalidQuery
	}
	return nil
}
