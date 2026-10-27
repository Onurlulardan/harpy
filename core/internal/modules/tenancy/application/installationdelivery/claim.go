package installationdelivery

import (
	"context"

	"github.com/google/uuid"
)

type ClaimCommand struct {
	OperationID, AttemptID uuid.UUID
	ExpectedRevision       int64
	AdapterID              string
	ClaimTokenSHA256       string
	LeaseSeconds           int
}

func Claim(ctx context.Context, store Store, command ClaimCommand) (Delivery, error) {
	if unusable(ctx, store) || command.OperationID.Version() != uuid.Version(7) ||
		command.AttemptID == uuid.Nil || command.ExpectedRevision < 1 ||
		!validAdapter(command.AdapterID) || !validDigest(command.ClaimTokenSHA256) ||
		command.LeaseSeconds < 10 || command.LeaseSeconds > 300 {
		return Delivery{}, ErrInvalid
	}
	delivery, err := store.Claim(ctx, command)
	if err != nil {
		return Delivery{}, err
	}
	if !delivery.Valid() || delivery.State != StateClaimed ||
		delivery.OperationID != command.OperationID ||
		delivery.CurrentAttemptID != command.AttemptID ||
		delivery.ClaimAdapterID != command.AdapterID {
		return Delivery{}, ErrUnavailable
	}
	return delivery, nil
}
