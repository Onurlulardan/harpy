package identity

type MutationEvidence struct {
	OperationID                  IdentityID
	RequestSHA256                [32]byte
	ActorID                      IdentityID
	PermissionID                 IdentityID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
}

func (evidence MutationEvidence) Validate() error {
	if evidence.OperationID == (IdentityID{}) || evidence.RequestSHA256 == [32]byte{} ||
		evidence.ActorID == (IdentityID{}) || evidence.PermissionID == (IdentityID{}) ||
		evidence.ExpectedTokenVersion < 1 ||
		evidence.ExpectedAuthorizationVersion < 1 {
		return ErrInvalidCommand
	}
	return nil
}
