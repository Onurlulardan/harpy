package permissionrequest

import (
	"time"

	"github.com/google/uuid"
)

func validCreateCommand() CreateCommand {
	requester := uuid.New()
	return CreateCommand{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1},
		GroupID: uuid.New(), RequesterID: requester, Mode: ModeSelfService,
		ExpectedAuthorizationVersion: 1,
		StepUpPurposeSHA256:          [32]byte{1},
		ExpiresAt:                    time.Now().Add(time.Hour),
		Requests: []CreateItem{{
			ID: uuid.New(), Ordinal: 1,
			Target:       Target{Kind: TargetUser, ID: requester},
			PermissionID: uuid.New(), PlannedBindingID: uuid.New(),
		}},
	}
}

func validStepUpEvidence() StepUpEvidence {
	return StepUpEvidence{
		PurposeSHA256: [32]byte{2}, AuthTime: time.Now(), ACR: StepUpACR,
		AMRWebAuthn: true,
	}
}
