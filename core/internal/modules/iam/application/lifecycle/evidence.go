package lifecycle

import "github.com/google/uuid"

type MutationEvidence struct {
	OperationID                  uuid.UUID
	RequestSHA256                [32]byte
	ActorID                      uuid.UUID
	PermissionID                 uuid.UUID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
}

func (evidence MutationEvidence) Validate() error {
	if evidence.OperationID == uuid.Nil || evidence.RequestSHA256 == [32]byte{} ||
		evidence.ActorID == uuid.Nil || evidence.PermissionID == uuid.Nil ||
		evidence.ExpectedTokenVersion < 1 || evidence.ExpectedAuthorizationVersion < 1 {
		return ErrInvalid
	}
	return nil
}

type CredentialEvidence struct {
	OperationID   uuid.UUID
	RequestSHA256 [32]byte
}

func (evidence CredentialEvidence) Validate() error {
	if evidence.OperationID == uuid.Nil || evidence.RequestSHA256 == [32]byte{} {
		return ErrInvalid
	}
	return nil
}
