package installationdelivery

import (
	"context"

	"github.com/google/uuid"
)

type StageCommand struct {
	OperationID, AttemptID, CredentialID uuid.UUID
	CredentialVersion                    int64
	InstallationManifestSHA256           string
	DeliveryManifestSHA256               string
}

func Stage(ctx context.Context, store Store, command StageCommand) (Delivery, error) {
	if unusable(ctx, store) || command.OperationID.Version() != uuid.Version(7) ||
		command.AttemptID == uuid.Nil || command.CredentialID == uuid.Nil ||
		command.CredentialVersion != 1 || !validDigest(command.InstallationManifestSHA256) ||
		!validDigest(command.DeliveryManifestSHA256) {
		return Delivery{}, ErrInvalid
	}
	delivery, err := store.Stage(ctx, command)
	if err != nil {
		return Delivery{}, err
	}
	if !delivery.Valid() || delivery.State != StateReady ||
		delivery.OperationID != command.OperationID ||
		delivery.CurrentAttemptID != command.AttemptID ||
		delivery.CurrentCredentialID != command.CredentialID ||
		delivery.CurrentCredentialVersion != command.CredentialVersion {
		return Delivery{}, ErrUnavailable
	}
	return delivery, nil
}
