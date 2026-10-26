package installationdelivery

import (
	"context"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	"github.com/google/uuid"
)

type UnknownCommand struct {
	OperationID, AttemptID uuid.UUID
	ExpectedRevision       int64
	AdapterID              string
	ClaimTokenSHA256       string
	Cause                  credentialdeliveryv1.Cause
}

func RecordUnknownOutcome(
	ctx context.Context, store Store, command UnknownCommand,
) (Transition, error) {
	if unusable(ctx, store) || command.OperationID.Version() != uuid.Version(7) ||
		command.AttemptID == uuid.Nil || command.ExpectedRevision < 1 ||
		!validAdapter(command.AdapterID) || !validDigest(command.ClaimTokenSHA256) ||
		!validCause(command.Cause) {
		return Transition{}, ErrInvalid
	}
	result, err := store.RecordUnknownOutcome(ctx, command)
	if err != nil {
		return Transition{}, err
	}
	if !result.valid(StateOutcomeUnknown) {
		return Transition{}, ErrUnavailable
	}
	return result, nil
}

func validCause(cause credentialdeliveryv1.Cause) bool {
	return cause == credentialdeliveryv1.CauseClaimLeaseExpired ||
		cause == credentialdeliveryv1.CauseOutcomeUnknown
}
