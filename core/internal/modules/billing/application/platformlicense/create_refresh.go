package platformlicense

import (
	"context"
	"fmt"
)

func (service *Service) Create(
	ctx context.Context,
	command RefreshCommand,
) (RefreshAccepted, error) {
	if service == nil || service.store == nil || ctx == nil || ctx.Err() != nil ||
		!validEnvironment(command.EnvironmentID) || command.TenantID == [16]byte{} ||
		command.ActorID == [16]byte{} ||
		!idempotencyPattern.MatchString(command.IdempotencyKey) ||
		!validDigest(command.RequestSHA256) {
		return RefreshAccepted{}, ErrInvalid
	}
	value, err := service.store.Create(ctx, command)
	if err != nil {
		return RefreshAccepted{}, fmt.Errorf("request platform license refresh: %w", err)
	}
	if value.OperationID.Version() != 7 || value.CreatedAt.IsZero() {
		return RefreshAccepted{}, ErrUnavailable
	}
	return value, nil
}
