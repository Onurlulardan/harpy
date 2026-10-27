package installationdelivery

import (
	"context"

	"github.com/google/uuid"
)

type AcknowledgeCommand struct {
	OperationID, AttemptID uuid.UUID
	ExpectedRevision       int64
	AdapterID              string
	ClaimTokenSHA256       string
}

func Acknowledge(ctx context.Context, store Store, command AcknowledgeCommand) (Delivery, error) {
	if unusable(ctx, store) || command.OperationID.Version() != uuid.Version(7) ||
		command.AttemptID == uuid.Nil || command.ExpectedRevision < 1 ||
		!validAdapter(command.AdapterID) || !validDigest(command.ClaimTokenSHA256) {
		return Delivery{}, ErrInvalid
	}
	delivery, err := store.Acknowledge(ctx, command)
	if err != nil {
		return Delivery{}, err
	}
	if !delivery.Valid() || delivery.State != StateAcknowledged ||
		delivery.OperationID != command.OperationID ||
		delivery.CurrentAttemptID != command.AttemptID ||
		delivery.ClaimAdapterID != command.AdapterID {
		return Delivery{}, ErrUnavailable
	}
	return delivery, nil
}
