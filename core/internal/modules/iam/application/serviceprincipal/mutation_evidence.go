package serviceprincipal

import (
	"time"

	"github.com/google/uuid"
)

type MutationEvidence struct {
	OperationID                  uuid.UUID
	RequestSHA256                [32]byte
	ActorID                      uuid.UUID
	ManagePermissionID           uuid.UUID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
	PurposeSHA256                [32]byte
	AuthTime                     time.Time
	AMRPassword                  bool
	AMRWebAuthn                  bool
	AMRTOTP                      bool
}

func (evidence MutationEvidence) Validate() error {
	if evidence.OperationID == uuid.Nil || evidence.RequestSHA256 == [32]byte{} ||
		evidence.ActorID == uuid.Nil || evidence.ManagePermissionID == uuid.Nil ||
		evidence.ExpectedTokenVersion < 1 ||
		evidence.ExpectedAuthorizationVersion < 1 ||
		evidence.PurposeSHA256 == [32]byte{} || evidence.AuthTime.IsZero() ||
		!(evidence.AMRPassword || evidence.AMRWebAuthn || evidence.AMRTOTP) {
		return ErrInvalidGrant
	}
	return nil
}
