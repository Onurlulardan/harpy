package platformlicense

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (service *Service) Get(
	ctx context.Context,
	environmentID uuid.UUID,
	operationID uuid.UUID,
) (RefreshOperation, error) {
	if service == nil || service.store == nil || ctx == nil || ctx.Err() != nil ||
		!validEnvironment(environmentID) ||
		operationID.Version() != uuid.Version(7) {
		return RefreshOperation{}, ErrInvalid
	}
	value, err := service.store.Get(
		ctx, environmentID, operationID,
	)
	if err != nil {
		return RefreshOperation{}, fmt.Errorf(
			"read platform license refresh operation: %w", err,
		)
	}
	if !validOperation(value, operationID) {
		return RefreshOperation{}, ErrUnavailable
	}
	return value, nil
}
