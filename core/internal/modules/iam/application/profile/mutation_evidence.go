package profile

import "github.com/google/uuid"

type MutationEvidence struct {
	OperationID                  uuid.UUID
	RequestSHA256                [32]byte
	ActorID                      uuid.UUID
	ManagePermissionID           uuid.UUID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
}

func (evidence MutationEvidence) Validate() error {
	if evidence.OperationID == uuid.Nil || evidence.OperationID.Variant() != uuid.RFC4122 ||
		evidence.RequestSHA256 == [32]byte{} || evidence.ActorID == uuid.Nil ||
		evidence.ActorID.Variant() != uuid.RFC4122 ||
		evidence.ManagePermissionID == uuid.Nil ||
		evidence.ManagePermissionID.Variant() != uuid.RFC4122 ||
		evidence.ExpectedTokenVersion < 1 ||
		evidence.ExpectedAuthorizationVersion < 1 {
		return ErrInvalidCommand
	}
	return nil
}
