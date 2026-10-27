package rbac

type MutationEvidence struct {
	OperationID                  ID
	RequestSHA256                [32]byte
	ActorID                      ID
	RequiredPermissionID         ID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
}

func (evidence MutationEvidence) Validate() error {
	if !evidence.OperationID.Valid() || evidence.RequestSHA256 == [32]byte{} ||
		!evidence.ActorID.Valid() || !evidence.RequiredPermissionID.Valid() ||
		evidence.ExpectedTokenVersion < 1 ||
		evidence.ExpectedAuthorizationVersion < 1 {
		return ErrInvalidMutation
	}
	return nil
}

type QueryEvidence struct {
	ActorID              ID
	RequiredPermissionID ID
}

func (evidence QueryEvidence) Validate() error {
	if !evidence.ActorID.Valid() || !evidence.RequiredPermissionID.Valid() {
		return ErrInvalidQuery
	}
	return nil
}
